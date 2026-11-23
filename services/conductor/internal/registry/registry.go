package registry

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/zeitlos/lucity/services/conductor/internal/publicnet"
)

const (
	defaultCacheTTL = 10 * time.Minute
	errorCacheTTL   = time.Minute
	fetchTimeout    = 30 * time.Second
	maxConfigSize   = 1 << 20
	maxCacheEntries = 256
)

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
	public       http.RoundTripper

	mu    sync.Mutex
	cache map[string]entry
}

type entry struct {
	config  *v1.Config
	err     error
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
		public: &http.Transport{
			Proxy:                 nil,
			DialContext:           publicnet.Dialer(5 * time.Second).DialContext,
			ForceAttemptHTTP2:     true,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
		},
		cache: make(map[string]entry),
	}
}

func (c *Client) ImageConfig(ctx context.Context, imageRef string) (*v1.Config, error) {
	if e, ok := c.cached(imageRef); ok {
		return e.config, e.err
	}

	cfg, err := c.fetch(ctx, imageRef)
	c.store(imageRef, cfg, err)
	return cfg, err
}

func (c *Client) cached(ref string) (entry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.cache[ref]
	if !ok || time.Now().After(e.expires) {
		return entry{}, false
	}
	return e, true
}

func (c *Client) store(ref string, cfg *v1.Config, err error) {
	ttl := c.ttl
	if err != nil {
		ttl = errorCacheTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	for key, e := range c.cache {
		if now.After(e.expires) {
			delete(c.cache, key)
		}
	}

	if len(c.cache) >= maxCacheEntries {
		c.cache = make(map[string]entry)
	}

	c.cache[ref] = entry{config: cfg, err: err, expires: now.Add(ttl)}
}

func (c *Client) fetch(ctx context.Context, imageRef string) (*v1.Config, error) {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, err
	}

	registry := ref.Context().RegistryStr()
	auth := authn.Anonymous
	transport := c.public
	var options []name.Option

	if c.endpoint != "" && strings.EqualFold(registry, c.endpoint) {
		registry = c.dialEndpoint
		transport = remote.DefaultTransport
		options = append(options, name.Insecure)

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

	repo, err := name.NewRepository(registry+"/"+ref.Context().RepositoryStr(), options...)
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
		remote.WithTransport(transport),
	)
	if err != nil {
		return nil, err
	}

	manifest, err := img.Manifest()
	if err != nil {
		return nil, err
	}

	if manifest.Config.Size > maxConfigSize {
		return nil, fmt.Errorf("image config is %d bytes, more than %d", manifest.Config.Size, maxConfigSize)
	}

	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, err
	}

	return &v1.Config{
		Entrypoint:   cfg.Config.Entrypoint,
		Cmd:          cfg.Config.Cmd,
		ExposedPorts: cfg.Config.ExposedPorts,
	}, nil
}
