package ovh

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ovh/go-ovh/ovh"

	"github.com/zeitlos/lucity/services/conductor/internal/objectstorage"
)

const objectStoreRole = "objectstore_operator"

type Backend struct {
	client    *ovh.Client
	projectID string
	region    string
}

func New(endpoint, applicationKey, applicationSecret, consumerKey, projectID, region string) (*Backend, error) {
	client, err := ovh.NewClient(endpoint, applicationKey, applicationSecret, consumerKey)

	if err != nil {
		return nil, fmt.Errorf("create ovh client: %w", err)
	}

	return &Backend{
		client:    client,
		projectID: projectID,
		region:    region,
	}, nil
}

var _ objectstorage.Backend = (*Backend)(nil)

type ovhUser struct {
	ID          int64     `json:"id"`
	Description string    `json:"description"`
	Roles       []ovhRole `json:"roles"`
}

type ovhRole struct {
	Name string `json:"name"`
}

type ovhUserCreate struct {
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
}

type ovhCredential struct {
	Access string `json:"access"`
	Secret string `json:"secret"`
}

type ovhSecret struct {
	Secret string `json:"secret"`
}

type ovhStorageCreate struct {
	Name    string            `json:"name"`
	OwnerID int64             `json:"ownerId"`
	Tags    map[string]string `json:"tags,omitempty"`
}

type ovhStorage struct {
	Name         string `json:"name"`
	OwnerID      int64  `json:"ownerId"`
	ObjectsSize  int64  `json:"objectsSize"`
	ObjectsCount int64  `json:"objectsCount"`
}

type ovhStorageObject struct {
	Key       string `json:"key"`
	VersionID string `json:"versionId"`
}

func (b *Backend) CreateBucket(ctx context.Context, req objectstorage.BucketRequest) (objectstorage.BucketConnection, error) {
	userID, err := b.ensureUser(ctx, req.Workspace)

	if err != nil {
		return objectstorage.BucketConnection{}, err
	}

	access, secret, err := b.ensureCredential(ctx, userID)

	if err != nil {
		return objectstorage.BucketConnection{}, err
	}

	createPath := fmt.Sprintf("/cloud/project/%s/region/%s/storage", b.projectID, b.apiRegion())

	if err := b.client.PostWithContext(ctx, createPath, ovhStorageCreate{
		Name:    req.PhysicalName,
		OwnerID: userID,
		Tags:    req.Tags,
	}, nil); err != nil {
		return objectstorage.BucketConnection{}, fmt.Errorf("create bucket: %w", err)
	}

	if err := b.syncPolicy(ctx, userID); err != nil {
		return objectstorage.BucketConnection{}, fmt.Errorf("sync access policy: %w", err)
	}

	return objectstorage.BucketConnection{
		Endpoint:        b.endpoint(),
		Region:          b.region,
		AccessKeyID:     access,
		SecretAccessKey: secret,
	}, nil
}

func (b *Backend) DeleteBucket(ctx context.Context, region, physicalName string) error {
	base := fmt.Sprintf("/cloud/project/%s/region/%s/storage/%s", b.projectID, strings.ToUpper(region), physicalName)

	if err := b.empty(ctx, base); err != nil {
		return err
	}

	if err := b.client.DeleteWithContext(ctx, base, nil); err != nil {
		if isNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete bucket: %w", err)
	}

	return nil
}

func (b *Backend) Stats(ctx context.Context) (map[string]objectstorage.BucketStats, error) {
	var containers []ovhStorage

	if err := b.client.GetWithContext(ctx, fmt.Sprintf("/cloud/project/%s/region/%s/storage", b.projectID, b.apiRegion()), &containers); err != nil {
		return nil, fmt.Errorf("list storage: %w", err)
	}

	stats := make(map[string]objectstorage.BucketStats, len(containers))

	for _, container := range containers {
		stats[container.Name] = objectstorage.BucketStats{
			SizeBytes:   container.ObjectsSize,
			ObjectCount: container.ObjectsCount,
		}
	}

	return stats, nil
}

func (b *Backend) ensureUser(ctx context.Context, workspace string) (int64, error) {
	description := userDescription(workspace)

	var users []ovhUser

	if err := b.client.GetWithContext(ctx, fmt.Sprintf("/cloud/project/%s/user", b.projectID), &users); err != nil {
		return 0, fmt.Errorf("list users: %w", err)
	}

	for _, user := range users {
		if user.Description == description {
			return user.ID, nil
		}
	}

	var created ovhUser

	if err := b.client.PostWithContext(ctx, fmt.Sprintf("/cloud/project/%s/user", b.projectID), ovhUserCreate{
		Description: description,
		Roles:       []string{objectStoreRole},
	}, &created); err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}

	return created.ID, nil
}

func (b *Backend) ensureCredential(ctx context.Context, userID int64) (string, string, error) {
	listPath := fmt.Sprintf("/cloud/project/%s/user/%d/s3Credentials", b.projectID, userID)

	var existing []ovhCredential

	if err := b.client.GetWithContext(ctx, listPath, &existing); err != nil {
		return "", "", fmt.Errorf("list credentials: %w", err)
	}

	if len(existing) > 0 {
		access := existing[0].Access

		var secret ovhSecret

		if err := b.client.PostWithContext(ctx, listPath+"/"+access+"/secret", nil, &secret); err != nil {
			return "", "", fmt.Errorf("fetch credential secret: %w", err)
		}

		return access, secret.Secret, nil
	}

	var created ovhCredential

	if err := b.client.PostWithContext(ctx, listPath, nil, &created); err != nil {
		return "", "", fmt.Errorf("create credential: %w", err)
	}

	return created.Access, created.Secret, nil
}

func (b *Backend) syncPolicy(ctx context.Context, userID int64) error {
	var containers []ovhStorage

	if err := b.client.GetWithContext(ctx, fmt.Sprintf("/cloud/project/%s/region/%s/storage", b.projectID, b.apiRegion()), &containers); err != nil {
		return fmt.Errorf("list storage: %w", err)
	}

	resources := make([]string, 0, len(containers)*2)

	for _, container := range containers {
		if container.OwnerID != userID {
			continue
		}
		resources = append(resources, "arn:aws:s3:::"+container.Name, "arn:aws:s3:::"+container.Name+"/*")
	}

	if len(resources) == 0 {
		return nil
	}

	document, err := json.Marshal(map[string]any{
		"Statement": []map[string]any{{
			"Sid":    "lucity",
			"Effect": "Allow",
			"Action": []string{
				"s3:GetObject", "s3:PutObject", "s3:DeleteObject",
				"s3:ListBucket", "s3:ListMultipartUploadParts",
				"s3:ListBucketMultipartUploads", "s3:AbortMultipartUpload",
				"s3:GetBucketLocation",
			},
			"Resource": resources,
		}},
	})

	if err != nil {
		return err
	}

	return b.client.PostWithContext(ctx, fmt.Sprintf("/cloud/project/%s/user/%d/policy", b.projectID, userID), map[string]string{"policy": string(document)}, nil)
}

func (b *Backend) empty(ctx context.Context, base string) error {
	for range 1000 {
		var objects []ovhStorageObject

		if err := b.client.GetWithContext(ctx, base+"/object?withVersions=true", &objects); err != nil {
			if isNotFound(err) {
				return nil
			}
			return fmt.Errorf("list objects: %w", err)
		}

		if len(objects) == 0 {
			return nil
		}

		if err := b.client.PostWithContext(ctx, base+"/bulkDeleteObjects", map[string]any{"objects": objects}, nil); err != nil {
			return fmt.Errorf("bulk delete objects: %w", err)
		}
	}

	return errors.New("bucket did not drain within the deletion budget")
}

func (b *Backend) apiRegion() string {
	return strings.ToUpper(b.region)
}

func (b *Backend) endpoint() string {
	return "https://s3." + strings.ToLower(b.region) + ".io.cloud.ovh.net"
}

func userDescription(workspace string) string {
	sum := sha256.Sum256([]byte(workspace))
	hash := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum[:]))[:12]
	return "lucity-ws-" + hash
}

func isNotFound(err error) bool {
	var apiErr *ovh.APIError
	return errors.As(err, &apiErr) && apiErr.Code == 404
}
