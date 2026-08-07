package backuparchive

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Endpoint        string
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

type Client struct {
	config Config
	s3     *s3.Client
}

func New(config Config) *Client {
	if config.AccessKeyID == "" || config.Bucket == "" {
		return nil
	}

	return &Client{
		config: config,
		s3: s3.New(s3.Options{
			Region:       strings.ToLower(config.Region),
			BaseEndpoint: aws.String(config.Endpoint),
			Credentials:  credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""),
			UsePathStyle: true,
		}),
	}
}

// LatestRestorePoint reports how far forward a database can be recovered: the
// moment its most recently archived write-ahead log segment reached the store.
//
// A segment is only archived once it is full, or once the archive timeout closes
// it, and the timeout only fires when something was written. An idle database can
// therefore sit hours behind, and recovering past this point fails.
//
// Segment names increase monotonically and object stores list keys in binary
// order, so the newest is the last key under the newest directory. Both listings
// are bounded: barman holds at most 256 segments per directory.
func (c *Client) LatestRestorePoint(ctx context.Context, workspace, serverName string) (*time.Time, error) {
	root := workspace + "/" + serverName + "/wals/"

	directories, err := c.s3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(c.config.Bucket),
		Prefix:    aws.String(root),
		Delimiter: aws.String("/"),
	})

	if err != nil {
		return nil, err
	}

	if len(directories.CommonPrefixes) == 0 {
		return nil, nil
	}

	newest := directories.CommonPrefixes[len(directories.CommonPrefixes)-1]

	segments, err := c.s3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.config.Bucket),
		Prefix: newest.Prefix,
	})

	if err != nil {
		return nil, err
	}

	var latest *time.Time

	for _, object := range segments.Contents {
		if object.LastModified == nil {
			continue
		}

		if latest == nil || object.LastModified.After(*latest) {
			latest = object.LastModified
		}
	}

	return latest, nil
}
