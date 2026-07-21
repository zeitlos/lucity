package objectstorage

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

const (
	presignTTL    = 15 * time.Minute
	maxObjectKeys = 1000
)

type Object struct {
	Key          string
	Size         int64
	LastModified time.Time
}

type Folder struct {
	Prefix string
}

type ObjectListing struct {
	Prefix  string
	Folders []Folder
	Objects []Object
}

func (m *Manager) Objects(ctx context.Context, id platform.BucketID, prefix string) (*ObjectListing, error) {
	creds, err := m.Credentials(ctx, id)

	if err != nil {
		return nil, err
	}

	client := s3client(creds)

	page, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(creds.Bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int32(maxObjectKeys),
	})

	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}

	listing := &ObjectListing{Prefix: prefix, Folders: []Folder{}, Objects: []Object{}}

	for _, common := range page.CommonPrefixes {
		listing.Folders = append(listing.Folders, Folder{Prefix: aws.ToString(common.Prefix)})
	}

	for _, object := range page.Contents {
		key := aws.ToString(object.Key)

		if key == prefix {
			continue
		}

		listing.Objects = append(listing.Objects, Object{
			Key:          key,
			Size:         aws.ToInt64(object.Size),
			LastModified: aws.ToTime(object.LastModified),
		})
	}

	return listing, nil
}

func (m *Manager) PresignDownload(ctx context.Context, id platform.BucketID, key string) (string, error) {
	creds, err := m.Credentials(ctx, id)

	if err != nil {
		return "", err
	}

	presigner := s3.NewPresignClient(s3client(creds))

	request, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(creds.Bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=%q", path.Base(key))),
	}, s3.WithPresignExpires(presignTTL))

	if err != nil {
		return "", fmt.Errorf("presign download: %w", err)
	}

	return request.URL, nil
}

func (m *Manager) PresignUpload(ctx context.Context, id platform.BucketID, key string) (string, error) {
	creds, err := m.Credentials(ctx, id)

	if err != nil {
		return "", err
	}

	presigner := s3.NewPresignClient(s3client(creds))

	request, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(creds.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(presignTTL))

	if err != nil {
		return "", fmt.Errorf("presign upload: %w", err)
	}

	return request.URL, nil
}

func (m *Manager) DeleteObject(ctx context.Context, id platform.BucketID, key string) error {
	creds, err := m.Credentials(ctx, id)

	if err != nil {
		return err
	}

	if _, err := s3client(creds).DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(creds.Bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}

	return nil
}

func (m *Manager) ensureCORS(ctx context.Context, creds *Credentials) {
	_, err := s3client(creds).PutBucketCors(ctx, &s3.PutBucketCorsInput{
		Bucket: aws.String(creds.Bucket),
		CORSConfiguration: &s3types.CORSConfiguration{
			CORSRules: []s3types.CORSRule{{
				AllowedMethods: []string{"GET", "PUT", "HEAD", "DELETE"},
				AllowedOrigins: []string{"*"},
				AllowedHeaders: []string{"*"},
				ExposeHeaders:  []string{"ETag"},
				MaxAgeSeconds:  aws.Int32(3600),
			}},
		},
	})

	if err != nil {
		slog.WarnContext(ctx, "configure bucket CORS failed", "bucket", creds.Bucket, "error", err)
	}
}

func s3client(creds *Credentials) *s3.Client {
	return s3.New(s3.Options{
		Region:       strings.ToLower(creds.Region),
		BaseEndpoint: aws.String(creds.Endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, ""),
		UsePathStyle: true,
	})
}
