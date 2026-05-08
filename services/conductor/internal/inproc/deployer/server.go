package deployer

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

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/data"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/argo/argocd"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/argo/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PackagerService is the narrow slice of the inproc packager that the
// deployer calls back into. Defined here to break the import cycle
// between inproc/deployer and inproc/packager.
type PackagerService interface {
	SetSuspended(ctx context.Context, workspace, project, environment string, suspended bool) error
}

// ServiceLogEntry is a single log line from a workload pod.
type ServiceLogEntry struct {
	Line string
	Pod  string
}

type Server struct {
	argo     *argocd.Client
	packager PackagerService
	k8s      kubernetes.Interface
	dynamic  dynamic.Interface

	// softServeHTTP is the cluster-internal Soft-serve HTTP URL for ArgoCD to clone from.
	softServeHTTP string
	// softServeToken is the Soft-serve access token for HTTP git operations.
	softServeToken string

	// Gateway API and cert-manager configuration for custom domains.
	gatewayName      string
	gatewayNamespace string
	clusterIssuer    string

	// registryPullSecret is the name of the source dockerconfigjson Secret in
	// the platform namespace (lucity-system). The deployer clones this Secret
	// into each workload namespace so kubelet can authenticate image pulls.
	registryPullSecret string
}

func NewServer(argo *argocd.Client, packagerSvc PackagerService, softServeHTTP, softServeToken string, k8s kubernetes.Interface, dyn dynamic.Interface, gatewayName, gatewayNamespace, clusterIssuer, registryPullSecret string) *Server {
	return &Server{
		argo:               argo,
		packager:           packagerSvc,
		k8s:                k8s,
		dynamic:            dyn,
		softServeHTTP:      softServeHTTP,
		softServeToken:     softServeToken,
		gatewayName:        gatewayName,
		gatewayNamespace:   gatewayNamespace,
		clusterIssuer:      clusterIssuer,
		registryPullSecret: registryPullSecret,
	}
}

// SetPackager wires the packager after construction. Used to break
// the deployer↔packager cycle. See packager.Server.SetDeployer for
// the symmetric setter.
func (s *Server) SetPackager(p PackagerService) {
	s.packager = p
}

// DeployEnvironment provisions an ArgoCD Application + namespace for an environment.
// Idempotent: returns the existing app if already created.
func (s *Server) DeployEnvironment(ctx context.Context, workspace, project, environment, gitopsRepoURL, targetNamespace string) (deploymentName string, err error) {
	appName := applicationName(workspace, project, environment)

	// Idempotent: if the application already exists, return it.
	existing, err := s.argo.Application(ctx, appName)
	if err == nil && existing != nil {
		slog.Info("ArgoCD application already exists",
			"app", appName,
			"project", project,
			"environment", environment,
		)
		return existing.Metadata.Name, nil
	}

	// Ensure the namespace exists with platform labels.
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: targetNamespace,
			Labels: map[string]string{
				labels.Workspace:   workspace,
				labels.Project:     project,
				labels.Environment: environment,
				labels.ManagedBy:   labels.ManagedByLucity,
			},
		},
	}
	if _, err := s.k8s.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		return "", fmt.Errorf("failed to create namespace %s: %w", targetNamespace, err)
	}

	// Set up default ECO tier: LimitRange for per-pod defaults + namespace label.
	s.ensureDefaultEcoTier(ctx, targetNamespace)

	// Clone registry pull credentials so kubelet can authenticate image pulls.
	s.ensureRegistryPullSecret(ctx, targetNamespace)

	// Ensure the GitOps repo is registered in ArgoCD with credentials.
	repoURL := s.repoURL(workspace, project)
	if err := s.argo.CreateRepository(ctx, argocd.Repository{
		Repo:     repoURL,
		Username: "lucity",
		Password: s.softServeToken,
		Type:     "git",
	}); err != nil {
		return "", fmt.Errorf("failed to register repository in ArgoCD: %w", err)
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
						fmt.Sprintf("../environments/%s/values.yaml", environment),
					},
				},
			},
			Destination: argocd.Destination{
				Server:    "https://kubernetes.default.svc",
				Namespace: targetNamespace,
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
		return "", fmt.Errorf("failed to create ArgoCD application: %w", err)
	}

	slog.Info("created ArgoCD application",
		"app", result.Metadata.Name,
		"project", project,
		"environment", environment,
		"namespace", targetNamespace,
	)
	return result.Metadata.Name, nil
}

// RemoveDeployment removes the ArgoCD Application and namespace for an environment.
func (s *Server) RemoveDeployment(ctx context.Context, workspace, project, environment string) error {
	appName := applicationName(workspace, project, environment)

	if err := s.argo.DeleteApplication(ctx, appName, true); err != nil {
		// Idempotent: if the application is already gone, that's fine — still clean up the namespace.
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to delete ArgoCD application: %w", err)
		}
		slog.Info("ArgoCD application already deleted", "app", appName)
	} else {
		slog.Info("deleted ArgoCD application", "app", appName)
	}

	// Delete the namespace. ArgoCD cascade already cleaned up resources inside.
	ns := labels.NamespaceFor(workspace, project, environment)
	if err := s.k8s.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete namespace %s: %w", ns, err)
	}
	slog.Info("deleted namespace", "namespace", ns)
	return nil
}

// DeleteRepository removes the ArgoCD repository credential for a project.
func (s *Server) DeleteRepository(ctx context.Context, workspace, project string) error {
	repoURL := s.repoURL(workspace, project)

	if err := s.argo.DeleteRepository(ctx, repoURL); err != nil {
		return fmt.Errorf("failed to delete ArgoCD repository: %w", err)
	}

	slog.Info("deleted ArgoCD repository", "repo", repoURL)
	return nil
}

func (s *Server) GetDeploymentStatus(ctx context.Context, workspace, project, environment string) (data.DeploymentStatus, string, error) {
	appName := applicationName(workspace, project, environment)

	app, err := s.argo.Application(ctx, appName)
	if err != nil {
		return data.DeploymentStatusUnspecified, "", fmt.Errorf("failed to get ArgoCD application: %w", err)
	}

	status, message := mapStatus(app.Status)
	return status, message, nil
}

func (s *Server) SyncDeployment(ctx context.Context, workspace, project, environment string) (data.DeploymentStatus, error) {
	appName := applicationName(workspace, project, environment)

	app, err := s.argo.SyncApplication(ctx, appName)
	if err != nil {
		return data.DeploymentStatusUnspecified, fmt.Errorf("failed to sync ArgoCD application: %w", err)
	}

	slog.Info("triggered sync", "app", appName)

	status, _ := mapStatus(app.Status)
	return status, nil
}

// applicationName derives the ArgoCD Application name from workspace, project, and environment.
func applicationName(workspace, project, environment string) string {
	return labels.NamespaceFor(workspace, project, environment)
}

// repoURL returns the Soft-serve HTTP clone URL for a project's GitOps repo.
func (s *Server) repoURL(workspace, project string) string {
	return strings.TrimSuffix(s.softServeHTTP, "/") + "/" + workspace + "-" + project + "-gitops.git"
}

// ServiceLogs returns a channel of log entries from the workload's
// pods. The channel is closed when all pod log streams end or ctx is
// cancelled.
func (s *Server) ServiceLogs(ctx context.Context, workspace, project, environment, service string, tailLines int) (<-chan ServiceLogEntry, error) {
	namespace := labels.NamespaceFor(workspace, project, environment)

	labelSelector := fmt.Sprintf("app.kubernetes.io/name=%s", service)

	podList, err := s.k8s.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list pods: %v", err)
	}
	if len(podList.Items) == 0 {
		return nil, status.Errorf(codes.NotFound, "no pods found for service %q in %q", service, namespace)
	}

	multiplePods := len(podList.Items) > 1

	tail := int64(1000)
	if tailLines > 0 {
		tail = int64(tailLines)
	}

	out := make(chan ServiceLogEntry, 128)
	var wg sync.WaitGroup
	for _, pod := range podList.Items {
		wg.Add(1)
		go func(podName string) {
			defer wg.Done()
			s.streamPodLogs(ctx, namespace, podName, service, tail, multiplePods, out)
		}(pod.Name)
	}
	go func() {
		wg.Wait()
		close(out)
	}()

	return out, nil
}

// streamPodLogs follows logs from a single pod/container and sends entries to the channel.
func (s *Server) streamPodLogs(ctx context.Context, namespace, podName, container string, tailLines int64, prefixPod bool, out chan<- ServiceLogEntry) {
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
		case out <- ServiceLogEntry{Line: line, Pod: podName}:
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

func (s *Server) DatabaseTables(ctx context.Context, workspace, project, environment, dbName string) ([]data.DatabaseTable, error) {
	creds, err := database.CredentialsFromSecret(ctx, s.k8s, workspace, project, environment, dbName)
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
		slog.Error("failed to connect to database", "project", project, "environment", environment, "database", dbName, "error", err)
		return nil, status.Errorf(codes.Unavailable, "database connection failed")
	}
	defer conn.Close(queryCtx)

	tables, err := database.Tables(queryCtx, conn)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tables: %v", err)
	}
	return tables, nil
}

func (s *Server) DatabaseTableData(ctx context.Context, workspace, project, environment, dbName, schema, table string, limit, offset int) (data.DatabaseTableData, error) {
	creds, err := database.CredentialsFromSecret(ctx, s.k8s, workspace, project, environment, dbName)
	if err != nil {
		if errors.Is(err, database.ErrNotReady) {
			return data.DatabaseTableData{}, status.Errorf(codes.FailedPrecondition, "database is provisioning")
		}
		return data.DatabaseTableData{}, status.Errorf(codes.NotFound, "database credentials not found: %v", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()

	conn, err := database.Connect(queryCtx, creds)
	if err != nil {
		slog.Error("failed to connect to database", "project", project, "environment", environment, "database", dbName, "error", err)
		return data.DatabaseTableData{}, status.Errorf(codes.Unavailable, "database connection failed")
	}
	defer conn.Close(queryCtx)

	columns, rows, estimatedRows, err := database.TableData(queryCtx, conn, schema, table, limit, offset)
	if err != nil {
		return data.DatabaseTableData{}, status.Errorf(codes.Internal, "failed to query table data: %v", err)
	}
	return data.DatabaseTableData{
		Columns:            columns,
		Rows:               rows,
		TotalEstimatedRows: estimatedRows,
	}, nil
}

func (s *Server) DatabaseQuery(ctx context.Context, workspace, project, environment, dbName, query string) (data.DatabaseQueryResult, error) {
	creds, err := database.CredentialsFromSecret(ctx, s.k8s, workspace, project, environment, dbName)
	if err != nil {
		if errors.Is(err, database.ErrNotReady) {
			return data.DatabaseQueryResult{}, status.Errorf(codes.FailedPrecondition, "database is provisioning")
		}
		return data.DatabaseQueryResult{}, status.Errorf(codes.NotFound, "database credentials not found: %v", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()

	conn, err := database.Connect(queryCtx, creds)
	if err != nil {
		slog.Error("failed to connect to database", "project", project, "environment", environment, "database", dbName, "error", err)
		return data.DatabaseQueryResult{}, status.Errorf(codes.Unavailable, "database connection failed")
	}
	defer conn.Close(queryCtx)

	columns, rows, affected, err := database.Query(queryCtx, conn, query)
	if err != nil {
		return data.DatabaseQueryResult{}, status.Errorf(codes.Internal, "query failed: %v", err)
	}
	return data.DatabaseQueryResult{
		Columns:      columns,
		Rows:         rows,
		AffectedRows: affected,
	}, nil
}

func (s *Server) DatabaseStatus(ctx context.Context, workspace, project, environment, dbName string) (data.DatabaseStatus, error) {
	namespace := labels.NamespaceFor(workspace, project, environment)
	clusterName := project + "-pg-" + dbName
	labelSelector := fmt.Sprintf("cnpg.io/cluster=%s", clusterName)

	// Check CNPG pods for readiness.
	podList, err := s.k8s.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return data.DatabaseStatus{}, status.Errorf(codes.Internal, "failed to list database pods: %v", err)
	}

	var runningInstances int
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
	var volumeInfo *data.VolumeInfo
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
		volumeInfo = &data.VolumeInfo{
			Name:          pvc.Name,
			Size:          capacity,
			RequestedSize: requested,
		}

		// Try to get actual disk usage from kubelet stats.
		usedBytes, capacityBytes := s.pvcUsage(ctx, namespace, pvc.Name, podList.Items)
		volumeInfo.UsedBytes = usedBytes
		volumeInfo.CapacityBytes = capacityBytes
	}

	return data.DatabaseStatus{
		Ready:     runningInstances > 0,
		Instances: runningInstances,
		Volume:    volumeInfo,
	}, nil
}

func (s *Server) ServiceStatus(ctx context.Context, workspace, project, environment, service string) (data.ServiceStatus, error) {
	namespace := labels.NamespaceFor(workspace, project, environment)
	labelSelector := fmt.Sprintf("app.kubernetes.io/name=%s", service)

	deployList, err := s.k8s.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return data.ServiceStatus{}, status.Errorf(codes.Internal, "failed to list deployments: %v", err)
	}

	var totalReplicas, totalReady, configuredReplicas int32
	for _, d := range deployList.Items {
		totalReplicas += d.Status.Replicas
		totalReady += d.Status.ReadyReplicas
		if d.Spec.Replicas != nil {
			configuredReplicas = *d.Spec.Replicas
		}
	}

	// Check for an HPA targeting this deployment.
	scaling := &data.ServiceScaling{
		Replicas: int(configuredReplicas),
	}

	hpaList, err := s.k8s.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err == nil && len(hpaList.Items) > 0 {
		hpa := hpaList.Items[0]
		scaling.AutoscalingEnabled = true
		scaling.MinReplicas = 1
		if hpa.Spec.MinReplicas != nil {
			scaling.MinReplicas = int(*hpa.Spec.MinReplicas)
		}
		scaling.MaxReplicas = int(hpa.Spec.MaxReplicas)
		for _, metric := range hpa.Spec.Metrics {
			if metric.Type == autoscalingv2.ContainerResourceMetricSourceType && metric.ContainerResource != nil {
				if metric.ContainerResource.Target.AverageUtilization != nil {
					scaling.TargetCPU = int(*metric.ContainerResource.Target.AverageUtilization)
				}
			}
		}
	}

	// Extract container resource requests/limits from a running Pod.
	// We read from Pods rather than the Deployment spec because LimitRange defaults
	// are injected at pod admission time and don't appear in the Deployment spec.
	var resources *data.ServiceResources
	podList, podErr := s.k8s.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
		Limit:         1,
	})
	if podErr == nil && len(podList.Items) > 0 {
		containers := podList.Items[0].Spec.Containers
		if len(containers) > 0 {
			ctr := containers[0]
			res := &data.ServiceResources{}
			if req, ok := ctr.Resources.Requests[corev1.ResourceCPU]; ok {
				res.CPUMillicores = int(req.MilliValue())
			}
			if req, ok := ctr.Resources.Requests[corev1.ResourceMemory]; ok {
				res.MemoryMB = int(req.Value() / (1024 * 1024))
			}
			if lim, ok := ctr.Resources.Limits[corev1.ResourceCPU]; ok {
				res.CPULimitMillicores = int(lim.MilliValue())
			}
			if lim, ok := ctr.Resources.Limits[corev1.ResourceMemory]; ok {
				res.MemoryLimitMB = int(lim.Value() / (1024 * 1024))
			}
			// Only return resources if at least one value is set.
			if res.CPUMillicores > 0 || res.MemoryMB > 0 || res.CPULimitMillicores > 0 || res.MemoryLimitMB > 0 {
				resources = res
			}
		}
	}

	return data.ServiceStatus{
		Ready:         totalReady > 0 && totalReady >= totalReplicas,
		Replicas:      int(totalReplicas),
		ReadyReplicas: int(totalReady),
		Scaling:       scaling,
		Resources:     resources,
	}, nil
}

func (s *Server) SetServiceScaling(ctx context.Context, workspace, project, environment, service string, replicas int, autoscaling *data.AutoscalingConfig) error {
	namespace := labels.NamespaceFor(workspace, project, environment)
	deploymentName := fmt.Sprintf("%s-%s-%s", project, environment, service)

	// Find the deployment by label first (more reliable than guessing the name).
	labelSelector := fmt.Sprintf("app.kubernetes.io/name=%s", service)
	deployList, err := s.k8s.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to list deployments: %v", err)
	}
	if len(deployList.Items) == 0 {
		return status.Errorf(codes.NotFound, "no deployment found for service %q in %q", service, namespace)
	}

	dep := &deployList.Items[0]
	deploymentName = dep.Name
	hpaName := deploymentName

	if autoscaling != nil && autoscaling.Enabled {
		// Autoscaling enabled: create/update HPA, remove replicas from Deployment.
		minReplicas := int32(autoscaling.MinReplicas)
		if minReplicas < 1 {
			minReplicas = 1
		}
		maxReplicas := int32(autoscaling.MaxReplicas)
		if maxReplicas < minReplicas {
			maxReplicas = minReplicas
		}
		targetCPU := int32(autoscaling.TargetCPU)
		if targetCPU <= 0 {
			targetCPU = 70
		}

		hpa := &autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      hpaName,
				Namespace: namespace,
				Labels: map[string]string{
					"app.kubernetes.io/name": service,
					labels.ManagedBy:         labels.ManagedByLucity,
				},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       deploymentName,
				},
				MinReplicas: &minReplicas,
				MaxReplicas: maxReplicas,
				Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
					ScaleUp: &autoscalingv2.HPAScalingRules{
						StabilizationWindowSeconds: int32Ptr(30),
					},
					ScaleDown: &autoscalingv2.HPAScalingRules{
						StabilizationWindowSeconds: int32Ptr(300),
						Policies: []autoscalingv2.HPAScalingPolicy{
							{
								Type:          autoscalingv2.PodsScalingPolicy,
								Value:         1,
								PeriodSeconds: 60,
							},
						},
					},
				},
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ContainerResourceMetricSourceType,
						ContainerResource: &autoscalingv2.ContainerResourceMetricSource{
							Name:      corev1.ResourceCPU,
							Container: service,
							Target: autoscalingv2.MetricTarget{
								Type:               autoscalingv2.UtilizationMetricType,
								AverageUtilization: &targetCPU,
							},
						},
					},
				},
			},
		}

		existing, err := s.k8s.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, hpaName, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return status.Errorf(codes.Internal, "failed to get HPA: %v", err)
			}
			if _, err := s.k8s.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(ctx, hpa, metav1.CreateOptions{}); err != nil {
				return status.Errorf(codes.Internal, "failed to create HPA: %v", err)
			}
		} else {
			existing.Spec = hpa.Spec
			existing.Labels = hpa.Labels
			if _, err := s.k8s.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
				return status.Errorf(codes.Internal, "failed to update HPA: %v", err)
			}
		}

		// Remove replicas from Deployment — HPA owns it now.
		dep.Spec.Replicas = nil
		if _, err := s.k8s.AppsV1().Deployments(namespace).Update(ctx, dep, metav1.UpdateOptions{}); err != nil {
			slog.Warn("failed to clear deployment replicas for HPA", "deployment", deploymentName, "error", err)
		}

		slog.Info("set autoscaling", "deployment", deploymentName, "min", minReplicas, "max", maxReplicas, "targetCPU", targetCPU)
	} else {
		// Manual scaling: set replicas on Deployment, delete HPA if it exists.
		r := int32(replicas)
		if r < 1 {
			r = 1
		}
		dep.Spec.Replicas = &r
		if _, err := s.k8s.AppsV1().Deployments(namespace).Update(ctx, dep, metav1.UpdateOptions{}); err != nil {
			return status.Errorf(codes.Internal, "failed to scale deployment: %v", err)
		}

		// Delete HPA if it exists (idempotent).
		if err := s.k8s.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, hpaName, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			slog.Warn("failed to delete HPA", "hpa", hpaName, "error", err)
		}

		slog.Info("set manual scaling", "deployment", deploymentName, "replicas", r)
	}
	return nil
}

func int32Ptr(v int32) *int32 { return &v }

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

// mapStatus converts ArgoCD health/sync status to data.DeploymentStatus.
func mapStatus(status *argocd.AppStatus) (data.DeploymentStatus, string) {
	if status == nil {
		return data.DeploymentStatusUnknown, "no status available"
	}

	switch status.Health.Status {
	case "Healthy":
		if status.Sync.Status == "Synced" {
			return data.DeploymentStatusSynced, "all resources synced and healthy"
		}
		return data.DeploymentStatusOutOfSync, "healthy but out of sync"
	case "Progressing":
		return data.DeploymentStatusProgressing, status.Health.Message
	case "Degraded":
		return data.DeploymentStatusDegraded, status.Health.Message
	default:
		if status.Sync.Status == "OutOfSync" {
			return data.DeploymentStatusOutOfSync, "out of sync"
		}
		return data.DeploymentStatusUnknown, status.Health.Status
	}
}

// WorkspaceByInstallationID returns the workspace that owns a GitHub App
// installation, by inspecting deployment labels across the cluster.
func (s *Server) WorkspaceByInstallationID(ctx context.Context, installationID int64) (string, error) {
	// Query Deployments across all namespaces with the github-installation label.
	installationLabel := fmt.Sprintf("%s=%d", labels.GitHubInstallation, installationID)

	deployList, err := s.k8s.AppsV1().Deployments("").List(ctx, metav1.ListOptions{
		LabelSelector: installationLabel,
	})
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to list deployments: %v", err)
	}

	if len(deployList.Items) == 0 {
		return "", status.Errorf(codes.NotFound, "no workspace found for installation ID %d", installationID)
	}

	// Get the workspace from the namespace's labels.
	namespace := deployList.Items[0].Namespace
	ns, err := s.k8s.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to get namespace %q: %v", namespace, err)
	}

	ws := ns.Labels[labels.Workspace]
	if ws == "" {
		return "", status.Errorf(codes.Internal, "namespace %q missing workspace label", namespace)
	}
	return ws, nil
}

const (
	resourceQuotaName = "lucity-resources"
	limitRangeName    = "lucity-defaults"
)

// SetResourceQuota creates or updates the ResourceQuota + LimitRange
// for an environment namespace based on tier.
func (s *Server) SetResourceQuota(ctx context.Context, workspace, project, environment string, tier data.ResourceTier, cpuMillicores, memoryMB, diskMB int) (data.ResourceQuota, error) {
	namespace := labels.NamespaceFor(workspace, project, environment)

	// 1. Manage ResourceQuota based on tier.
	if tier == data.ResourceTierProduction {
		// PRODUCTION: create or update ResourceQuota with reserved allocations.
		cpuQty := fmt.Sprintf("%dm", cpuMillicores)
		memQty := fmt.Sprintf("%dMi", memoryMB)
		diskQty := fmt.Sprintf("%dMi", diskMB)

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
				return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to get resource quota: %v", err)
			}
			if _, err := s.k8s.CoreV1().ResourceQuotas(namespace).Create(ctx, quota, metav1.CreateOptions{}); err != nil {
				return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to create resource quota: %v", err)
			}
		} else {
			existing.Spec.Hard = quota.Spec.Hard
			existing.Labels = quota.Labels
			if _, err := s.k8s.CoreV1().ResourceQuotas(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
				return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to update resource quota: %v", err)
			}
		}
	} else {
		// ECO: no ResourceQuota — metered billing, no namespace-level cap.
		err := s.k8s.CoreV1().ResourceQuotas(namespace).Delete(ctx, resourceQuotaName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to delete resource quota: %v", err)
		}
	}

	// 2. Create or update LimitRange.
	lr := buildLimitRange(namespace, tier)
	existingLR, err := s.k8s.CoreV1().LimitRanges(namespace).Get(ctx, limitRangeName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to get limit range: %v", err)
		}
		if _, err := s.k8s.CoreV1().LimitRanges(namespace).Create(ctx, lr, metav1.CreateOptions{}); err != nil {
			return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to create limit range: %v", err)
		}
	} else {
		existingLR.Spec = lr.Spec
		existingLR.Labels = lr.Labels
		if _, err := s.k8s.CoreV1().LimitRanges(namespace).Update(ctx, existingLR, metav1.UpdateOptions{}); err != nil {
			return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to update limit range: %v", err)
		}
	}

	// 3. Set resource-tier label on namespace.
	ns, err := s.k8s.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to get namespace: %v", err)
	}
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels[labels.ResourceTier] = tierToString(tier)
	if _, err := s.k8s.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{}); err != nil {
		return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to update namespace tier label: %v", err)
	}

	slog.Info("set resource quota", "namespace", namespace, "tier", tierToString(tier), "cpu", cpuMillicores, "memory", memoryMB, "disk", diskMB)
	return data.ResourceQuota{
		Tier:          tier,
		CPUMillicores: cpuMillicores,
		MemoryMB:      memoryMB,
		DiskMB:        diskMB,
	}, nil
}

// ResourceQuota returns the current resource quota and tier for an environment.
func (s *Server) ResourceQuota(ctx context.Context, workspace, project, environment string) (data.ResourceQuota, error) {
	namespace := labels.NamespaceFor(workspace, project, environment)

	ns, err := s.k8s.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to get namespace: %v", err)
	}

	tier := tierFromString(ns.Labels[labels.ResourceTier])

	// ECO tier has no ResourceQuota — return tier with zero allocations.
	quota, err := s.k8s.CoreV1().ResourceQuotas(namespace).Get(ctx, resourceQuotaName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return data.ResourceQuota{Tier: tier}, nil
		}
		return data.ResourceQuota{}, status.Errorf(codes.Internal, "failed to get resource quota: %v", err)
	}

	cpuMillis := int(quota.Spec.Hard.Cpu().MilliValue())
	memMB := int(quota.Spec.Hard.Memory().Value() / (1024 * 1024))

	var diskMB int
	if storageQty, ok := quota.Spec.Hard[corev1.ResourceRequestsStorage]; ok {
		diskMB = int(storageQty.Value() / (1024 * 1024))
	}

	return data.ResourceQuota{
		Tier:          tier,
		CPUMillicores: cpuMillis,
		MemoryMB:      memMB,
		DiskMB:        diskMB,
	}, nil
}

// ensureDefaultEcoTier sets up the default ECO tier for a new environment: LimitRange for
// per-pod defaults and namespace tier label. No ResourceQuota — ECO uses metered billing
// based on actual usage, so there's no namespace-level resource cap.
func (s *Server) ensureDefaultEcoTier(ctx context.Context, namespace string) {
	// Skip if already configured (idempotent).
	ns, err := s.k8s.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		slog.Warn("failed to get namespace for tier setup", "namespace", namespace, "error", err)
		return
	}
	if ns.Labels[labels.ResourceTier] != "" {
		return
	}

	tier := data.ResourceTierEco

	// Create LimitRange for per-pod defaults.
	lr := buildLimitRange(namespace, tier)
	if _, err := s.k8s.CoreV1().LimitRanges(namespace).Create(ctx, lr, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		slog.Warn("failed to create default limit range", "namespace", namespace, "error", err)
	}

	// Set resource-tier label on namespace.
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels[labels.ResourceTier] = tierToString(tier)
	if _, err := s.k8s.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{}); err != nil {
		slog.Warn("failed to set default resource tier label", "namespace", namespace, "error", err)
	}

	slog.Info("set default ECO tier", "namespace", namespace)
}

// registryPullSecretName is the well-known name for the registry pull Secret
// in workload namespaces. Referenced by the lucity-app chart's imagePullSecrets.
const registryPullSecretName = "lucity-registry"

// ensureRegistryPullSecret clones the platform's registry pull Secret into a
// workload namespace so kubelet can authenticate when pulling images.
// Best-effort: logs a warning and returns if the source Secret is missing
// (e.g., in dev environments without a private registry).
func (s *Server) ensureRegistryPullSecret(ctx context.Context, namespace string) {
	if s.registryPullSecret == "" {
		return
	}

	// Read the source Secret from the platform namespace.
	source, err := s.k8s.CoreV1().Secrets(s.gatewayNamespace).Get(ctx, s.registryPullSecret, metav1.GetOptions{})
	if err != nil {
		slog.Warn("registry pull secret not found, skipping", "secret", s.registryPullSecret, "namespace", s.gatewayNamespace, "error", err)
		return
	}

	target := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryPullSecretName,
			Namespace: namespace,
			Labels: map[string]string{
				labels.ManagedBy: labels.ManagedByLucity,
			},
		},
		Type: source.Type,
		Data: source.Data,
	}

	existing, err := s.k8s.CoreV1().Secrets(namespace).Get(ctx, registryPullSecretName, metav1.GetOptions{})
	if err != nil {
		// Create new.
		if _, err := s.k8s.CoreV1().Secrets(namespace).Create(ctx, target, metav1.CreateOptions{}); err != nil {
			slog.Warn("failed to create registry pull secret", "namespace", namespace, "error", err)
			return
		}
	} else {
		// Update existing.
		existing.Type = source.Type
		existing.Data = source.Data
		if _, err := s.k8s.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
			slog.Warn("failed to update registry pull secret", "namespace", namespace, "error", err)
			return
		}
	}

	slog.Info("ensured registry pull secret", "namespace", namespace)
}

func buildLimitRange(namespace string, tier data.ResourceTier) *corev1.LimitRange {
	lr := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      limitRangeName,
			Namespace: namespace,
			Labels: map[string]string{
				labels.ManagedBy: labels.ManagedByLucity,
			},
		},
	}

	if tier == data.ResourceTierProduction {
		// Guaranteed QoS: requests = limits.
		lr.Spec.Limits = []corev1.LimitRangeItem{{
			Type: corev1.LimitTypeContainer,
			Default: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
			DefaultRequest: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
		}}
	} else {
		// Burstable QoS: requests < limits.
		lr.Spec.Limits = []corev1.LimitRangeItem{{
			Type: corev1.LimitTypeContainer,
			Default: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
			DefaultRequest: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		}}
	}
	return lr
}

func tierToString(t data.ResourceTier) string {
	switch t {
	case data.ResourceTierEco:
		return "eco"
	case data.ResourceTierProduction:
		return "production"
	default:
		return "eco"
	}
}

func tierFromString(s string) data.ResourceTier {
	switch s {
	case "production":
		return data.ResourceTierProduction
	default:
		return data.ResourceTierEco
	}
}

// ListResourceAllocations returns resource quotas + tiers for every managed namespace.
func (s *Server) ListResourceAllocations(ctx context.Context) ([]data.ResourceAllocation, error) {
	// List all namespaces with resource-tier label.
	selector := labels.Selector(labels.ManagedBy, labels.ManagedByLucity) + "," + labels.ResourceTier

	nsList, err := s.k8s.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list namespaces: %v", err)
	}

	var allocations []data.ResourceAllocation
	for _, ns := range nsList.Items {
		ws := ns.Labels[labels.Workspace]
		project := ns.Labels[labels.Project]
		env := ns.Labels[labels.Environment]
		if ws == "" || project == "" || env == "" {
			continue
		}

		tier := tierFromString(ns.Labels[labels.ResourceTier])
		cpuMillis, memMB, diskMB := s.namespaceAllocations(ctx, ns.Name)

		allocations = append(allocations, data.ResourceAllocation{
			Workspace:     ws,
			Project:       project,
			Environment:   env,
			Tier:          tier,
			CPUMillicores: cpuMillis,
			MemoryMB:      memMB,
			DiskMB:        diskMB,
		})
	}
	return allocations, nil
}

// namespaceAllocations returns the actual resource usage for a namespace by summing
// pod container requests (CPU, memory) and PVC storage requests (disk).
// This reflects what's actually deployed, not what the ResourceQuota allows.
func (s *Server) namespaceAllocations(ctx context.Context, namespace string) (cpuMillis, memMB, diskMB int) {
	// Sum CPU and memory requests from all running pods.
	pods, err := s.k8s.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		slog.Warn("failed to list pods for allocations", "namespace", namespace, "error", err)
		return 0, 0, 0
	}
	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers {
			if req, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
				cpuMillis += int(req.MilliValue())
			}
			if req, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
				memMB += int(req.Value() / (1024 * 1024))
			}
		}
	}

	// Sum storage requests from all PVCs.
	pvcs, err := s.k8s.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list PVCs for allocations", "namespace", namespace, "error", err)
		return cpuMillis, memMB, 0
	}
	for _, pvc := range pvcs.Items {
		if req, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			diskMB += int(req.Value() / (1024 * 1024))
		}
	}

	return cpuMillis, memMB, diskMB
}

// DatabaseCredentials returns the resolved connection credentials for a database.
func (s *Server) DatabaseCredentials(ctx context.Context, workspace, project, environment, dbName string) (data.DatabaseCredentials, error) {
	creds, err := database.CredentialsFromSecret(ctx, s.k8s, workspace, project, environment, dbName)
	if err != nil {
		if errors.Is(err, database.ErrNotReady) {
			return data.DatabaseCredentials{}, status.Errorf(codes.FailedPrecondition, "database is provisioning")
		}
		return data.DatabaseCredentials{}, status.Errorf(codes.Internal, "failed to read credentials: %v", err)
	}

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", creds.User, creds.Password, creds.Host, creds.Port, creds.DBName)
	return data.DatabaseCredentials{
		Host:     creds.Host,
		Port:     creds.Port,
		DBName:   creds.DBName,
		User:     creds.User,
		Password: creds.Password,
		URI:      uri,
	}, nil
}

// --- User GitHub token CRUD (K8s Secrets) ---

func githubTokenSecretName(userID string) string {
	return "github-token-" + strings.ToLower(userID)
}

// StoreUserGitHubToken persists a user's GitHub OAuth token in a K8s Secret.
func (s *Server) StoreUserGitHubToken(ctx context.Context, userID, accessToken, refreshToken string, expiresAt int64) error {
	secretName := githubTokenSecretName(userID)

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
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    strconv.FormatInt(expiresAt, 10),
		},
	}

	existing, err := s.k8s.CoreV1().Secrets(labels.LucityNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return status.Errorf(codes.Internal, "failed to get token secret: %v", err)
		}
		if _, err := s.k8s.CoreV1().Secrets(labels.LucityNamespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			return status.Errorf(codes.Internal, "failed to create token secret: %v", err)
		}
	} else {
		existing.Data = nil // clear binary data, use StringData
		existing.StringData = secret.StringData
		existing.Labels = secret.Labels
		if _, err := s.k8s.CoreV1().Secrets(labels.LucityNamespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
			return status.Errorf(codes.Internal, "failed to update token secret: %v", err)
		}
	}

	slog.Info("stored GitHub token", "user", userID)
	return nil
}

// UserGitHubToken returns a user's stored GitHub OAuth token. Connected is
// false (with empty fields) when no token exists yet.
func (s *Server) UserGitHubToken(ctx context.Context, userID string) (data.UserGitHubToken, error) {
	secretName := githubTokenSecretName(userID)

	secret, err := s.k8s.CoreV1().Secrets(labels.LucityNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return data.UserGitHubToken{Connected: false}, nil
		}
		return data.UserGitHubToken{}, status.Errorf(codes.Internal, "failed to get token secret: %v", err)
	}

	var expiresAt int64
	if raw := string(secret.Data["expires_at"]); raw != "" {
		expiresAt, _ = strconv.ParseInt(raw, 10, 64)
	}

	return data.UserGitHubToken{
		Connected:    true,
		AccessToken:  string(secret.Data["access_token"]),
		RefreshToken: string(secret.Data["refresh_token"]),
		ExpiresAt:    expiresAt,
	}, nil
}

// DeleteUserGitHubToken removes a user's stored GitHub OAuth token. Idempotent.
func (s *Server) DeleteUserGitHubToken(ctx context.Context, userID string) error {
	secretName := githubTokenSecretName(userID)

	if err := s.k8s.CoreV1().Secrets(labels.LucityNamespace).Delete(ctx, secretName, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return status.Errorf(codes.Internal, "failed to delete token secret: %v", err)
	}

	slog.Info("deleted GitHub token", "user", userID)
	return nil
}

// SuspendWorkspace flips the suspended flag for every project/environment
// in a workspace by writing through the inproc packager.
func (s *Server) SuspendWorkspace(ctx context.Context, workspace string, suspended bool) error {
	if workspace == "" {
		return status.Errorf(codes.InvalidArgument, "workspace required")
	}

	// 1. List all namespaces belonging to this workspace.
	nsList, err := s.k8s.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: labels.Selector(labels.Workspace, workspace),
	})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to list workspace namespaces: %v", err)
	}

	// 2. Deduplicate project/environment pairs.
	type envKey struct{ project, environment string }
	seen := make(map[envKey]bool)
	for _, ns := range nsList.Items {
		project := ns.Labels[labels.Project]
		environment := ns.Labels[labels.Environment]
		if project == "" || environment == "" {
			continue
		}
		seen[envKey{project, environment}] = true
	}

	// 3. Write suspended flag via packager (GitOps values.yaml).
	// The packager commits the change and triggers ArgoCD sync.
	// ArgoCD then enforces the suspension (replicas=0, CronJobs suspended, CNPG hibernated, HTTPRoutes removed).
	var failed int
	for ek := range seen {
		if err := s.packager.SetSuspended(ctx, workspace, ek.project, ek.environment, suspended); err != nil {
			slog.Error("failed to set suspended in gitops repo", "project", ek.project, "environment", ek.environment, "error", err)
			failed++
		}
	}

	action := "suspended"
	if !suspended {
		action = "resumed"
	}

	if failed > 0 {
		verb := "suspend"
		if !suspended {
			verb = "resume"
		}
		return status.Errorf(codes.Internal, "failed to %s %d of %d environments", verb, failed, len(seen))
	}

	slog.Info("workspace "+action, "workspace", workspace, "environments", len(seen))
	return nil
}
