package vulnerabilities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityUnknown  Severity = "UNKNOWN"
)

type Source string

const (
	SourceOperatingSystem Source = "OPERATING_SYSTEM"
	SourceApplication     Source = "APPLICATION"
	SourceUnknown         Source = "UNKNOWN"
)

type Package struct {
	Name             string
	InstalledVersion string
	FixedVersion     string
	Path             string
}

type Vulnerability struct {
	ID          string
	Title       string
	Description string
	Severity    Severity
	Source      Source
	Reference   string
	Packages    []Package
}

type Summary struct {
	Critical int
	High     int
	Medium   int
	Low      int
	Unknown  int
	Total    int
}

type Report struct {
	Image           string
	Summary         Summary
	Vulnerabilities []Vulnerability
}

type Config struct {
	Endpoint     string
	DialEndpoint string
	Keychain     authn.Keychain
}

type Client struct {
	config Config
	http   *http.Client
}

func New(config Config) *Client {
	if config.DialEndpoint == "" {
		config.DialEndpoint = config.Endpoint
	}

	return &Client{
		config: config,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

const cveQuery = `query CVEsForImage($image: String!) {
  CVEListForImage(image: $image) {
    CVEList {
      Id
      Title
      Description
      Severity
      Reference
      PackageList { Name InstalledVersion FixedVersion PackagePath }
    }
    Summary {
      Count
      UnknownCount
      LowCount
      MediumCount
      HighCount
      CriticalCount
    }
  }
}`

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type graphQLResponse struct {
	Data struct {
		CVEListForImage struct {
			CVEList []struct {
				ID          string `json:"Id"`
				Title       string `json:"Title"`
				Description string `json:"Description"`
				Severity    string `json:"Severity"`
				Reference   string `json:"Reference"`
				PackageList []struct {
					Name             string `json:"Name"`
					InstalledVersion string `json:"InstalledVersion"`
					FixedVersion     string `json:"FixedVersion"`
					PackagePath      string `json:"PackagePath"`
				} `json:"PackageList"`
			} `json:"CVEList"`
			Summary struct {
				Count         int `json:"Count"`
				UnknownCount  int `json:"UnknownCount"`
				LowCount      int `json:"LowCount"`
				MediumCount   int `json:"MediumCount"`
				HighCount     int `json:"HighCount"`
				CriticalCount int `json:"CriticalCount"`
			} `json:"Summary"`
		} `json:"CVEListForImage"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// ForImage returns the vulnerability report the registry's built-in scanner holds
// for the given image reference (repository[:tag|@digest], without a host prefix),
// or nil when the image is unknown to the registry.
func (c *Client) ForImage(ctx context.Context, imageRef string) (*Report, error) {
	body, err := json.Marshal(graphQLRequest{
		Query:     cveQuery,
		Variables: map[string]any{"image": imageRef},
	})

	if err != nil {
		return nil, err
	}

	url := "http://" + c.config.DialEndpoint + "/v2/_zot/ext/search"

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	if err := c.authorize(request); err != nil {
		return nil, err
	}

	response, err := c.http.Do(request)

	if err != nil {
		return nil, fmt.Errorf("query vulnerability scanner: %w", err)
	}

	defer response.Body.Close()

	data, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))

	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vulnerability scanner returned %s: %s", response.Status, strings.TrimSpace(string(data)))
	}

	var result graphQLResponse

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse vulnerability report: %w", err)
	}

	if len(result.Errors) > 0 {
		if isUnknownImage(result.Errors[0].Message) {
			return nil, nil
		}

		return nil, fmt.Errorf("vulnerability scanner: %s", result.Errors[0].Message)
	}

	report := &Report{
		Image: imageRef,
		Summary: Summary{
			Critical: result.Data.CVEListForImage.Summary.CriticalCount,
			High:     result.Data.CVEListForImage.Summary.HighCount,
			Medium:   result.Data.CVEListForImage.Summary.MediumCount,
			Low:      result.Data.CVEListForImage.Summary.LowCount,
			Unknown:  result.Data.CVEListForImage.Summary.UnknownCount,
			Total:    result.Data.CVEListForImage.Summary.Count,
		},
	}

	for _, cve := range result.Data.CVEListForImage.CVEList {
		packages := make([]Package, 0, len(cve.PackageList))

		for _, pkg := range cve.PackageList {
			packages = append(packages, Package{
				Name:             pkg.Name,
				InstalledVersion: pkg.InstalledVersion,
				FixedVersion:     omitPlaceholder(pkg.FixedVersion),
				Path:             omitPlaceholder(pkg.PackagePath),
			})
		}

		report.Vulnerabilities = append(report.Vulnerabilities, Vulnerability{
			ID:          cve.ID,
			Title:       cve.Title,
			Description: cve.Description,
			Severity:    normalizeSeverity(cve.Severity),
			Source:      packageSource(packages),
			Reference:   cve.Reference,
			Packages:    packages,
		})
	}

	sort.SliceStable(report.Vulnerabilities, func(i, j int) bool {
		return severityRank(report.Vulnerabilities[i].Severity) > severityRank(report.Vulnerabilities[j].Severity)
	})

	return report, nil
}

func (c *Client) authorize(request *http.Request) error {
	if c.config.Keychain == nil {
		return nil
	}

	registry, err := name.NewRegistry(c.config.Endpoint)

	if err != nil {
		return err
	}

	authenticator, err := c.config.Keychain.Resolve(registry)

	if err != nil {
		return err
	}

	config, err := authenticator.Authorization()

	if err != nil {
		return err
	}

	if config.Username != "" || config.Password != "" {
		request.SetBasicAuth(config.Username, config.Password)
	}

	return nil
}

func omitPlaceholder(value string) string {
	if strings.EqualFold(value, "Not Specified") {
		return ""
	}

	return value
}

func packageSource(packages []Package) Source {
	if len(packages) == 0 {
		return SourceUnknown
	}

	for _, pkg := range packages {
		if !operatingSystemPackage(pkg) {
			return SourceApplication
		}
	}

	return SourceOperatingSystem
}

func operatingSystemPackage(pkg Package) bool {
	if pkg.Path != "" {
		return false
	}

	switch pkg.Name {
	case "stdlib", "toolchain":
		return false
	}

	return !strings.Contains(pkg.Name, "/")
}

func normalizeSeverity(severity string) Severity {
	switch strings.ToUpper(severity) {
	case "CRITICAL":
		return SeverityCritical
	case "HIGH":
		return SeverityHigh
	case "MEDIUM":
		return SeverityMedium
	case "LOW":
		return SeverityLow
	default:
		return SeverityUnknown
	}
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

func isUnknownImage(message string) bool {
	message = strings.ToLower(message)

	return strings.Contains(message, "not found") ||
		strings.Contains(message, "no such") ||
		strings.Contains(message, "unknown") ||
		strings.Contains(message, "does not exist")
}
