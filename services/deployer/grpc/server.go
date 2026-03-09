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
		req.Service, req.Project)

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

	var installationID int64
	if raw := cm.Data["github_installation_id"]; raw != "" {
		installationID, err = strconv.ParseInt(raw, 10, 64)
		if err != nil {
			slog.Warn("failed to parse github_installation_id", "workspace", req.Workspace, "value", raw, "error", err)
		}
	}

	return &deployer.WorkspaceMetadataResponse{
		Name:                 cm.Data["name"],
		Personal:             personal,
		GithubInstallationId: installationID,
	}, nil
}

func (s *Server) WorkspaceByInstallationID(ctx context.Context, req *deployer.WorkspaceByInstallationIDRequest) (*deployer.WorkspaceByInstallationIDResponse, error) {
	selector := labels.Selector(labels.ResourceType, "workspace-metadata")

	cmList, err := s.k8s.CoreV1().ConfigMaps(labels.LucityNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list workspace ConfigMaps: %v", err)
	}

	targetID := strconv.FormatInt(req.InstallationId, 10)
	for _, cm := range cmList.Items {
		if cm.Data["github_installation_id"] == targetID {
			ws := cm.Labels[labels.Workspace]
			if ws == "" {
				return nil, status.Errorf(codes.Internal, "workspace ConfigMap %q missing workspace label", cm.Name)
			}
			return &deployer.WorkspaceByInstallationIDResponse{
				Workspace: ws,
			}, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "no workspace found for installation ID %d", req.InstallationId)
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
