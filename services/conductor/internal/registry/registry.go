package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const defaultCacheTTL = 10 * time.Minute

type Config struct {
	Endpoint     string
	DialEndpoint string
	Keychain     authn.Keychain
	CacheTTL     time.Duration
}

type Client struct {
	endpoint     string
	dialEndpoint string
	keychain     authn.Keychain
	ttl          time.Duration

	mu    sync.Mutex
	cache map[string]entry
}

type entry struct {
	config  *v1.Config
	expires time.Time
}

func New(cfg Config) *Client {
	ttl := cfg.CacheTTL
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}

	dialEndpoint := cfg.DialEndpoint
	if dialEndpoint == "" {
		dialEndpoint = cfg.Endpoint
	}

	return &Client{
		endpoint:     cfg.Endpoint,
		dialEndpoint: dialEndpoint,
		keychain:     cfg.Keychain,
		ttl:          ttl,
		cache:        make(map[string]entry),
	}
}

func (c *Client) ImageConfig(ctx context.Context, imageRef string) (*v1.Config, error) {
	if cfg, ok := c.cached(imageRef); ok {
		return cfg, nil
	}

	cfg, err := c.fetch(ctx, imageRef)
	if err != nil {
		return nil, err
	}

	c.store(imageRef, cfg)
	return cfg, nil
}

func (c *Client) cached(ref string) (*v1.Config, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.cache[ref]
	if !ok || time.Now().After(e.expires) {
		return nil, false
	}
	return e.config, true
}

func (c *Client) store(ref string, cfg *v1.Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[ref] = entry{config: cfg, expires: time.Now().Add(c.ttl)}
}

func (c *Client) fetch(ctx context.Context, imageRef string) (*v1.Config, error) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, err
	}

	registry := ref.Context().RegistryStr()
	auth := authn.Anonymous

	if c.endpoint != "" && strings.EqualFold(registry, c.endpoint) {
		registry = c.dialEndpoint

		endpoint, err := name.NewRegistry(c.endpoint)
		if err != nil {
			return nil, err
		}

		if c.keychain != nil {
			auth, err = c.keychain.Resolve(endpoint)
			if err != nil {
				return nil, err
			}
		}
	}

	repo, err := name.NewRepository(registry + "/" + ref.Context().RepositoryStr())
	if err != nil {
		return nil, err
	}

	switch r := ref.(type) {
	case name.Tag:
		ref = repo.Tag(r.TagStr())
	case name.Digest:
		ref = repo.Digest(r.DigestStr())
	default:
		return nil, fmt.Errorf("unexpected reference type %T", ref)
	}

	img, err := remote.Image(ref,
		remote.WithContext(ctx),
		remote.WithAuth(auth),
	)
	if err != nil {
		return nil, err
	}

	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, err
	}

	return &cfg.Config, nil
}
