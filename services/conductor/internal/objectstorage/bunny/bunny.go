package bunny

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/zeitlos/lucity/pkg/bunny"
	"github.com/zeitlos/lucity/services/conductor/internal/objectstorage"
)

type Backend struct {
	inner  objectstorage.Backend
	api    *bunny.Client
	domain string
}

func New(inner objectstorage.Backend, api *bunny.Client, domain string) *Backend {
	return &Backend{
		inner:  inner,
		api:    api,
		domain: domain,
	}
}

var _ objectstorage.Backend = (*Backend)(nil)

func (b *Backend) CreateBucket(ctx context.Context, req objectstorage.BucketRequest) (objectstorage.BucketConnection, error) {
	return b.inner.CreateBucket(ctx, req)
}

func (b *Backend) DeleteBucket(ctx context.Context, region, physicalName string) error {
	return b.inner.DeleteBucket(ctx, region, physicalName)
}

func (b *Backend) Stats(ctx context.Context) (map[string]objectstorage.BucketStats, error) {
	return b.inner.Stats(ctx)
}

func (b *Backend) Credentials(ctx context.Context, req objectstorage.CredentialsRequest) (objectstorage.BucketConnection, error) {
	return b.inner.Credentials(ctx, req)
}

func (b *Backend) SetPublic(ctx context.Context, req objectstorage.SetPublicRequest) (objectstorage.SetPublicResult, error) {
	if !req.Public {
		if err := b.unpublish(ctx, req); err != nil {
			return objectstorage.SetPublicResult{}, err
		}

		if err := b.deleteDNSRecord(ctx, req.Slug); err != nil {
			return objectstorage.SetPublicResult{}, fmt.Errorf("delete dns record: %w", err)
		}

		if _, err := b.inner.Credentials(ctx, objectstorage.CredentialsRequest{
			Workspace:  req.Workspace,
			Permission: objectstorage.ReadOnly,
			ReadScope:  req.PublicSet,
		}); err != nil {
			return objectstorage.SetPublicResult{}, fmt.Errorf("sync read credentials: %w", err)
		}

		return objectstorage.SetPublicResult{}, nil
	}

	origin, err := b.inner.Credentials(ctx, objectstorage.CredentialsRequest{
		Workspace:  req.Workspace,
		Permission: objectstorage.ReadOnly,
		ReadScope:  req.PublicSet,
	})

	if err != nil {
		return objectstorage.SetPublicResult{}, fmt.Errorf("read-only credentials: %w", err)
	}

	pullZoneID, err := b.ensurePullZone(ctx, req, origin)

	if err != nil {
		return objectstorage.SetPublicResult{}, err
	}

	hostname := req.Slug + "." + b.domain

	if err := b.api.AddHostname(ctx, pullZoneID, hostname); err != nil && !bunny.IsAlreadyExists(err) {
		return objectstorage.SetPublicResult{}, err
	}

	if err := b.ensureDNSRecord(ctx, req.Slug); err != nil {
		return objectstorage.SetPublicResult{}, fmt.Errorf("ensure dns record: %w", err)
	}

	go b.issueCertificate(hostname)

	return objectstorage.SetPublicResult{
		PullZoneID:     strconv.FormatInt(pullZoneID, 10),
		PublicEndpoint: "https://" + hostname,
	}, nil
}

func (b *Backend) ensurePullZone(ctx context.Context, req objectstorage.SetPublicRequest, origin objectstorage.BucketConnection) (int64, error) {
	if req.PullZoneID != "" {
		return strconv.ParseInt(req.PullZoneID, 10, 64)
	}

	existing, err := b.api.PullZoneByName(ctx, req.Slug)

	if err != nil {
		return 0, err
	}

	if existing != nil {
		return existing.ID, nil
	}

	created, err := b.api.CreatePullZone(ctx, bunny.PullZoneSpec{
		Name:                 req.Slug,
		OriginURL:            req.OriginURL,
		AWSSigningEnabled:    true,
		AWSSigningKey:        origin.AccessKeyID,
		AWSSigningSecret:     origin.SecretAccessKey,
		AWSSigningRegionName: req.Region,
	})

	if err != nil {
		return 0, err
	}

	return created.ID, nil
}

func (b *Backend) issueCertificate(hostname string) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	for attempt := 0; attempt < 12; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				slog.WarnContext(ctx, "bunny free certificate not issued within budget", "hostname", hostname)
				return
			case <-time.After(30 * time.Second):
			}
		}

		if err := b.api.IssueCertificate(ctx, hostname); err == nil {
			slog.InfoContext(ctx, "bunny free certificate issued", "hostname", hostname)
			return
		}
	}

	slog.WarnContext(ctx, "bunny free certificate not issued within budget", "hostname", hostname)
}

func (b *Backend) unpublish(ctx context.Context, req objectstorage.SetPublicRequest) error {
	var id int64

	if req.PullZoneID != "" {
		parsed, err := strconv.ParseInt(req.PullZoneID, 10, 64)

		if err != nil {
			return err
		}

		id = parsed
	} else {
		found, err := b.api.PullZoneByName(ctx, req.Slug)

		if err != nil {
			return err
		}

		if found == nil {
			return nil
		}

		id = found.ID
	}

	if err := b.api.DeletePullZone(ctx, id); err != nil && !bunny.IsNotFound(err) {
		return err
	}

	return nil
}

func (b *Backend) ensureDNSRecord(ctx context.Context, slug string) error {
	zone, err := b.dnsZone(ctx)

	if err != nil {
		return err
	}

	for _, record := range zone.Records {
		if record.Name == slug {
			return nil
		}
	}

	return b.api.CreateDNSRecord(ctx, zone.ID, bunny.DNSRecord{
		Type:  bunny.DNSRecordTypeCNAME,
		Name:  slug,
		Value: slug + ".b-cdn.net",
		TTL:   300,
	})
}

func (b *Backend) deleteDNSRecord(ctx context.Context, slug string) error {
	zone, err := b.dnsZone(ctx)

	if err != nil {
		return err
	}

	for _, record := range zone.Records {
		if record.Name != slug {
			continue
		}

		if err := b.api.DeleteDNSRecord(ctx, zone.ID, record.ID); err != nil && !bunny.IsNotFound(err) {
			return err
		}

		return nil
	}

	return nil
}

func (b *Backend) dnsZone(ctx context.Context) (*bunny.DNSZone, error) {
	zone, err := b.api.DNSZoneByDomain(ctx, b.domain)

	if err != nil {
		return nil, err
	}

	if zone == nil {
		return nil, fmt.Errorf("dns zone %q not found", b.domain)
	}

	return b.api.DNSZone(ctx, zone.ID)
}
