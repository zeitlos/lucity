package metering

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/logto"
	stripelib "github.com/zeitlos/lucity/services/cashier/stripe"
)

// maxBackfillDays is the maximum number of days we can backfill (Stripe's meter event limit).
const maxBackfillDays = 35

// meterWindow is the fixed billing window size. Usage is always aggregated in 1-hour
// windows regardless of tick interval. This ensures consistent billing: a 5m tick
// interval reports the same totals as a 1h tick interval — the interval only controls
// how quickly we detect a completed window, not what we report.
const meterWindow = time.Hour

// ingestionDelay is the time to wait after a window closes before querying SigNoz.
// OTel collectors scrape every ~60s and batch exports to ClickHouse. 5 minutes
// gives enough margin for all metrics to land before we query.
const ingestionDelay = 5 * time.Minute

const checkpointCMName = "metering-checkpoint"
const checkpointKey = "last_window_end"

// Worker periodically queries resource usage and reports it to Stripe.
type Worker struct {
	stripe   *stripelib.Client
	deployer deployer.DeployerServiceClient
	logto    *logto.Client
	signoz   *SigNozClient
	k8s      kubernetes.Interface // nil if K8s not available (no checkpoint/backfill)
	issuer   *auth.Issuer
	interval time.Duration
	cancel   context.CancelFunc
	done     chan struct{}
}

// NewWorker creates a metering worker. k8sClient may be nil (disables checkpoint/backfill).
func NewWorker(stripeClient *stripelib.Client, deployerClient deployer.DeployerServiceClient, logtoClient *logto.Client, signozClient *SigNozClient, k8sClient kubernetes.Interface, issuer *auth.Issuer, interval time.Duration) *Worker {
	return &Worker{
		stripe:   stripeClient,
		deployer: deployerClient,
		logto:    logtoClient,
		signoz:   signozClient,
		k8s:      k8sClient,
		issuer:   issuer,
		interval: interval,
		done:     make(chan struct{}),
	}
}

func (w *Worker) Label() string { return "Metering Worker" }

func (w *Worker) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	defer close(w.done)

	// Backfill missed windows on startup.
	caught := w.backfill(ctx)

	// Only process the current window if backfill is caught up.
	// Otherwise the checkpoint would jump ahead, skipping the gap.
	if caught {
		w.tick(ctx)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	slog.Info("metering: next tick", "in", w.interval, "at", time.Now().Add(w.interval).Format(time.RFC3339))

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			w.tick(ctx)
			slog.Info("metering: next tick", "in", w.interval, "at", time.Now().Add(w.interval).Format(time.RFC3339))
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

// alignWindow returns the start and end of the most recently completed 1-hour window.
// At 14:37, this returns (13:00, 14:00). The window size is always meterWindow (1h),
// independent of tick interval.
func alignWindow(now time.Time) (start, end time.Time) {
	windowSec := int64(meterWindow.Seconds())
	nowUnix := now.Unix()
	endUnix := nowUnix - (nowUnix % windowSec)
	return time.Unix(endUnix-windowSec, 0).UTC(), time.Unix(endUnix, 0).UTC()
}

// backfill processes all missed metering windows since the last checkpoint.
// Returns true if backfill completed (caught up to the current window) or
// was not needed. Returns false if interrupted (e.g. shutdown) before finishing.
func (w *Worker) backfill(ctx context.Context) bool {
	if w.k8s == nil {
		return true
	}

	lastEnd := w.lastCheckpoint(ctx)
	if lastEnd.IsZero() {
		slog.Info("metering: no checkpoint found, skipping backfill")
		return true
	}

	_, currentEnd := alignWindow(time.Now().Add(-ingestionDelay))

	// No gap — already caught up.
	if !lastEnd.Before(currentEnd) {
		return true
	}

	// Cap backfill to Stripe's limit.
	earliest := time.Now().Add(-maxBackfillDays * 24 * time.Hour)
	if lastEnd.Before(earliest) {
		slog.Warn("metering: checkpoint older than 35 days, capping backfill",
			"checkpoint", lastEnd, "capped_to", earliest)
		// Re-align the capped time to a proper window boundary.
		_, lastEnd = alignWindow(earliest)
	}

	// Process each missed window in chronological order.
	windowStart := lastEnd
	var count int
	for windowStart.Add(meterWindow).Before(currentEnd) || windowStart.Add(meterWindow).Equal(currentEnd) {
		// Check for shutdown between windows.
		select {
		case <-ctx.Done():
			slog.Info("metering: backfill interrupted", "completed", count)
			return false
		default:
		}

		windowEnd := windowStart.Add(meterWindow)
		count++
		slog.Info("metering: backfilling window", "start", windowStart, "end", windowEnd, "window", count)
		w.processWindow(ctx, windowStart, windowEnd)
		windowStart = windowEnd
	}
	if count > 0 {
		slog.Info("metering: backfill complete", "windows", count)
	}
	return true
}

func (w *Worker) tick(ctx context.Context) {
	// Delay window selection so SigNoz has time to fully ingest metrics.
	// At 15:05 with 5min delay: alignWindow(15:00) → processes 14:00-15:00.
	start, end := alignWindow(time.Now().Add(-ingestionDelay))
	slog.Info("metering tick", "window_start", start, "window_end", end)
	w.processWindow(ctx, start, end)
}

// deployerCtx creates a system-level auth context for calling the deployer.
func (w *Worker) deployerCtx(ctx context.Context) context.Context {
	ctx = auth.WithClaims(ctx, &auth.Claims{
		Subject: "cashier",
		Roles:   []auth.Role{auth.RoleUser},
	})
	ctx = auth.WithIssuer(ctx, w.issuer)
	return auth.OutgoingContext(ctx)
}

// billableWorkspaces returns workspaces that have both stripeCustomerId and
// stripeSubscriptionId in their Logto org customData.
func (w *Worker) billableWorkspaces(ctx context.Context) (map[string]*workspaceData, error) {
	orgs, err := w.logto.Organizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list logto organizations: %w", err)
	}

	result := make(map[string]*workspaceData)
	for _, org := range orgs {
		if org.CustomData == nil {
			continue
		}
		customerID, _ := org.CustomData["stripeCustomerId"].(string)
		subscriptionID, _ := org.CustomData["stripeSubscriptionId"].(string)
		if customerID == "" || subscriptionID == "" {
			continue
		}
		result[org.Name] = &workspaceData{
			customerID:     customerID,
			subscriptionID: subscriptionID,
		}
	}
	return result, nil
}

// workspaceData holds billing metadata for a workspace.
type workspaceData struct {
	customerID      string
	subscriptionID  string
	ecoNamespaces   []string
	ecoAllocations  []allocEntry
	prodAllocations []allocEntry
}

type allocEntry struct {
	namespace string
	cpuMillis int32
	memoryMB  int32
	diskMB    int32
}

func (w *Worker) processWindow(ctx context.Context, windowStart, windowEnd time.Time) {
	start := time.Now()
	callCtx := w.deployerCtx(ctx)

	// 1. List all billable workspaces from Logto.
	workspaces, err := w.billableWorkspaces(ctx)
	if err != nil {
		slog.Error("metering: failed to list billable workspaces", "error", err)
		return
	}

	if len(workspaces) == 0 {
		slog.Info("metering: no billable workspaces")
		w.saveCheckpointQuiet(ctx, windowEnd)
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
		cpuByNs, err = w.signoz.CPUByNamespace(ctx, allEcoNamespaces, windowStart, windowEnd)
		if err != nil {
			slog.Error("metering: failed to query CPU usage", "error", err)
			cpuByNs = make(map[string]float64)
		}

		memByNs, err = w.signoz.MemoryByNamespace(ctx, allEcoNamespaces, windowStart, windowEnd)
		if err != nil {
			slog.Error("metering: failed to query memory usage", "error", err)
			memByNs = make(map[string]float64)
		}

		diskByNs, err = w.signoz.DiskByNamespace(ctx, allEcoNamespaces, windowStart, windowEnd)
		if err != nil {
			slog.Error("metering: failed to query disk usage", "error", err)
			diskByNs = make(map[string]float64)
		}
	}

	// 6. Report to Stripe per workspace and ensure credit grants exist.
	var anyError bool
	for wsID, ws := range workspaces {
		if err := w.reportWorkspace(ctx, wsID, ws, cpuByNs, memByNs, diskByNs, windowStart, windowEnd); err != nil {
			slog.Error("metering: failed to report workspace", "workspace", wsID, "error", err)
			anyError = true
		}

		// Ensure a credit grant exists for the current billing period.
		w.ensureCreditGrant(ctx, wsID, ws)
	}

	// Only checkpoint if all workspaces succeeded — failed windows will be retried,
	// and Stripe's identifier dedup prevents double-counting for already-reported events.
	if anyError {
		slog.Warn("metering: skipping checkpoint due to errors, will retry next tick",
			"window_start", windowStart, "window_end", windowEnd)
	} else {
		w.saveCheckpointQuiet(ctx, windowEnd)
	}

	slog.Info("metering tick completed",
		"duration", time.Since(start),
		"workspaces", len(workspaces),
		"window_start", windowStart,
		"window_end", windowEnd,
	)
}

func (w *Worker) reportWorkspace(ctx context.Context, wsID string, ws *workspaceData, cpuByNs, memByNs, diskByNs map[string]float64, windowStart, windowEnd time.Time) error {
	intervalMinutes := windowEnd.Sub(windowStart).Minutes()
	timestamp := windowEnd.Unix()

	// Report eco usage from SigNoz via Billing Meter events.
	if len(ws.ecoNamespaces) > 0 {
		var totalCPUSeconds, totalMemBytes, totalDiskBytes float64
		for _, ns := range ws.ecoNamespaces {
			totalCPUSeconds += cpuByNs[ns]
			totalMemBytes += memByNs[ns]
			totalDiskBytes += diskByNs[ns]
		}

		// CPU: seconds -> vCPU-minutes (Stripe unit).
		cpuMinutes := int64(math.Ceil(totalCPUSeconds / 60))
		if cpuMinutes > 0 && w.stripe.Meters.EcoCPUEventName != "" {
			id := meterEventID(wsID, w.stripe.Meters.EcoCPUEventName, windowStart)
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.EcoCPUEventName, ws.customerID, cpuMinutes, timestamp, id); err != nil {
				slog.Error("metering: eco CPU report failed", "workspace", wsID, "error", err)
			}
		}

		// Memory: avg bytes -> GB-minutes.
		gbMinutes := int64(math.Ceil(totalMemBytes / (1024 * 1024 * 1024) * intervalMinutes))
		if gbMinutes > 0 && w.stripe.Meters.EcoMemEventName != "" {
			id := meterEventID(wsID, w.stripe.Meters.EcoMemEventName, windowStart)
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.EcoMemEventName, ws.customerID, gbMinutes, timestamp, id); err != nil {
				slog.Error("metering: eco memory report failed", "workspace", wsID, "error", err)
			}
		}

		// Disk: bytes -> GB.
		diskGB := int64(math.Ceil(totalDiskBytes / (1024 * 1024 * 1024)))
		if diskGB > 0 && w.stripe.Meters.EcoDiskEventName != "" {
			id := meterEventID(wsID, w.stripe.Meters.EcoDiskEventName, windowStart)
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.EcoDiskEventName, ws.customerID, diskGB, timestamp, id); err != nil {
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

	// Report production allocations via Billing Meter events.
	if len(ws.prodAllocations) > 0 {
		var totalCPUMillis, totalMemMB, totalDiskMB int32
		for _, alloc := range ws.prodAllocations {
			totalCPUMillis += alloc.cpuMillis
			totalMemMB += alloc.memoryMB
			totalDiskMB += alloc.diskMB
		}

		// CPU: millicores -> vCPU-minutes (allocation × interval).
		cpuMinutes := int64(math.Ceil(float64(totalCPUMillis) / 1000 * intervalMinutes))
		if cpuMinutes > 0 && w.stripe.Meters.ProdCPUEventName != "" {
			id := meterEventID(wsID, w.stripe.Meters.ProdCPUEventName, windowStart)
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.ProdCPUEventName, ws.customerID, cpuMinutes, timestamp, id); err != nil {
				slog.Error("metering: prod CPU report failed", "workspace", wsID, "error", err)
			}
		}

		// Memory: MB -> GB-minutes (allocation × interval).
		memGBMinutes := int64(math.Ceil(float64(totalMemMB) / 1024 * intervalMinutes))
		if memGBMinutes > 0 && w.stripe.Meters.ProdMemEventName != "" {
			id := meterEventID(wsID, w.stripe.Meters.ProdMemEventName, windowStart)
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.ProdMemEventName, ws.customerID, memGBMinutes, timestamp, id); err != nil {
				slog.Error("metering: prod memory report failed", "workspace", wsID, "error", err)
			}
		}

		// Disk: MB -> GB (point-in-time, same as eco).
		diskGB := int64(math.Ceil(float64(totalDiskMB) / 1024))
		if diskGB > 0 && w.stripe.Meters.ProdDiskEventName != "" {
			id := meterEventID(wsID, w.stripe.Meters.ProdDiskEventName, windowStart)
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.ProdDiskEventName, ws.customerID, diskGB, timestamp, id); err != nil {
				slog.Error("metering: prod disk report failed", "workspace", wsID, "error", err)
			}
		}

		slog.Info("metering: prod usage reported",
			"workspace", wsID,
			"cpu_minutes", cpuMinutes,
			"memory_gb_minutes", memGBMinutes,
			"disk_gb", diskGB,
		)
	}

	return nil
}

// meterEventID returns a deterministic identifier for Stripe meter event deduplication.
// Format: {workspace}:{eventName}:{windowStartUnix}
// Same logical usage period always produces the same ID — Stripe rejects duplicates within 24h.
func meterEventID(workspace, eventName string, windowStart time.Time) string {
	return fmt.Sprintf("%s:%s:%d", workspace, eventName, windowStart.Unix())
}

// ensureCreditGrant creates a Stripe credit grant for the current billing period if one doesn't exist.
func (w *Worker) ensureCreditGrant(ctx context.Context, wsID string, ws *workspaceData) {
	sub, err := w.stripe.Subscription(ctx, ws.subscriptionID)
	if err != nil {
		slog.Error("metering: failed to get subscription for credit grant", "workspace", wsID, "error", err)
		return
	}

	// Find the plan price and period boundaries from subscription items.
	var planPriceID string
	var periodStart, periodEnd int64
	for _, item := range sub.Items.Data {
		if item.Price.ID == w.stripe.Prices.HobbyPriceID || item.Price.ID == w.stripe.Prices.ProPriceID {
			planPriceID = item.Price.ID
		}
		if periodEnd == 0 && item.CurrentPeriodEnd > 0 {
			periodStart = item.CurrentPeriodStart
			periodEnd = item.CurrentPeriodEnd
		}
	}
	if planPriceID == "" || periodEnd == 0 {
		return
	}

	creditCents := stripelib.PlanCreditCents(planPriceID, w.stripe.Prices)

	if err := w.stripe.CreateCreditGrantForPeriod(ctx, ws.customerID, creditCents, periodStart, periodEnd); err != nil {
		slog.Error("metering: failed to ensure credit grant", "workspace", wsID, "error", err)
	}
}

// lastCheckpoint reads the last successfully completed window end time from K8s.
// Returns zero time if no checkpoint exists (first run or lost state).
func (w *Worker) lastCheckpoint(ctx context.Context) time.Time {
	if w.k8s == nil {
		return time.Time{}
	}

	cm, err := w.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).Get(ctx, checkpointCMName, metav1.GetOptions{})
	if err != nil {
		return time.Time{}
	}

	ts, err := strconv.ParseInt(cm.Data[checkpointKey], 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

// saveCheckpoint writes the window end time to the ConfigMap (create or update).
func (w *Worker) saveCheckpoint(ctx context.Context, windowEnd time.Time) error {
	if w.k8s == nil {
		return nil
	}

	data := map[string]string{checkpointKey: strconv.FormatInt(windowEnd.Unix(), 10)}

	cm, err := w.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).Get(ctx, checkpointCMName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = w.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      checkpointCMName,
				Namespace: labels.LucityNamespace,
				Labels: map[string]string{
					labels.ManagedBy: labels.ManagedByLucity,
				},
			},
			Data: data,
		}, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return fmt.Errorf("failed to get checkpoint configmap: %w", err)
	}

	cm.Data = data
	_, err = w.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

// saveCheckpointQuiet saves the checkpoint and logs errors without returning them.
func (w *Worker) saveCheckpointQuiet(ctx context.Context, windowEnd time.Time) {
	if err := w.saveCheckpoint(ctx, windowEnd); err != nil {
		slog.Error("metering: failed to save checkpoint", "error", err)
	}
}
