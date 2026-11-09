package cockpit

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ChartVersion is how many environments run a version of the app chart. Since
// is when the oldest of their releases was deployed.
type ChartVersion struct {
	Version  string
	Releases int
	Since    time.Time
}

const (
	deployedAppReleases = "owner=helm,name=lucity-app,status=deployed"

	// chartVersionLookups is how many unknown releases are read one by one
	// before a paged list of every release is cheaper.
	chartVersionLookups = 50
)

// AppChartVersions returns the versions of the app chart that the deployed
// environment releases run, the most recently introduced first. A release
// revision's version never changes, so each one is read once and remembered.
func (c *Client) AppChartVersions(ctx context.Context) ([]ChartVersion, error) {
	list, err := c.objects.Resource(core.SchemeGroupVersion.WithResource("secrets")).Namespace("").List(ctx, meta.ListOptions{LabelSelector: deployedAppReleases})

	if err != nil {
		return nil, fmt.Errorf("list app releases: %w", err)
	}

	c.mu.Lock()
	var unknown []meta.PartialObjectMetadata

	for _, item := range list.Items {
		if _, ok := c.chartVersions[item.Namespace+"/"+item.Name]; !ok {
			unknown = append(unknown, item)
		}
	}
	c.mu.Unlock()

	if len(unknown) > chartVersionLookups {
		err = c.readAllChartVersions(ctx)
	} else {
		err = c.readChartVersions(ctx, unknown)
	}

	if err != nil {
		return nil, err
	}

	versions := map[string]*ChartVersion{}
	current := map[string]bool{}

	c.mu.Lock()

	for _, item := range list.Items {
		current[item.Namespace+"/"+item.Name] = true
		version, ok := c.chartVersions[item.Namespace+"/"+item.Name]

		if !ok {
			continue
		}

		if versions[version] == nil {
			versions[version] = &ChartVersion{Version: version, Since: item.CreationTimestamp.Time}
		}

		versions[version].Releases++

		if item.CreationTimestamp.Time.Before(versions[version].Since) {
			versions[version].Since = item.CreationTimestamp.Time
		}
	}

	for key := range c.chartVersions {
		if !current[key] {
			delete(c.chartVersions, key)
		}
	}

	c.mu.Unlock()

	result := make([]ChartVersion, 0, len(versions))

	for _, version := range versions {
		result = append(result, *version)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Since.After(result[j].Since)
	})

	return result, nil
}

func (c *Client) readChartVersions(ctx context.Context, items []meta.PartialObjectMetadata) error {
	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(8)

	for _, item := range items {
		group.Go(func() error {
			secret, err := c.kubernetes.CoreV1().Secrets(item.Namespace).Get(ctx, item.Name, meta.GetOptions{})

			if err != nil {
				return fmt.Errorf("read release %s/%s: %w", item.Namespace, item.Name, err)
			}

			return c.rememberChartVersion(*secret)
		})
	}

	return group.Wait()
}

func (c *Client) readAllChartVersions(ctx context.Context) error {
	options := meta.ListOptions{LabelSelector: deployedAppReleases, Limit: 100}

	for {
		secrets, err := c.kubernetes.CoreV1().Secrets("").List(ctx, options)

		if err != nil {
			return fmt.Errorf("list app releases: %w", err)
		}

		for _, secret := range secrets.Items {
			if err := c.rememberChartVersion(secret); err != nil {
				return err
			}
		}

		if secrets.Continue == "" {
			return nil
		}

		options.Continue = secrets.Continue
	}
}

func (c *Client) rememberChartVersion(secret core.Secret) error {
	version, err := chartVersion(secret.Data["release"])

	if err != nil {
		return fmt.Errorf("read release %s/%s: %w", secret.Namespace, secret.Name, err)
	}

	c.mu.Lock()
	c.chartVersions[secret.Namespace+"/"+secret.Name] = version
	c.mu.Unlock()

	return nil
}

// chartVersion reads the chart version from a release as Helm stores it:
// gzipped JSON in base64.
func chartVersion(data []byte) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(string(data))

	if err != nil {
		return "", err
	}

	var reader io.Reader = bytes.NewReader(raw)

	if bytes.HasPrefix(raw, []byte{0x1f, 0x8b, 0x08}) {
		unzipped, err := gzip.NewReader(reader)

		if err != nil {
			return "", err
		}

		defer unzipped.Close()

		reader = unzipped
	}

	var release struct {
		Chart struct {
			Metadata struct {
				Version string `json:"version"`
			} `json:"metadata"`
		} `json:"chart"`
	}

	if err := json.NewDecoder(reader).Decode(&release); err != nil {
		return "", err
	}

	return release.Chart.Metadata.Version, nil
}
