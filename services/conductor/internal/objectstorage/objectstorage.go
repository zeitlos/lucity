package objectstorage

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	"github.com/zeitlos/lucity/pkg/labels"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var ErrNotImplemented = errors.New("objectstorage: not implemented")

type Permission int

const (
	ReadWrite Permission = iota
	ReadOnly
)

const (
	secretKeyName            = "name"
	secretKeyPhysicalName    = "bucket"
	secretKeyRegion          = "region"
	secretKeyEndpoint        = "endpoint"
	secretKeyAccessKeyID     = "accessKeyId"
	secretKeySecretAccessKey = "secretAccessKey"
	secretKeyPublic          = "public"
	secretKeyPublicEndpoint  = "publicEndpoint"
	secretKeyPullZoneID      = "cdnPullZoneId"
)

type Bucket struct {
	ID             platform.BucketID
	Name           string
	Region         string
	Endpoint       string
	PublicEndpoint string
	SizeBytes      int64
	ObjectCount    int64
	Public         bool
	CreatedAt      time.Time
}

type Credentials struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
}

type Interface interface {
	Buckets(ctx context.Context, environment platform.EnvironmentID) ([]Bucket, error)
	Bucket(ctx context.Context, id platform.BucketID) (*Bucket, error)
	Create(ctx context.Context, environment platform.EnvironmentID, name string) (*Bucket, error)
	Delete(ctx context.Context, id platform.BucketID) error
	Credentials(ctx context.Context, id platform.BucketID) (*Credentials, error)
	SetPublic(ctx context.Context, id platform.BucketID, public bool) (*Bucket, error)
}

type Backend interface {
	CreateBucket(ctx context.Context, req BucketRequest) (BucketConnection, error)
	DeleteBucket(ctx context.Context, region, physicalName string) error
	Stats(ctx context.Context) (map[string]BucketStats, error)
	Credentials(ctx context.Context, req CredentialsRequest) (BucketConnection, error)
	SetPublic(ctx context.Context, req SetPublicRequest) (SetPublicResult, error)
}

type CredentialsRequest struct {
	Workspace  string
	Permission Permission
	ReadScope  []string
}

type SetPublicRequest struct {
	Public     bool
	Slug       string
	OriginURL  string
	Region     string
	Workspace  string
	PublicSet  []string
	PullZoneID string
}

type SetPublicResult struct {
	PullZoneID     string
	PublicEndpoint string
}

type BucketStats struct {
	SizeBytes   int64
	ObjectCount int64
}

type BucketRequest struct {
	Workspace    string
	PhysicalName string
	Tags         map[string]string
}

type BucketConnection struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

type bucketState struct {
	Name            string
	PhysicalName    string
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Public          bool
	PublicEndpoint  string
	PullZoneID      string
}

func parseBucketState(secret *corev1.Secret) bucketState {
	return bucketState{
		Name:            string(secret.Data[secretKeyName]),
		PhysicalName:    string(secret.Data[secretKeyPhysicalName]),
		Region:          string(secret.Data[secretKeyRegion]),
		Endpoint:        string(secret.Data[secretKeyEndpoint]),
		AccessKeyID:     string(secret.Data[secretKeyAccessKeyID]),
		SecretAccessKey: string(secret.Data[secretKeySecretAccessKey]),
		Public:          string(secret.Data[secretKeyPublic]) == "true",
		PublicEndpoint:  string(secret.Data[secretKeyPublicEndpoint]),
		PullZoneID:      string(secret.Data[secretKeyPullZoneID]),
	}
}

type Manager struct {
	backend Backend
	kube    kubernetes.Interface
}

func NewManager(backend Backend, kube kubernetes.Interface) *Manager {
	return &Manager{backend: backend, kube: kube}
}

var _ Interface = (*Manager)(nil)

func (m *Manager) Create(ctx context.Context, environment platform.EnvironmentID, name string) (*Bucket, error) {
	id := platform.BucketID{
		Workspace:   environment.Workspace,
		Project:     environment.Project,
		Environment: environment.Name,
		Name:        name,
	}

	namespace := id.Namespace()
	secretName := id.SecretName()

	if _, err := m.kube.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{}); err == nil {
		return nil, fmt.Errorf("bucket %q already exists", name)
	} else if !apierrors.IsNotFound(err) {
		return nil, err
	}

	physical := physicalName(name)

	conn, err := m.backend.CreateBucket(ctx, BucketRequest{
		Workspace:    environment.Workspace,
		PhysicalName: physical,
		Tags: map[string]string{
			labels.Workspace:           environment.Workspace,
			labels.Project:             environment.Project,
			labels.Environment:         environment.Name,
			labels.ObjectStorageBucket: name,
		},
	})

	if err != nil {
		return nil, err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				labels.ObjectStorageBucket: name,
				labels.Workspace:           environment.Workspace,
				labels.Project:             environment.Project,
				labels.Environment:         environment.Name,
				labels.ManagedBy:           labels.ManagedByLucity,
			},
		},
		Data: map[string][]byte{
			secretKeyName:            []byte(name),
			secretKeyPhysicalName:    []byte(physical),
			secretKeyRegion:          []byte(conn.Region),
			secretKeyEndpoint:        []byte(conn.Endpoint),
			secretKeyAccessKeyID:     []byte(conn.AccessKeyID),
			secretKeySecretAccessKey: []byte(conn.SecretAccessKey),
		},
	}

	if _, err := m.kube.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		if delErr := m.backend.DeleteBucket(ctx, conn.Region, physical); delErr != nil {
			return nil, fmt.Errorf("write bucket secret: %w (orphaned bucket cleanup failed: %v)", err, delErr)
		}

		return nil, err
	}

	return &Bucket{
		ID:       id,
		Name:     name,
		Region:   conn.Region,
		Endpoint: conn.Endpoint,
	}, nil
}

func (m *Manager) Buckets(ctx context.Context, environment platform.EnvironmentID) ([]Bucket, error) {
	list, err := m.kube.CoreV1().Secrets(environment.Namespace()).List(ctx, metav1.ListOptions{
		LabelSelector: labels.ObjectStorageBucket,
	})

	if err != nil {
		return nil, fmt.Errorf("list bucket secrets: %w", err)
	}

	stats, statsErr := m.backend.Stats(ctx)

	if statsErr != nil {
		slog.WarnContext(ctx, "object storage stats unavailable", "error", statsErr)
	}

	buckets := make([]Bucket, 0, len(list.Items))

	for i := range list.Items {
		bucket := bucketFromSecret(&list.Items[i])

		if stat, ok := stats[parseBucketState(&list.Items[i]).PhysicalName]; ok {
			bucket.SizeBytes = stat.SizeBytes
			bucket.ObjectCount = stat.ObjectCount
		}

		buckets = append(buckets, bucket)
	}

	return buckets, nil
}

func (m *Manager) Bucket(ctx context.Context, id platform.BucketID) (*Bucket, error) {
	secret, err := m.secret(ctx, id)

	if err != nil {
		return nil, err
	}

	bucket := bucketFromSecret(secret)

	if stats, err := m.backend.Stats(ctx); err == nil {
		if stat, ok := stats[parseBucketState(secret).PhysicalName]; ok {
			bucket.SizeBytes = stat.SizeBytes
			bucket.ObjectCount = stat.ObjectCount
		}
	}

	return &bucket, nil
}

func (m *Manager) Credentials(ctx context.Context, id platform.BucketID) (*Credentials, error) {
	secret, err := m.secret(ctx, id)

	if err != nil {
		return nil, err
	}

	state := parseBucketState(secret)

	return &Credentials{
		Endpoint:        state.Endpoint,
		Region:          state.Region,
		Bucket:          state.PhysicalName,
		AccessKeyID:     state.AccessKeyID,
		SecretAccessKey: state.SecretAccessKey,
	}, nil
}

func (m *Manager) Delete(ctx context.Context, id platform.BucketID) error {
	secret, err := m.secret(ctx, id)

	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	state := parseBucketState(secret)

	if state.Public {
		remaining, err := m.publicPhysicalNames(ctx, id.Workspace, state.PhysicalName, false)

		if err != nil {
			return err
		}

		if _, err := m.backend.SetPublic(ctx, SetPublicRequest{
			Public:     false,
			Slug:       state.PhysicalName,
			Workspace:  id.Workspace,
			PublicSet:  remaining,
			PullZoneID: state.PullZoneID,
		}); err != nil {
			slog.WarnContext(ctx, "unpublish bucket during delete failed", "bucket", id.String(), "error", err)
		}
	}

	if err := m.backend.DeleteBucket(ctx, state.Region, state.PhysicalName); err != nil {
		return fmt.Errorf("delete backend bucket: %w", err)
	}

	if err := m.kube.CoreV1().Secrets(id.Namespace()).Delete(ctx, id.SecretName(), metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete bucket secret: %w", err)
	}

	return nil
}

func (m *Manager) SetPublic(ctx context.Context, id platform.BucketID, public bool) (*Bucket, error) {
	secret, err := m.secret(ctx, id)

	if err != nil {
		return nil, err
	}

	state := parseBucketState(secret)

	if state.Public == public {
		bucket := bucketFromSecret(secret)
		return &bucket, nil
	}

	publicSet, err := m.publicPhysicalNames(ctx, id.Workspace, state.PhysicalName, public)

	if err != nil {
		return nil, err
	}

	result, err := m.backend.SetPublic(ctx, SetPublicRequest{
		Public:     public,
		Slug:       state.PhysicalName,
		OriginURL:  originURL(state.Endpoint, state.PhysicalName),
		Region:     strings.ToLower(state.Region),
		Workspace:  id.Workspace,
		PublicSet:  publicSet,
		PullZoneID: state.PullZoneID,
	})

	if err != nil {
		return nil, err
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	if public {
		secret.Data[secretKeyPublic] = []byte("true")
		secret.Data[secretKeyPullZoneID] = []byte(result.PullZoneID)
		secret.Data[secretKeyPublicEndpoint] = []byte(result.PublicEndpoint)
	} else {
		secret.Data[secretKeyPublic] = []byte("false")
		delete(secret.Data, secretKeyPullZoneID)
		delete(secret.Data, secretKeyPublicEndpoint)
	}

	updated, err := m.kube.CoreV1().Secrets(id.Namespace()).Update(ctx, secret, metav1.UpdateOptions{})

	if err != nil {
		return nil, fmt.Errorf("update bucket secret: %w", err)
	}

	bucket := bucketFromSecret(updated)

	return &bucket, nil
}

func (m *Manager) publicPhysicalNames(ctx context.Context, workspace, toggled string, include bool) ([]string, error) {
	list, err := m.kube.CoreV1().Secrets("").List(ctx, metav1.ListOptions{
		LabelSelector: labels.Workspace + "=" + workspace + "," + labels.ObjectStorageBucket,
	})

	if err != nil {
		return nil, fmt.Errorf("list workspace bucket secrets: %w", err)
	}

	set := make(map[string]struct{})

	for i := range list.Items {
		state := parseBucketState(&list.Items[i])

		if state.Public {
			set[state.PhysicalName] = struct{}{}
		}
	}

	if include {
		set[toggled] = struct{}{}
	} else {
		delete(set, toggled)
	}

	names := make([]string, 0, len(set))

	for name := range set {
		names = append(names, name)
	}

	return names, nil
}

func (m *Manager) secret(ctx context.Context, id platform.BucketID) (*corev1.Secret, error) {
	secret, err := m.kube.CoreV1().Secrets(id.Namespace()).Get(ctx, id.SecretName(), metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("bucket %q not found", id)
	}

	if err != nil {
		return nil, fmt.Errorf("get bucket secret: %w", err)
	}

	return secret, nil
}

func physicalName(name string) string {
	return name + "-" + randCrockford32(12)
}

func randCrockford32(n int) string {
	const alphabet = "0123456789abcdefghjkmnpqrstvwxyz"

	b := make([]byte, n)
	_, _ = rand.Read(b)

	for i, v := range b {
		b[i] = alphabet[v&31]
	}

	return string(b)
}

func originURL(endpoint, physical string) string {
	return "https://" + physical + "." + strings.TrimPrefix(endpoint, "https://")
}

func bucketFromSecret(secret *corev1.Secret) Bucket {
	state := parseBucketState(secret)

	return Bucket{
		ID: platform.BucketID{
			Workspace:   secret.Labels[labels.Workspace],
			Project:     secret.Labels[labels.Project],
			Environment: secret.Labels[labels.Environment],
			Name:        secret.Labels[labels.ObjectStorageBucket],
		},
		Name:           state.Name,
		Region:         state.Region,
		Endpoint:       state.Endpoint,
		PublicEndpoint: state.PublicEndpoint,
		Public:         state.Public,
		CreatedAt:      secret.CreationTimestamp.Time,
	}
}
