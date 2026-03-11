package grpc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/deployer/argocd"
	"github.com/zeitlos/lucity/services/deployer/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	deployer.UnimplementedDeployerServiceServer
	argo *argocd.Client
	k8s  kubernetes.Interface

	// softServeHTTP is the cluster-internal Soft-serve HTTP URL for ArgoCD to clone from.
	softServeHTTP string
	// softServeToken is the Soft-serve access token for HTTP git operations.
	softServeToken string
}

func NewServer(argo *argocd.Client, softServeHTTP, softServeToken string, k8s kubernetes.Interface) *Server {
	return &Server{
		argo:           argo,
		k8s:            k8s,
		softServeHTTP:  softServeHTTP,
		softServeToken: softServeToken,
	}
}

func (s *Server) DeployEnvironment(ctx context.Context, req *deployer.DeployEnvironmentRequest) (*deployer.DeployEnvironmentResponse, error) {
	ws := tenant.FromContext(ctx)
	appName := applicationName(ws, req.Project, req.Environment)

	// Idempotent: if the application already exists, return it.
	existing, err := s.argo.Application(ctx, appName)
	if err == nil && existing != nil {
		slog.Info("ArgoCD application already exists",
			"app", appName,
			"project", req.Project,
			"environment", req.Environment,
		)
		return &deployer.DeployEnvironmentResponse{
			DeploymentName: existing.Metadata.Name,
		}, nil
	}

	// Ensure the namespace exists with platform labels.
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.TargetNamespace,
			Labels: map[string]string{
				labels.Workspace:   ws,
				labels.Project:     req.Project,
				labels.Environment: req.Environment,
				labels.ManagedBy:   labels.ManagedByLucity,
			},
		},
	}
	if _, err := s.k8s.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("failed to create namespace %s: %w", req.TargetNamespace, err)
	}

	// Ensure the GitOps repo is registered in ArgoCD with credentials.
	repoURL := s.repoURL(ws, req.Project)
	if err := s.argo.CreateRepository(ctx, argocd.Repository{
		Repo:     repoURL,
		Username: "lucity",
		Password: s.softServeToken,
		Type:     "git",
	}); err != nil {
		return nil, fmt.Errorf("failed to register repository in ArgoCD: %w", err)
	}

	app := argocd.Application{
		Metadata: argocd.Metadata{
			Name: appName,
		},
		Spec: argocd.ApplicationSpec{
			Source: argocd.Source{
				RepoURL:        repoURL,
				Path:           "base",
				TargetRevision: "HEAD",
				Helm: &argocd.Helm{
					ValueFiles: []string{
						fmt.Sprintf("../environments/%s/values.yaml", req.Environment),
					},
				},
			},
			Destination: argocd.Destination{
				Server:    "https://kubernetes.default.svc",
				Namespace: req.TargetNamespace,
			},
			Project: "default",
			SyncPolicy: &argocd.SyncPolicy{
				Automated: &argocd.Automated{
					Prune:    true,
					SelfHeal: true,
				},
				SyncOptions: []string{"CreateNamespace=true"},
			},
		},
	}

	result, err := s.argo.CreateApplication(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to create ArgoCD application: %w", err)
	}

	slog.Info("created ArgoCD application",
		"app", result.Metadata.Name,
		"project", req.Project,
		"environment", req.Environment,
		"namespace", req.TargetNamespace,
	)

	return &deployer.DeployEnvironmentResponse{
		DeploymentName: result.Metadata.Name,
	}, nil
}

func (s *Server) RemoveDeployment(ctx context.Context, req *deployer.RemoveDeploymentRequest) (*deployer.RemoveDeploymentResponse, error) {
	ws := tenant.FromContext(ctx)
	appName := applicationName(ws, req.Project, req.Environment)

	if err := s.argo.DeleteApplication(ctx, appName, true); err != nil {
		// Idempotent: if the application is already gone, that's fine — still clean up the namespace.
		if !strings.Contains(err.Error(), "404") {
			return nil, fmt.Errorf("failed to delete ArgoCD application: %w", err)
		}
		slog.Info("ArgoCD application already deleted", "app", appName)
	} else {
		slog.Info("deleted ArgoCD application", "app", appName)
	}

	// Delete the namespace. ArgoCD cascade already cleaned up resources inside.
	ns := labels.NamespaceFor(ws, req.Project, req.Environment)
	if err := s.k8s.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to delete namespace %s: %w", ns, err)
	}
	slog.Info("deleted namespace", "namespace", ns)

	return &deployer.RemoveDeploymentResponse{}, nil
}

func (s *Server) DeleteRepository(ctx context.Context, req *deployer.DeleteRepositoryRequest) (*deployer.DeleteRepositoryResponse, error) {
	ws := tenant.FromContext(ctx)
	repoURL := s.repoURL(ws, req.Project)

	if err := s.argo.DeleteRepository(ctx, repoURL); err != nil {
		return nil, fmt.Errorf("failed to delete ArgoCD repository: %w", err)
	}

	slog.Info("deleted ArgoCD repository", "repo", repoURL)
	return &deployer.DeleteRepositoryResponse{}, nil
}

func (s *Server) GetDeploymentStatus(ctx context.Context, req *deployer.GetDeploymentStatusRequest) (*deployer.GetDeploymentStatusResponse, error) {
	ws := tenant.FromContext(ctx)
	appName := applicationName(ws, req.Project, req.Environment)

	app, err := s.argo.Application(ctx, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to get ArgoCD application: %w", err)
	}

	status, message := mapStatus(app.Status)

	return &deployer.GetDeploymentStatusResponse{
		Status:  status,
		Message: message,
	}, nil
}

func (s *Server) SyncDeployment(ctx context.Context, req *deployer.SyncDeploymentRequest) (*deployer.SyncDeploymentResponse, error) {
	ws := tenant.FromContext(ctx)
	appName := applicationName(ws, req.Project, req.Environment)

	app, err := s.argo.SyncApplication(ctx, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to sync ArgoCD application: %w", err)
	}

	slog.Info("triggered sync", "app", appName)

	status, _ := mapStatus(app.Status)
	return &deployer.SyncDeploymentResponse{
		Status: status,
	}, nil
}

// applicationName derives the ArgoCD Application name from workspace, project, and environment.
func applicationName(workspace, project, environment string) string {
	return labels.NamespaceFor(workspace, project, environment)
}

// repoURL returns the Soft-serve HTTP clone URL for a project's GitOps repo.
func (s *Server) repoURL(workspace, project string) string {
	return strings.TrimSuffix(s.softServeHTTP, "/") + "/" + workspace + "-" + project + "-gitops.git"
}

func (s *Server) ServiceLogs(req *deployer.ServiceLogsRequest, stream deployer.DeployerService_ServiceLogsServer) error {
	ctx := stream.Context()
	ws := tenant.FromContext(ctx)
	namespace := labels.NamespaceFor(ws, req.Project, req.Environment)

	labelSelector := fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s",
		req.Service, namespace)

	podList, err := s.k8s.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to list pods: %v", err)
	}
	if len(podList.Items) == 0 {
		return status.Errorf(codes.NotFound, "no pods found for service %q in %q", req.Service, namespace)
	}

	multiplePods := len(podList.Items) > 1

	tailLines := int64(1000)
	if req.TailLines > 0 {
		tailLines = int64(req.TailLines)
	}

	lines := make(chan *deployer.ServiceLogEntry, 128)
	var wg sync.WaitGroup

	for _, pod := range podList.Items {
		wg.Add(1)
		go func(podName string) {
			defer wg.Done()
			s.streamPodLogs(ctx, namespace, podName, req.Service, tailLines, multiplePods, lines)
		}(pod.Name)
	}

	go func() {
		wg.Wait()
		close(lines)
	}()

	for entry := range lines {
		if err := stream.Send(entry); err != nil {
			return err
		}
	}

	return nil
}

// streamPodLogs follows logs from a single pod/container and sends entries to the channel.
func (s *Server) streamPodLogs(ctx context.Context, namespace, podName, container string, tailLines int64, prefixPod bool, out chan<- *deployer.ServiceLogEntry) {
	opts := &corev1.PodLogOptions{
		Container: container,
		Follow:    true,
		TailLines: &tailLines,
	}

	logStream, err := s.k8s.CoreV1().Pods(namespace).GetLogs(podName, opts).Stream(ctx)
	if err != nil {
		slog.Warn("failed to open log stream", "pod", podName, "error", err)
		return
	}
	defer logStream.Close()

	podSuffix := shortPodID(podName)
	scanner := bufio.NewScanner(logStream)
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if prefixPod {
			line = fmt.Sprintf("[%s] %s", podSuffix, line)
		}
		select {
		case out <- &deployer.ServiceLogEntry{Line: line, Pod: podName}:
		case <-ctx.Done():
			return
		}
	}
}

// shortPodID extracts the unique suffix from a pod name.
func shortPodID(podName string) string {
	parts := strings.Split(podName, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return podName
}

const dbQueryTimeout = 30 * time.Second

func (s *Server) DatabaseTables(ctx context.Context, req *deployer.DatabaseTablesRequest) (*deployer.DatabaseTablesResponse, error) {
	creds, err := database.CredentialsFromSecret(ctx, s.k8s, tenant.FromContext(ctx), req.Project, req.Environment, req.Database)
	if err != nil {
		if errors.Is(err, database.ErrNotReady) {
			return nil, status.Errorf(codes.FailedPrecondition, "database is provisioning")
		}
		return nil, status.Errorf(codes.NotFound, "database credentials not found: %v", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()

	conn, err := database.Connect(queryCtx, creds)
	if err != nil {
		slog.Error("failed to connect to database", "project", req.Project, "environment", req.Environment, "database", req.Database, "error", err)
		return nil, status.Errorf(codes.Unavailable, "database connection failed")
	}
	defer conn.Close(queryCtx)

	tables, err := database.Tables(queryCtx, conn)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tables: %v", err)
	}

	return &deployer.DatabaseTablesResponse{Tables: tables}, nil
}

func (s *Server) DatabaseTableData(ctx context.Context, req *deployer.DatabaseTableDataRequest) (*deployer.DatabaseTableDataResponse, error) {
	creds, err := database.CredentialsFromSecret(ctx, s.k8s, tenant.FromContext(ctx), req.Project, req.Environment, req.Database)
	if err != nil {
		if errors.Is(err, database.ErrNotReady) {
			return nil, status.Errorf(codes.FailedPrecondition, "database is provisioning")
		}
		return nil, status.Errorf(codes.NotFound, "database credentials not found: %v", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()

	conn, err := database.Connect(queryCtx, creds)
	if err != nil {
		slog.Error("failed to connect to database", "project", req.Project, "environment", req.Environment, "database", req.Database, "error", err)
		return nil, status.Errorf(codes.Unavailable, "database connection failed")
	}
	defer conn.Close(queryCtx)

	columns, rows, estimatedRows, err := database.TableData(queryCtx, conn, req.Schema, req.Table, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query table data: %v", err)
	}

	return &deployer.DatabaseTableDataResponse{
		Columns:            columns,
		Rows:               rows,
		TotalEstimatedRows: estimatedRows,
	}, nil
}

func (s *Server) DatabaseQuery(ctx context.Context, req *deployer.DatabaseQueryRequest) (*deployer.DatabaseQueryResponse, error) {
	creds, err := database.CredentialsFromSecret(ctx, s.k8s, tenant.FromContext(ctx), req.Project, req.Environment, req.Database)
	if err != nil {
		if errors.Is(err, database.ErrNotReady) {
			return nil, status.Errorf(codes.FailedPrecondition, "database is provisioning")
		}
		return nil, status.Errorf(codes.NotFound, "database credentials not found: %v", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()

	conn, err := database.Connect(queryCtx, creds)
	if err != nil {
		slog.Error("failed to connect to database", "project", req.Project, "environment", req.Environment, "database", req.Database, "error", err)
		return nil, status.Errorf(codes.Unavailable, "database connection failed")
	}
	defer conn.Close(queryCtx)

	columns, rows, affected, err := database.Query(queryCtx, conn, req.Query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}

	return &deployer.DatabaseQueryResponse{
		Columns:      columns,
		Rows:         rows,
		AffectedRows: affected,
	}, nil
}

func (s *Server) DatabaseStatus(ctx context.Context, req *deployer.DatabaseStatusRequest) (*deployer.DatabaseStatusResponse, error) {
	ws := tenant.FromContext(ctx)
	namespace := labels.NamespaceFor(ws, req.Project, req.Environment)
	clusterName := req.Project + "-pg-" + req.Database
	labelSelector := fmt.Sprintf("cnpg.io/cluster=%s", clusterName)

	// Check CNPG pods for readiness.
	podList, err := s.k8s.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list database pods: %v", err)
	}

	var runningInstances int32
	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				runningInstances++
				break
			}
		}
	}

	// Read PVC info.
	var volumeInfo *deployer.VolumeInfo
	pvcList, err := s.k8s.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err == nil && len(pvcList.Items) > 0 {
		pvc := pvcList.Items[0]
		capacity := ""
		if qty, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
			capacity = qty.String()
		}
		requested := ""
		if qty, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			requested = qty.String()
		}
		volumeInfo = &deployer.VolumeInfo{
			Name:          pvc.Name,
			Size:          capacity,
			RequestedSize: requested,
		}

		// Try to get actual disk usage from kubelet stats.
		usedBytes, capacityBytes := s.pvcUsage(ctx, namespace, pvc.Name, podList.Items)
		volumeInfo.UsedBytes = usedBytes
		volumeInfo.CapacityBytes = capacityBytes
	}

	return &deployer.DatabaseStatusResponse{
		Ready:     runningInstances > 0,
		Instances: runningInstances,
		Volume:    volumeInfo,
	}, nil
}

func (s *Server) ServiceStatus(ctx context.Context, req *deployer.ServiceStatusRequest) (*deployer.ServiceStatusResponse, error) {
	ws := tenant.FromContext(ctx)
	namespace := labels.NamespaceFor(ws, req.Project, req.Environment)
	labelSelector := fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s",
		req.Service, namespace)

	deployList, err := s.k8s.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list deployments: %v", err)
	}

	var totalReplicas, totalReady int32
	for _, d := range deployList.Items {
		totalReplicas += d.Status.Replicas
		totalReady += d.Status.ReadyReplicas
	}

	return &deployer.ServiceStatusResponse{
		Ready:         totalReady > 0 && totalReady >= totalReplicas,
		Replicas:      totalReplicas,
		ReadyReplicas: totalReady,
	}, nil
}

// pvcUsage queries the kubelet stats API via the Kubernetes API proxy to get
// actual disk usage for a PVC. Returns (usedBytes, capacityBytes) or (0, 0)
// if the stats cannot be retrieved.
func (s *Server) pvcUsage(ctx context.Context, namespace, pvcName string, pods []corev1.Pod) (int64, int64) {
	// Find a running pod to query for volume stats.
	var nodeName string
	var podNamespace, podName string
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodRunning && pod.Spec.NodeName != "" {
			nodeName = pod.Spec.NodeName
			podNamespace = pod.Namespace
			podName = pod.Name
			break
		}
	}
	if nodeName == "" {
		return 0, 0
	}

	// Query kubelet stats via API server proxy.
	path := fmt.Sprintf("/api/v1/nodes/%s/proxy/stats/summary", nodeName)
	raw, err := s.k8s.CoreV1().RESTClient().Get().AbsPath(path).DoRaw(ctx)
	if err != nil {
		slog.Warn("failed to query kubelet stats", "node", nodeName, "error", err)
		return 0, 0
	}

	var summary kubeletStatsSummary
	if err := json.Unmarshal(raw, &summary); err != nil {
		slog.Warn("failed to parse kubelet stats", "error", err)
		return 0, 0
	}

	// Find the pod and its volume stats matching the PVC.
	for _, ps := range summary.Pods {
		if ps.PodRef.Namespace != podNamespace || ps.PodRef.Name != podName {
			continue
		}
		for _, vs := range ps.VolumeStats {
			if vs.PVCRef.Name == pvcName && vs.PVCRef.Namespace == namespace {
				return vs.UsedBytes, vs.CapacityBytes
			}
		}
	}
	return 0, 0
}

// kubelet stats API response types (minimal subset).

type kubeletStatsSummary struct {
	Pods []kubeletPodStats `json:"pods"`
}

type kubeletPodStats struct {
	PodRef      kubeletPodRef        `json:"podRef"`
	VolumeStats []kubeletVolumeStats `json:"volume"`
}

type kubeletPodRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type kubeletVolumeStats struct {
	UsedBytes     int64         `json:"usedBytes"`
	CapacityBytes int64         `json:"capacityBytes"`
	PVCRef        kubeletPVCRef `json:"pvcRef"`
}

type kubeletPVCRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// mapStatus converts ArgoCD health/sync status to proto DeploymentStatus.
func mapStatus(status *argocd.AppStatus) (deployer.DeploymentStatus, string) {
	if status == nil {
		return deployer.DeploymentStatus_DEPLOYMENT_STATUS_UNKNOWN, "no status available"
	}

	switch status.Health.Status {
	case "Healthy":
		if status.Sync.Status == "Synced" {
			return deployer.DeploymentStatus_DEPLOYMENT_STATUS_SYNCED, "all resources synced and healthy"
		}
		return deployer.DeploymentStatus_DEPLOYMENT_STATUS_OUT_OF_SYNC, "healthy but out of sync"
	case "Progressing":
		return deployer.DeploymentStatus_DEPLOYMENT_STATUS_PROGRESSING, status.Health.Message
	case "Degraded":
		return deployer.DeploymentStatus_DEPLOYMENT_STATUS_DEGRADED, status.Health.Message
	default:
		if status.Sync.Status == "OutOfSync" {
			return deployer.DeploymentStatus_DEPLOYMENT_STATUS_OUT_OF_SYNC, "out of sync"
		}
		return deployer.DeploymentStatus_DEPLOYMENT_STATUS_UNKNOWN, status.Health.Status
	}
}

func (s *Server) WorkspaceMetadata(ctx context.Context, req *deployer.WorkspaceMetadataRequest) (*deployer.WorkspaceMetadataResponse, error) {
	cmName := fmt.Sprintf("workspace-%s", req.Workspace)

	cm, err := s.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).Get(ctx, cmName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "workspace %q not found", req.Workspace)
		}
		return nil, status.Errorf(codes.Internal, "failed to get workspace ConfigMap: %v", err)
	}

	personal, _ := strconv.ParseBool(cm.Data["personal"])

	return &deployer.WorkspaceMetadataResponse{
		Name:                 cm.Data["name"],
		Personal:             personal,
		StripeCustomerId:     cm.Data["stripe_customer_id"],
		StripeSubscriptionId: cm.Data["stripe_subscription_id"],
		Owner:                cm.Data["owner"],
	}, nil
}

func (s *Server) WorkspaceByInstallationID(ctx context.Context, req *deployer.WorkspaceByInstallationIDRequest) (*deployer.WorkspaceByInstallationIDResponse, error) {
	// Query Deployments across all namespaces with the github-installation label.
	installationLabel := fmt.Sprintf("%s=%d", labels.GitHubInstallation, req.InstallationId)

	deployList, err := s.k8s.AppsV1().Deployments("").List(ctx, metav1.ListOptions{
		LabelSelector: installationLabel,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list deployments: %v", err)
	}

	if len(deployList.Items) == 0 {
		return nil, status.Errorf(codes.NotFound, "no workspace found for installation ID %d", req.InstallationId)
	}

	// Get the workspace from the namespace's labels.
	namespace := deployList.Items[0].Namespace
	ns, err := s.k8s.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get namespace %q: %v", namespace, err)
	}

	ws := ns.Labels[labels.Workspace]
	if ws == "" {
		return nil, status.Errorf(codes.Internal, "namespace %q missing workspace label", namespace)
	}

	return &deployer.WorkspaceByInstallationIDResponse{
		Workspace: ws,
	}, nil
}

func (s *Server) CreateWorkspaceMetadata(ctx context.Context, req *deployer.CreateWorkspaceMetadataRequest) (*deployer.CreateWorkspaceMetadataResponse, error) {
	cmName := fmt.Sprintf("workspace-%s", req.Workspace)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: labels.LucityNamespace,
			Labels: map[string]string{
				labels.ManagedBy:    labels.ManagedByLucity,
				labels.ResourceType: "workspace-metadata",
				labels.Workspace:    req.Workspace,
			},
		},
		Data: map[string]string{
			"name":     req.Name,
			"personal": strconv.FormatBool(req.Personal),
			"owner":    req.Owner,
		},
	}

	_, err := s.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return &deployer.CreateWorkspaceMetadataResponse{}, nil // idempotent
		}
		return nil, status.Errorf(codes.Internal, "failed to create workspace ConfigMap: %v", err)
	}
	return &deployer.CreateWorkspaceMetadataResponse{}, nil
}

func (s *Server) UpdateWorkspaceMetadata(ctx context.Context, req *deployer.UpdateWorkspaceMetadataRequest) (*deployer.UpdateWorkspaceMetadataResponse, error) {
	cmName := fmt.Sprintf("workspace-%s", req.Workspace)

	cm, err := s.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).Get(ctx, cmName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "workspace %q not found", req.Workspace)
		}
		return nil, status.Errorf(codes.Internal, "failed to get workspace ConfigMap: %v", err)
	}

	if req.Name != "" {
		cm.Data["name"] = req.Name
	}
	if req.StripeCustomerId != "" {
		cm.Data["stripe_customer_id"] = req.StripeCustomerId
	}
	if req.StripeSubscriptionId != "" {
		cm.Data["stripe_subscription_id"] = req.StripeSubscriptionId
	}

	_, err = s.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update workspace ConfigMap: %v", err)
	}
	return &deployer.UpdateWorkspaceMetadataResponse{}, nil
}

func (s *Server) DeleteWorkspaceMetadata(ctx context.Context, req *deployer.DeleteWorkspaceMetadataRequest) (*deployer.DeleteWorkspaceMetadataResponse, error) {
	cmName := fmt.Sprintf("workspace-%s", req.Workspace)

	err := s.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).Delete(ctx, cmName, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return &deployer.DeleteWorkspaceMetadataResponse{}, nil // idempotent
		}
		return nil, status.Errorf(codes.Internal, "failed to delete workspace ConfigMap: %v", err)
	}
	return &deployer.DeleteWorkspaceMetadataResponse{}, nil
}

func (s *Server) ListWorkspaces(ctx context.Context, req *deployer.ListWorkspacesRequest) (*deployer.ListWorkspacesResponse, error) {
	selector := labels.Selector(labels.ResourceType, "workspace-metadata")

	cmList, err := s.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list workspace ConfigMaps: %v", err)
	}

	workspaces := make([]*deployer.WorkspaceInfo, 0, len(cmList.Items))
	for _, cm := range cmList.Items {
		ws := cm.Labels[labels.Workspace]
		if ws == "" {
			continue
		}

		personal, _ := strconv.ParseBool(cm.Data["personal"])

		workspaces = append(workspaces, &deployer.WorkspaceInfo{
			Id:       ws,
			Name:     cm.Data["name"],
			Personal: personal,
		})
	}

	return &deployer.ListWorkspacesResponse{Workspaces: workspaces}, nil
}

const (
	resourceQuotaName = "lucity-resources"
	limitRangeName    = "lucity-defaults"
)

func (s *Server) SetResourceQuota(ctx context.Context, req *deployer.SetResourceQuotaRequest) (*deployer.SetResourceQuotaResponse, error) {
	ws := tenant.FromContext(ctx)
	namespace := labels.NamespaceFor(ws, req.Project, req.Environment)

	// 1. Create or update ResourceQuota.
	cpuQty := fmt.Sprintf("%dm", req.CpuMillicores)
	memQty := fmt.Sprintf("%dMi", req.MemoryMb)
	diskQty := fmt.Sprintf("%dMi", req.DiskMb)

	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceQuotaName,
			Namespace: namespace,
			Labels: map[string]string{
				labels.ManagedBy: labels.ManagedByLucity,
			},
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceRequestsCPU:     resource.MustParse(cpuQty),
				corev1.ResourceRequestsMemory:  resource.MustParse(memQty),
				corev1.ResourceRequestsStorage: resource.MustParse(diskQty),
			},
		},
	}

	existing, err := s.k8s.CoreV1().ResourceQuotas(namespace).Get(ctx, resourceQuotaName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, status.Errorf(codes.Internal, "failed to get resource quota: %v", err)
		}
		if _, err := s.k8s.CoreV1().ResourceQuotas(namespace).Create(ctx, quota, metav1.CreateOptions{}); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create resource quota: %v", err)
		}
	} else {
		existing.Spec.Hard = quota.Spec.Hard
		existing.Labels = quota.Labels
		if _, err := s.k8s.CoreV1().ResourceQuotas(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update resource quota: %v", err)
		}
	}

	// 2. Create or update LimitRange.
	lr := buildLimitRange(namespace, req.Tier)
	existingLR, err := s.k8s.CoreV1().LimitRanges(namespace).Get(ctx, limitRangeName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, status.Errorf(codes.Internal, "failed to get limit range: %v", err)
		}
		if _, err := s.k8s.CoreV1().LimitRanges(namespace).Create(ctx, lr, metav1.CreateOptions{}); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create limit range: %v", err)
		}
	} else {
		existingLR.Spec = lr.Spec
		existingLR.Labels = lr.Labels
		if _, err := s.k8s.CoreV1().LimitRanges(namespace).Update(ctx, existingLR, metav1.UpdateOptions{}); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update limit range: %v", err)
		}
	}

	// 3. Set resource-tier label on namespace.
	ns, err := s.k8s.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get namespace: %v", err)
	}
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels[labels.ResourceTier] = tierToString(req.Tier)
	if _, err := s.k8s.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update namespace tier label: %v", err)
	}

	slog.Info("set resource quota",
		"namespace", namespace,
		"tier", tierToString(req.Tier),
		"cpu", cpuQty, "memory", memQty, "disk", diskQty,
	)

	return &deployer.SetResourceQuotaResponse{
		Tier:          req.Tier,
		CpuMillicores: req.CpuMillicores,
		MemoryMb:      req.MemoryMb,
		DiskMb:        req.DiskMb,
	}, nil
}

func (s *Server) ResourceQuota(ctx context.Context, req *deployer.ResourceQuotaRequest) (*deployer.ResourceQuotaResponse, error) {
	ws := tenant.FromContext(ctx)
	namespace := labels.NamespaceFor(ws, req.Project, req.Environment)

	quota, err := s.k8s.CoreV1().ResourceQuotas(namespace).Get(ctx, resourceQuotaName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "no resource quota for %s", namespace)
		}
		return nil, status.Errorf(codes.Internal, "failed to get resource quota: %v", err)
	}

	ns, err := s.k8s.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get namespace: %v", err)
	}

	tier := tierFromString(ns.Labels[labels.ResourceTier])

	cpuMillis := int32(quota.Spec.Hard.Cpu().MilliValue())
	memMB := int32(quota.Spec.Hard.Memory().Value() / (1024 * 1024))

	var diskMB int32
	if storageQty, ok := quota.Spec.Hard[corev1.ResourceRequestsStorage]; ok {
		diskMB = int32(storageQty.Value() / (1024 * 1024))
	}

	return &deployer.ResourceQuotaResponse{
		Tier:          tier,
		CpuMillicores: cpuMillis,
		MemoryMb:      memMB,
		DiskMb:        diskMB,
	}, nil
}

func buildLimitRange(namespace string, tier deployer.ResourceTier) *corev1.LimitRange {
	lr := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      limitRangeName,
			Namespace: namespace,
			Labels: map[string]string{
				labels.ManagedBy: labels.ManagedByLucity,
			},
		},
	}

	if tier == deployer.ResourceTier_RESOURCE_TIER_PRODUCTION {
		// Guaranteed QoS: requests = limits.
		lr.Spec.Limits = []corev1.LimitRangeItem{{
			Type: corev1.LimitTypeContainer,
			Default: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
			DefaultRequest: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		}}
	} else {
		// Burstable QoS: requests < limits.
		lr.Spec.Limits = []corev1.LimitRangeItem{{
			Type: corev1.LimitTypeContainer,
			Default: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
			DefaultRequest: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
		}}
	}
	return lr
}

func tierToString(t deployer.ResourceTier) string {
	switch t {
	case deployer.ResourceTier_RESOURCE_TIER_ECO:
		return "eco"
	case deployer.ResourceTier_RESOURCE_TIER_PRODUCTION:
		return "production"
	default:
		return "eco"
	}
}

func tierFromString(s string) deployer.ResourceTier {
	switch s {
	case "production":
		return deployer.ResourceTier_RESOURCE_TIER_PRODUCTION
	default:
		return deployer.ResourceTier_RESOURCE_TIER_ECO
	}
}

func (s *Server) ListResourceAllocations(ctx context.Context, req *deployer.ListResourceAllocationsRequest) (*deployer.ListResourceAllocationsResponse, error) {
	// List all namespaces with resource-tier label (managed by Lucity with quotas set).
	selector := labels.Selector(labels.ManagedBy, labels.ManagedByLucity) + "," + labels.ResourceTier

	nsList, err := s.k8s.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list namespaces: %v", err)
	}

	var allocations []*deployer.ResourceAllocation
	for _, ns := range nsList.Items {
		ws := ns.Labels[labels.Workspace]
		project := ns.Labels[labels.Project]
		env := ns.Labels[labels.Environment]
		if ws == "" || project == "" || env == "" {
			continue
		}

		tier := tierFromString(ns.Labels[labels.ResourceTier])

		quota, err := s.k8s.CoreV1().ResourceQuotas(ns.Name).Get(ctx, resourceQuotaName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue // No quota set — skip
			}
			slog.Warn("failed to get resource quota", "namespace", ns.Name, "error", err)
			continue
		}

		cpuMillis := int32(quota.Spec.Hard.Cpu().MilliValue())
		memMB := int32(quota.Spec.Hard.Memory().Value() / (1024 * 1024))

		var diskMB int32
		if storageQty, ok := quota.Spec.Hard[corev1.ResourceRequestsStorage]; ok {
			diskMB = int32(storageQty.Value() / (1024 * 1024))
		}

		allocations = append(allocations, &deployer.ResourceAllocation{
			Workspace:     ws,
			Project:       project,
			Environment:   env,
			Tier:          tier,
			CpuMillicores: cpuMillis,
			MemoryMb:      memMB,
			DiskMb:        diskMB,
		})
	}

	return &deployer.ListResourceAllocationsResponse{Allocations: allocations}, nil
}

func (s *Server) DatabaseCredentials(ctx context.Context, req *deployer.DatabaseCredentialsRequest) (*deployer.DatabaseCredentialsResponse, error) {
	creds, err := database.CredentialsFromSecret(ctx, s.k8s, tenant.FromContext(ctx), req.Project, req.Environment, req.Database)
	if err != nil {
		if errors.Is(err, database.ErrNotReady) {
			return nil, status.Errorf(codes.FailedPrecondition, "database is provisioning")
		}
		return nil, status.Errorf(codes.Internal, "failed to read credentials: %v", err)
	}

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", creds.User, creds.Password, creds.Host, creds.Port, creds.DBName)

	return &deployer.DatabaseCredentialsResponse{
		Host:     creds.Host,
		Port:     creds.Port,
		Dbname:   creds.DBName,
		User:     creds.User,
		Password: creds.Password,
		Uri:      uri,
	}, nil
}

// --- User GitHub token CRUD (K8s Secrets) ---

func githubTokenSecretName(userID string) string {
	return "github-token-" + strings.ToLower(userID)
}

func (s *Server) StoreUserGitHubToken(ctx context.Context, req *deployer.StoreUserGitHubTokenRequest) (*deployer.StoreUserGitHubTokenResponse, error) {
	secretName := githubTokenSecretName(req.UserId)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: labels.LucityNamespace,
			Labels: map[string]string{
				labels.ManagedBy:    labels.ManagedByLucity,
				labels.ResourceType: "user-github-token",
			},
		},
		StringData: map[string]string{
			"access_token":  req.AccessToken,
			"refresh_token": req.RefreshToken,
			"expires_at":    strconv.FormatInt(req.ExpiresAt, 10),
		},
	}

	existing, err := s.k8s.CoreV1().Secrets(labels.LucityNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, status.Errorf(codes.Internal, "failed to get token secret: %v", err)
		}
		if _, err := s.k8s.CoreV1().Secrets(labels.LucityNamespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create token secret: %v", err)
		}
	} else {
		existing.Data = nil // clear binary data, use StringData
		existing.StringData = secret.StringData
		existing.Labels = secret.Labels
		if _, err := s.k8s.CoreV1().Secrets(labels.LucityNamespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update token secret: %v", err)
		}
	}

	slog.Info("stored GitHub token", "user", req.UserId)
	return &deployer.StoreUserGitHubTokenResponse{}, nil
}

func (s *Server) UserGitHubToken(ctx context.Context, req *deployer.UserGitHubTokenRequest) (*deployer.UserGitHubTokenResponse, error) {
	secretName := githubTokenSecretName(req.UserId)

	secret, err := s.k8s.CoreV1().Secrets(labels.LucityNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return &deployer.UserGitHubTokenResponse{Connected: false}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to get token secret: %v", err)
	}

	var expiresAt int64
	if raw := string(secret.Data["expires_at"]); raw != "" {
		expiresAt, _ = strconv.ParseInt(raw, 10, 64)
	}

	return &deployer.UserGitHubTokenResponse{
		Connected:    true,
		AccessToken:  string(secret.Data["access_token"]),
		RefreshToken: string(secret.Data["refresh_token"]),
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *Server) DeleteUserGitHubToken(ctx context.Context, req *deployer.DeleteUserGitHubTokenRequest) (*deployer.DeleteUserGitHubTokenResponse, error) {
	secretName := githubTokenSecretName(req.UserId)

	if err := s.k8s.CoreV1().Secrets(labels.LucityNamespace).Delete(ctx, secretName, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return &deployer.DeleteUserGitHubTokenResponse{}, nil // idempotent
		}
		return nil, status.Errorf(codes.Internal, "failed to delete token secret: %v", err)
	}

	slog.Info("deleted GitHub token", "user", req.UserId)
	return &deployer.DeleteUserGitHubTokenResponse{}, nil
}
