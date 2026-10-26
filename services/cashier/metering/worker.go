package metering

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"time"

	gostripe "github.com/stripe/stripe-go/v87"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/conductor"
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

// ingestionDelay is the time to wait after a window closes before querying VictoriaMetrics.
// OTel collectors scrape every ~60s and batch exports. 5 minutes gives enough
// margin for all metrics to land before we query.
const ingestionDelay = 5 * time.Minute

const checkpointCMName = "metering-checkpoint"
const checkpointKey = "last_window_end"

// Suspender suspends or resumes every environment of a workspace.
type Suspender interface {
	SuspendWorkspace(workspace string, suspended bool) error
}

// Worker periodically queries resource usage and reports it to Stripe.
type Worker struct {
	stripe     *stripelib.Client
	billing    *billing.Client
	conductor  conductor.ConductorServiceClient
	logto      *logto.Client
	vm         *VMClient
	k8s        kubernetes.Interface // nil if K8s not available (no checkpoint/backfill)
	suspender  Suspender
	namespace  string
	issuer     *auth.Issuer
	interval   time.Duration
	checkpoint time.Time
	cancel     context.CancelFunc
	done       chan struct{}
}

// NewWorker creates a metering worker. k8sClient may be nil (disables checkpoint/backfill).
func NewWorker(stripeClient *stripelib.Client, billingClient *billing.Client, conductorClient conductor.ConductorServiceClient, logtoClient *logto.Client, vmClient *VMClient, k8sClient kubernetes.Interface, suspender Suspender, namespace string, issuer *auth.Issuer, interval time.Duration) *Worker {
	return &Worker{
		stripe:    stripeClient,
		billing:   billingClient,
		conductor: conductorClient,
		logto:     logtoClient,
		vm:        vmClient,
		k8s:       k8sClient,
		suspender: suspender,
		namespace: namespace,
		issuer:    issuer,
		interval:  interval,
		done:      make(chan struct{}),
	}
}

func (w *Worker) Label() string { return "Metering Worker" }

func (w *Worker) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	defer close(w.done)

	w.catchUp(ctx)
	w.endTrials(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	slog.Info("metering: next tick", "in", w.interval, "at", time.Now().Add(w.interval).Format(time.RFC3339))

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			w.catchUp(ctx)
			w.endTrials(ctx)
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

func (w *Worker) catchUp(ctx context.Context) {
	if w.checkpoint.IsZero() {
		checkpoint, err := w.lastCheckpoint(ctx)
		if err != nil {
			slog.Error("metering: failed to read checkpoint, retrying next tick", "error", err)
			return
		}
		w.checkpoint = checkpoint
	}

	_, currentEnd := alignWindow(time.Now().Add(-ingestionDelay))

	windowStart := w.checkpoint
	if windowStart.IsZero() {
		slog.Info("metering: no checkpoint found, starting at the current window")
		windowStart = currentEnd.Add(-meterWindow)
	}

	earliest := time.Now().Add(-maxBackfillDays * 24 * time.Hour)
	if windowStart.Before(earliest) {
		slog.Warn("metering: checkpoint older than 35 days, capping backfill",
			"checkpoint", windowStart, "capped_to", earliest)
		_, windowStart = alignWindow(earliest)
	}

	for !windowStart.Add(meterWindow).After(currentEnd) {
		if ctx.Err() != nil {
			return
		}

		windowEnd := windowStart.Add(meterWindow)
		slog.Info("metering tick", "window_start", windowStart, "window_end", windowEnd)
		if err := w.processWindow(ctx, windowStart, windowEnd); err != nil {
			slog.Error("metering: window failed, retrying next tick",
				"window_start", windowStart, "window_end", windowEnd, "error", err)
			return
		}

		w.checkpoint = windowEnd
		w.saveCheckpointQuiet(ctx, windowEnd)
		windowStart = windowEnd
	}
}

// conductorCtx creates a system-level auth context for calling conductor.
func (w *Worker) conductorCtx(ctx context.Context) context.Context {
	ctx = auth.NewContext(ctx, &auth.Claims{
		Subject: "cashier",
	})
	ctx = auth.WithIssuer(ctx, w.issuer)
	return auth.OutgoingContext(ctx)
}

// billableWorkspaces returns workspaces that have a stripeCustomerId in their
// Logto org customData.
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
		if customerID == "" {
			continue
		}
		suspended, _ := org.CustomData["suspended"].(bool)
		result[org.Name] = &workspaceData{
			customerID: customerID,
			suspended:  suspended,
		}
	}
	return result, nil
}

// workspaceData holds billing metadata for a workspace.
type workspaceData struct {
	customerID      string
	suspended       bool
	namespaces      []string
	ecoNamespaces   []string
	ecoAllocations  []allocEntry
	prodAllocations []allocEntry
	storageBytes    float64
	egressBytes     float64
}

type allocEntry struct {
	namespace string
	cpuMillis int32
	memoryMB  int32
	diskMB    int32
}

func (w *Worker) processWindow(ctx context.Context, windowStart, windowEnd time.Time) error {
	start := time.Now()
	callCtx := w.conductorCtx(ctx)

	// 1. List all billable workspaces from Logto.
	workspaces, err := w.billableWorkspaces(ctx)
	if err != nil {
		return err
	}

	if len(workspaces) == 0 {
		slog.Info("metering: no billable workspaces")
		return nil
	}

	// 3. List all resource allocations.
	allocResp, err := w.conductor.ListResourceAllocations(callCtx, &conductor.ListResourceAllocationsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list resource allocations: %w", err)
	}

	// 4. Group allocations by workspace.
	var allEcoNamespaces []string
	for _, alloc := range allocResp.Allocations {
		ws, ok := workspaces[alloc.Workspace]
		if !ok {
			continue
		}

		ns := alloc.Namespace
		ws.namespaces = append(ws.namespaces, ns)
		entry := allocEntry{
			namespace: ns,
			cpuMillis: alloc.CpuMillicores,
			memoryMB:  alloc.MemoryMb,
			diskMB:    alloc.DiskMb,
		}

		if alloc.Tier == conductor.ResourceTier_RESOURCE_TIER_ECO {
			ws.ecoNamespaces = append(ws.ecoNamespaces, ns)
			ws.ecoAllocations = append(ws.ecoAllocations, entry)
			allEcoNamespaces = append(allEcoNamespaces, ns)
		} else {
			ws.prodAllocations = append(ws.prodAllocations, entry)
		}
	}

	// 5. Query VictoriaMetrics for eco namespace usage.
	var cpuByNs, memByNs, diskByNs map[string]float64
	if len(allEcoNamespaces) > 0 {
		cpuByNs, err = w.vm.CPUByNamespace(ctx, windowStart, windowEnd)
		if err != nil {
			return fmt.Errorf("failed to query CPU usage: %w", err)
		}

		memByNs, err = w.vm.MemoryByNamespace(ctx, windowStart, windowEnd)
		if err != nil {
			return fmt.Errorf("failed to query memory usage: %w", err)
		}

		diskByNs, err = w.vm.DiskByNamespace(ctx, windowStart, windowEnd)
		if err != nil {
			return fmt.Errorf("failed to query disk usage: %w", err)
		}
	}

	storageByWorkspace, err := w.vm.StorageByWorkspace(ctx, windowStart, windowEnd)
	if err != nil {
		return fmt.Errorf("failed to query storage usage: %w", err)
	}

	egressByNs, err := w.vm.EgressByNamespace(ctx, windowStart, windowEnd)
	if err != nil {
		return fmt.Errorf("failed to query egress usage: %w", err)
	}

	for wsID, ws := range workspaces {
		ws.storageBytes = storageByWorkspace[wsID]
		for _, ns := range ws.namespaces {
			ws.egressBytes += egressByNs[ns]
		}
	}

	subscriptions, err := w.billing.CurrentSubscriptions(ctx)
	if err != nil {
		slog.Error("metering: failed to list subscriptions, skipping credit grants", "error", err)
	}

	// 6. Report to Stripe per workspace and ensure credit grants exist.
	for wsID, ws := range workspaces {
		w.reportWorkspace(ctx, wsID, ws, cpuByNs, memByNs, diskByNs, windowStart, windowEnd)

		// Ensure a credit grant exists for the current billing period.
		if sub := subscriptions[ws.customerID]; sub != nil {
			w.ensureCreditGrant(ctx, wsID, ws, sub)
		}
	}

	slog.Info("metering tick completed",
		"duration", time.Since(start),
		"workspaces", len(workspaces),
		"window_start", windowStart,
		"window_end", windowEnd,
	)
	return nil
}

func (w *Worker) reportWorkspace(ctx context.Context, wsID string, ws *workspaceData, cpuByNs, memByNs, diskByNs map[string]float64, windowStart, windowEnd time.Time) {
	intervalMinutes := windowEnd.Sub(windowStart).Minutes()
	timestamp := windowEnd.Unix()

	// Report eco usage from VictoriaMetrics via Billing Meter events.
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
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.EcoCPUEventName, ws.customerID, float64(cpuMinutes), timestamp, id); err != nil {
				slog.Error("metering: eco CPU report failed", "workspace", wsID, "error", err)
			}
		}

		// Memory: avg bytes -> GB-minutes.
		gbMinutes := int64(math.Ceil(totalMemBytes / (1024 * 1024 * 1024) * intervalMinutes))
		if gbMinutes > 0 && w.stripe.Meters.EcoMemEventName != "" {
			id := meterEventID(wsID, w.stripe.Meters.EcoMemEventName, windowStart)
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.EcoMemEventName, ws.customerID, float64(gbMinutes), timestamp, id); err != nil {
				slog.Error("metering: eco memory report failed", "workspace", wsID, "error", err)
			}
		}

		// Disk: bytes -> GB.
		diskGB := int64(math.Ceil(totalDiskBytes / (1024 * 1024 * 1024)))
		if diskGB > 0 && w.stripe.Meters.EcoDiskEventName != "" {
			id := meterEventID(wsID, w.stripe.Meters.EcoDiskEventName, windowStart)
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.EcoDiskEventName, ws.customerID, float64(diskGB), timestamp, id); err != nil {
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
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.ProdCPUEventName, ws.customerID, float64(cpuMinutes), timestamp, id); err != nil {
				slog.Error("metering: prod CPU report failed", "workspace", wsID, "error", err)
			}
		}

		// Memory: MB -> GB-minutes (allocation × interval).
		memGBMinutes := int64(math.Ceil(float64(totalMemMB) / 1024 * intervalMinutes))
		if memGBMinutes > 0 && w.stripe.Meters.ProdMemEventName != "" {
			id := meterEventID(wsID, w.stripe.Meters.ProdMemEventName, windowStart)
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.ProdMemEventName, ws.customerID, float64(memGBMinutes), timestamp, id); err != nil {
				slog.Error("metering: prod memory report failed", "workspace", wsID, "error", err)
			}
		}

		// Disk: MB -> GB (point-in-time, same as eco).
		diskGB := int64(math.Ceil(float64(totalDiskMB) / 1024))
		if diskGB > 0 && w.stripe.Meters.ProdDiskEventName != "" {
			id := meterEventID(wsID, w.stripe.Meters.ProdDiskEventName, windowStart)
			if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.ProdDiskEventName, ws.customerID, float64(diskGB), timestamp, id); err != nil {
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

	storageGBHours := roundUsage(ws.storageBytes / (1024 * 1024 * 1024) * windowEnd.Sub(windowStart).Hours())
	if storageGBHours > 0 && w.stripe.Meters.StorageEventName != "" {
		id := meterEventID(wsID, w.stripe.Meters.StorageEventName, windowStart)
		if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.StorageEventName, ws.customerID, storageGBHours, timestamp, id); err != nil {
			slog.Error("metering: storage report failed", "workspace", wsID, "error", err)
		}
	}

	egressGB := roundUsage(ws.egressBytes / (1024 * 1024 * 1024))
	if egressGB > 0 && w.stripe.Meters.EgressEventName != "" {
		id := meterEventID(wsID, w.stripe.Meters.EgressEventName, windowStart)
		if err := w.stripe.ReportMeterEvent(ctx, w.stripe.Meters.EgressEventName, ws.customerID, egressGB, timestamp, id); err != nil {
			slog.Error("metering: egress report failed", "workspace", wsID, "error", err)
		}
	}

	if storageGBHours > 0 || egressGB > 0 {
		slog.Info("metering: storage and egress usage reported",
			"workspace", wsID,
			"storage_gb_hours", storageGBHours,
			"egress_gb", egressGB,
		)
	}
}

func roundUsage(value float64) float64 {
	return math.Round(value*1e6) / 1e6
}

// meterEventID returns a deterministic identifier for Stripe meter event deduplication.
// Format: {workspace}:{eventName}:{windowStartUnix}
// Same logical usage period always produces the same ID — Stripe rejects duplicates within 24h.
func meterEventID(workspace, eventName string, windowStart time.Time) string {
	return fmt.Sprintf("%s:%s:%d", workspace, eventName, windowStart.Unix())
}

// ensureCreditGrant creates a Stripe credit grant for the current billing period if one doesn't exist.
func (w *Worker) ensureCreditGrant(ctx context.Context, wsID string, ws *workspaceData, sub *gostripe.Subscription) {
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

func (w *Worker) endTrials(ctx context.Context) {
	workspaces, err := w.billableWorkspaces(ctx)
	if err != nil {
		slog.Error("trials: failed to list workspaces", "error", err)
		return
	}

	subscriptions, err := w.billing.CurrentSubscriptions(ctx)
	if err != nil {
		slog.Error("trials: failed to list subscriptions", "error", err)
		return
	}

	now := time.Now()
	for wsID, ws := range workspaces {
		if ws.suspended {
			continue
		}

		sub := subscriptions[ws.customerID]
		if sub == nil || w.billing.Plan(sub).Paid() {
			continue
		}

		end := trialEnd(sub)
		if end.IsZero() || now.Before(end) {
			continue
		}

		hasPaymentMethod, err := w.stripe.HasPaymentMethod(ctx, sub)
		if err != nil {
			slog.Error("trials: failed to check payment method", "workspace", wsID, "error", err)
			continue
		}
		if hasPaymentMethod {
			continue
		}

		slog.Info("trial ended, no plan or payment method, suspending workspace",
			"workspace", wsID, "subscription", sub.ID, "trial_end", end)
		if err := w.suspender.SuspendWorkspace(wsID, true); err != nil {
			slog.Error("trials: failed to suspend workspace, retrying next tick", "workspace", wsID, "error", err)
		}
	}
}

func trialEnd(sub *gostripe.Subscription) time.Time {
	interval := sub.PendingInvoiceItemInterval
	if interval == nil || interval.Interval != gostripe.SubscriptionPendingInvoiceItemIntervalIntervalDay {
		return time.Time{}
	}
	return time.Unix(sub.StartDate, 0).AddDate(0, 0, int(interval.IntervalCount))
}

// lastCheckpoint reads the last successfully completed window end time from K8s.
// Returns zero time if no checkpoint exists (first run or lost state).
func (w *Worker) lastCheckpoint(ctx context.Context) (time.Time, error) {
	if w.k8s == nil {
		return time.Time{}, nil
	}

	cm, err := w.k8s.CoreV1().ConfigMaps(w.namespace).Get(ctx, checkpointCMName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get checkpoint configmap: %w", err)
	}

	ts, err := strconv.ParseInt(cm.Data[checkpointKey], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse checkpoint %q: %w", cm.Data[checkpointKey], err)
	}
	return time.Unix(ts, 0).UTC(), nil
}

// saveCheckpoint writes the window end time to the ConfigMap (create or update).
func (w *Worker) saveCheckpoint(ctx context.Context, windowEnd time.Time) error {
	if w.k8s == nil {
		return nil
	}

	data := map[string]string{checkpointKey: strconv.FormatInt(windowEnd.Unix(), 10)}

	cm, err := w.k8s.CoreV1().ConfigMaps(w.namespace).Get(ctx, checkpointCMName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = w.k8s.CoreV1().ConfigMaps(w.namespace).Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      checkpointCMName,
				Namespace: w.namespace,
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
	_, err = w.k8s.CoreV1().ConfigMaps(w.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

// saveCheckpointQuiet saves the checkpoint and logs errors without returning them.
func (w *Worker) saveCheckpointQuiet(ctx context.Context, windowEnd time.Time) {
	if err := w.saveCheckpoint(ctx, windowEnd); err != nil {
		slog.Error("metering: failed to save checkpoint", "error", err)
	}
}
