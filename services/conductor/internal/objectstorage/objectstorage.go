package objectstorage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	"github.com/zeitlos/lucity/pkg/labels"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Bucket struct {
	ID          platform.BucketID
	Name        string
	Region      string
	Endpoint    string
	SizeBytes   int64
	ObjectCount int64
	Public      bool
	CreatedAt   time.Time
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
}

type Backend interface {
	CreateBucket(ctx context.Context, req BucketRequest) (BucketConnection, error)
	DeleteBucket(ctx context.Context, region, physicalName string) error
	Stats(ctx context.Context) (map[string]BucketStats, error)
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

	physical := physicalName()

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
			"name":            []byte(name),
			"bucket":          []byte(physical),
			"region":          []byte(conn.Region),
			"endpoint":        []byte(conn.Endpoint),
			"accessKeyId":     []byte(conn.AccessKeyID),
			"secretAccessKey": []byte(conn.SecretAccessKey),
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

		if stat, ok := stats[string(list.Items[i].Data["bucket"])]; ok {
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
		if stat, ok := stats[string(secret.Data["bucket"])]; ok {
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

	return &Credentials{
		Endpoint:        string(secret.Data["endpoint"]),
		Region:          string(secret.Data["region"]),
		Bucket:          string(secret.Data["bucket"]),
		AccessKeyID:     string(secret.Data["accessKeyId"]),
		SecretAccessKey: string(secret.Data["secretAccessKey"]),
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

	region := string(secret.Data["region"])
	physical := string(secret.Data["bucket"])

	if err := m.backend.DeleteBucket(ctx, region, physical); err != nil {
		return fmt.Errorf("delete backend bucket: %w", err)
	}

	if err := m.kube.CoreV1().Secrets(id.Namespace()).Delete(ctx, id.SecretName(), metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete bucket secret: %w", err)
	}

	return nil
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

func physicalName() string {
	suffix := make([]byte, 16)
	_, _ = rand.Read(suffix)

	return "lc-" + hex.EncodeToString(suffix)
}

func bucketFromSecret(secret *corev1.Secret) Bucket {
	return Bucket{
		ID: platform.BucketID{
			Workspace:   secret.Labels[labels.Workspace],
			Project:     secret.Labels[labels.Project],
			Environment: secret.Labels[labels.Environment],
			Name:        secret.Labels[labels.ObjectStorageBucket],
		},
		Name:      string(secret.Data["name"]),
		Region:    string(secret.Data["region"]),
		Endpoint:  string(secret.Data["endpoint"]),
		CreatedAt: secret.CreationTimestamp.Time,
	}
}
