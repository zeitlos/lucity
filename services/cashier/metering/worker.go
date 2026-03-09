package metering

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/labels"
	stripelib "github.com/zeitlos/lucity/services/cashier/stripe"
)

// Worker periodically queries resource usage and reports it to Stripe.
type Worker struct {
	stripe   *stripelib.Client
	deployer deployer.DeployerServiceClient
	signoz   *SigNozClient
	interval time.Duration
	cancel   context.CancelFunc
	done     chan struct{}
}

// NewWorker creates a metering worker.
func NewWorker(stripeClient *stripelib.Client, deployerClient deployer.DeployerServiceClient, signozClient *SigNozClient, interval time.Duration) *Worker {
	return &Worker{
		stripe:   stripeClient,
		deployer: deployerClient,
		signoz:   signozClient,
		interval: interval,
		done:     make(chan struct{}),
	}
}

func (w *Worker) Label() string { return "Metering Worker" }

func (w *Worker) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	defer close(w.done)

	// Run once immediately on startup, then on ticker.
	w.tick(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) Shutdown(ctx context.Context) error {
	if w.cancel != nil {
		w.cancel()
	}
	select {
	case <-w.done:
	case <-ctx.Done():
	}
	return nil
}

// deployerCtx creates a system-level auth context for calling the deployer.
func deployerCtx(ctx context.Context) context.Context {
	ctx = auth.WithClaims(ctx, &auth.Claims{
		Subject: "cashier",
		Roles:   []auth.Role{auth.RoleUser},
	})
	return auth.OutgoingContext(ctx)
}

// workspaceData holds billing metadata for a workspace.
type workspaceData struct {
	customerID     string
	subscriptionID string
	ecoNamespaces  []string
	ecoAllocations []allocEntry
	prodAllocations []allocEntry
}

type allocEntry struct {
	namespace string
	cpuMillis int32
	memoryMB  int32
	diskMB    int32
}

func (w *Worker) tick(ctx context.Context) {
	slog.Info("metering tick started")
	start := time.Now()

	callCtx := deployerCtx(ctx)

	// 1. List all workspaces and their billing metadata.
	wsList, err := w.deployer.ListWorkspaces(callCtx, &deployer.ListWorkspacesRequest{})
	if err != nil {
		slog.Error("metering: failed to list workspaces", "error", err)
		return
	}

	// 2. Get stripe subscription IDs per workspace.
	workspaces := make(map[string]*workspaceData)
	for _, ws := range wsList.Workspaces {
		meta, err := w.deployer.WorkspaceMetadata(callCtx, &deployer.WorkspaceMetadataRequest{
			Workspace: ws.Id,
		})
		if err != nil {
			slog.Warn("metering: failed to get workspace metadata", "workspace", ws.Id, "error", err)
			continue
		}
		if meta.StripeSubscriptionId == "" || meta.StripeCustomerId == "" {
			continue // No billing
		}
		workspaces[ws.Id] = &workspaceData{
			customerID:     meta.StripeCustomerId,
			subscriptionID: meta.StripeSubscriptionId,
		}
	}

	if len(workspaces) == 0 {
		slog.Info("metering: no billable workspaces")
		return
	}

	// 3. List all resource allocations.
	allocResp, err := w.deployer.ListResourceAllocations(callCtx, &deployer.ListResourceAllocationsRequest{})
	if err != nil {
		slog.Error("metering: failed to list resource allocations", "error", err)
		return
	}

	// 4. Group allocations by workspace.
	var allEcoNamespaces []string
	for _, alloc := range allocResp.Allocations {
		ws, ok := workspaces[alloc.Workspace]
		if !ok {
			continue
		}

		ns := labels.NamespaceFor(alloc.Workspace, alloc.Project, alloc.Environment)
		entry := allocEntry{
			namespace: ns,
			cpuMillis: alloc.CpuMillicores,
			memoryMB:  alloc.MemoryMb,
			diskMB:    alloc.DiskMb,
		}

		if alloc.Tier == deployer.ResourceTier_RESOURCE_TIER_ECO {
			ws.ecoNamespaces = append(ws.ecoNamespaces, ns)
			ws.ecoAllocations = append(ws.ecoAllocations, entry)
			allEcoNamespaces = append(allEcoNamespaces, ns)
		} else {
			ws.prodAllocations = append(ws.prodAllocations, entry)
		}
	}

	// 5. Query SigNoz for eco namespace usage.
	var cpuByNs, memByNs, diskByNs map[string]float64
	if len(allEcoNamespaces) > 0 {
		cpuByNs, err = w.signoz.CPUByNamespace(ctx, allEcoNamespaces, w.interval)
		if err != nil {
			slog.Error("metering: failed to query CPU usage", "error", err)
			cpuByNs = make(map[string]float64)
		}

		memByNs, err = w.signoz.MemoryByNamespace(ctx, allEcoNamespaces, w.interval)
		if err != nil {
			slog.Error("metering: failed to query memory usage", "error", err)
			memByNs = make(map[string]float64)
		}

		diskByNs, err = w.signoz.DiskByNamespace(ctx, allEcoNamespaces, w.interval)
		if err != nil {
			slog.Error("metering: failed to query disk usage", "error", err)
			diskByNs = make(map[string]float64)
		}
	}

	// 6. Report to Stripe per workspace.
	now := time.Now().Unix()
	intervalMinutes := w.interval.Minutes()

	for wsID, ws := range workspaces {
		if err := w.reportWorkspace(ctx, wsID, ws, cpuByNs, memByNs, diskByNs, now, intervalMinutes); err != nil {
			slog.Error("metering: failed to report workspace", "workspace", wsID, "error", err)
		}
	}

	slog.Info("metering tick completed", "duration", time.Since(start), "workspaces", len(workspaces))
}

func (w *Worker) reportWorkspace(ctx context.Context, wsID string, ws *workspaceData, cpuByNs, memByNs, diskByNs map[string]float64, timestamp int64, intervalMinutes float64) error {
	// Report eco usage from SigNoz via Billing Meter events.
	if len(ws.ecoNamespaces) > 0 {
		var totalCPUSeconds, totalMemBytes, totalDiskBytes float64
		for _, ns := range ws.ecoNamespaces {
			totalCPUSeconds += cpuByNs[ns]
			totalMemBytes += memByNs[ns]
			totalDiskBytes += diskByNs[ns]
		}

		// CPU: seconds → vCPU-minutes (Stripe unit).
		cpuMinutes := int64(math.Ceil(totalCPUSeconds / 60))
		if cpuMinutes > 0 && w.stripe.Meters.EcoCPUEventName != "" {
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.EcoCPUEventName, ws.customerID, cpuMinutes, timestamp); err != nil {
				slog.Error("metering: eco CPU report failed", "workspace", wsID, "error", err)
			}
		}

		// Memory: avg bytes → GB-minutes.
		gbMinutes := int64(math.Ceil(totalMemBytes / (1024 * 1024 * 1024) * intervalMinutes))
		if gbMinutes > 0 && w.stripe.Meters.EcoMemEventName != "" {
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.EcoMemEventName, ws.customerID, gbMinutes, timestamp); err != nil {
				slog.Error("metering: eco memory report failed", "workspace", wsID, "error", err)
			}
		}

		// Disk: bytes → GB.
		diskGB := int64(math.Ceil(totalDiskBytes / (1024 * 1024 * 1024)))
		if diskGB > 0 && w.stripe.Meters.EcoDiskEventName != "" {
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.EcoDiskEventName, ws.customerID, diskGB, timestamp); err != nil {
				slog.Error("metering: eco disk report failed", "workspace", wsID, "error", err)
			}
		}

		slog.Info("metering: eco usage reported",
			"workspace", wsID,
			"cpu_minutes", cpuMinutes,
			"memory_gb_minutes", gbMinutes,
			"disk_gb", diskGB,
		)
	}

	// Report production allocations as licensed quantities (requires subscription lookup).
	if len(ws.prodAllocations) > 0 {
		sub, err := w.stripe.Subscription(ctx, ws.subscriptionID)
		if err != nil {
			return err
		}

		var totalCPUMillis, totalMemMB, totalDiskMB int32
		for _, alloc := range ws.prodAllocations {
			totalCPUMillis += alloc.cpuMillis
			totalMemMB += alloc.memoryMB
			totalDiskMB += alloc.diskMB
		}

		// CPU: millicores → vCPU (rounded up).
		cpuUnits := int64(math.Ceil(float64(totalCPUMillis) / 1000))
		itemID := stripelib.FindItemByPrice(sub, w.stripe.Prices.ProdCPUPriceID)
		if itemID != "" {
			if err := w.stripe.UpdateItemQuantity(ctx, itemID, cpuUnits); err != nil {
				slog.Error("metering: prod CPU update failed", "workspace", wsID, "error", err)
			}
		}

		// Memory: MB → GB (rounded up).
		memUnits := int64(math.Ceil(float64(totalMemMB) / 1024))
		itemID = stripelib.FindItemByPrice(sub, w.stripe.Prices.ProdMemPriceID)
		if itemID != "" {
			if err := w.stripe.UpdateItemQuantity(ctx, itemID, memUnits); err != nil {
				slog.Error("metering: prod memory update failed", "workspace", wsID, "error", err)
			}
		}

		// Disk: MB → GB (rounded up).
		diskUnits := int64(math.Ceil(float64(totalDiskMB) / 1024))
		itemID = stripelib.FindItemByPrice(sub, w.stripe.Prices.ProdDiskPriceID)
		if itemID != "" {
			if err := w.stripe.UpdateItemQuantity(ctx, itemID, diskUnits); err != nil {
				slog.Error("metering: prod disk update failed", "workspace", wsID, "error", err)
			}
		}

		slog.Info("metering: prod allocations reported",
			"workspace", wsID,
			"cpu_vcpu", cpuUnits,
			"memory_gb", memUnits,
			"disk_gb", diskUnits,
		)
	}

	return nil
}
