package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	gcrtypes "github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/kelseyhightower/envconfig"

	"github.com/zeitlos/lucity/pkg/logger"
)

type ScanConfig struct {
	ScanID      string `envconfig:"SCAN_ID" required:"true"`
	SourceURL   string `envconfig:"SCAN_SOURCE_URL" required:"true"`
	Commit      string `envconfig:"SCAN_COMMIT"`
	ReportRepo  string `envconfig:"SCAN_REPORT_REPO" required:"true"`
	GitHubToken string `envconfig:"GITHUB_TOKEN"`
}

const (
	reportMediaType       = "application/vnd.lucity.scan-report.v1+json"
	reportConfigMediaType = "application/vnd.lucity.scan-report.config.v1+json"

	reportTagPrefix = "secrets"
)

type scanReport struct {
	Commit    string        `json:"commit"`
	ScannedAt time.Time     `json:"scannedAt"`
	Findings  []scanFinding `json:"findings"`
}

type scanFinding struct {
	Rule     string `json:"rule"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Commit   string `json:"commit"`
	Secret   string `json:"secret"`
	Author   string `json:"author,omitempty"`
	Verified bool   `json:"verified"`
}

type rawFinding struct {
	scanFinding
	raw string
}

func runScan() {
	logger.Setup("info")

	var config ScanConfig

	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("scan runner starting", "scan_id", config.ScanID, "source_url", config.SourceURL)

	findings, err := executeScan(config)

	if err != nil {
		slog.Error("scan failed", "error", err)
		os.Exit(1)
	}

	writeTerminationSummary(findings)

	if len(findings) > 0 {
		slog.Warn("scan found potential secrets", "count", len(findings))
		return
	}

	slog.Info("scan clean")
}

func executeScan(config ScanConfig) ([]scanFinding, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	repoPath, commit, err := cloneForScan(ctx, config)

	if err != nil {
		return nil, fmt.Errorf("clone failed: %w", err)
	}

	defer os.RemoveAll(repoPath)

	gitleaksFindings, err := runGitleaks(ctx, repoPath)

	if err != nil {
		return nil, err
	}

	trufflehogFindings, err := runTrufflehog(ctx, repoPath)

	if err != nil {
		return nil, err
	}

	findings := mergeFindings(gitleaksFindings, trufflehogFindings)

	report := scanReport{
		Commit:    commit,
		ScannedAt: time.Now().UTC(),
		Findings:  findings,
	}

	if err := pushReport(ctx, config, report); err != nil {
		return nil, fmt.Errorf("push report: %w", err)
	}

	return findings, nil
}

func mergeFindings(sets ...[]rawFinding) []scanFinding {
	index := map[string]int{}
	merged := []scanFinding{}

	for _, set := range sets {
		for _, finding := range set {
			key := finding.File + "|" + strconv.Itoa(finding.Line) + "|" + finding.raw

			if at, ok := index[key]; ok {
				if finding.Verified {
					merged[at].Verified = true
				}

				continue
			}

			index[key] = len(merged)
			finding.Secret = maskSecret(finding.raw)
			merged = append(merged, finding.scanFinding)
		}
	}

	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].Verified && !merged[j].Verified
	})

	return merged
}

func cloneForScan(ctx context.Context, config ScanConfig) (string, string, error) {
	tmpDir, err := os.MkdirTemp("", "scan-*")

	if err != nil {
		return "", "", err
	}

	env := os.Environ()

	if config.GitHubToken != "" {
		header := "Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte("x-access-token:"+config.GitHubToken))
		env = append(env,
			"GIT_CONFIG_COUNT=1",
			"GIT_CONFIG_KEY_0=http.extraHeader",
			"GIT_CONFIG_VALUE_0="+header,
			"GIT_TERMINAL_PROMPT=0",
		)
	}

	clone := exec.CommandContext(ctx, "git", "clone", "--single-branch", "--no-tags", config.SourceURL, tmpDir)
	clone.Env = env
	clone.Stderr = os.Stderr

	if err := clone.Run(); err != nil {
		os.RemoveAll(tmpDir)
		return "", "", fmt.Errorf("git clone: %w", err)
	}

	commit := config.Commit

	if commit != "" {
		checkout := exec.CommandContext(ctx, "git", "-C", tmpDir, "checkout", "--quiet", commit)
		checkout.Env = env
		checkout.Stderr = os.Stderr

		if err := checkout.Run(); err != nil {
			slog.Warn("checkout failed, scanning default branch head", "commit", commit, "error", err)
			commit = ""
		}
	}

	if commit == "" {
		out, err := exec.CommandContext(ctx, "git", "-C", tmpDir, "rev-parse", "HEAD").Output()

		if err != nil {
			os.RemoveAll(tmpDir)
			return "", "", fmt.Errorf("resolve HEAD: %w", err)
		}

		commit = strings.TrimSpace(string(out))
	}

	return tmpDir, commit, nil
}

type gitleaksFinding struct {
	RuleID    string `json:"RuleID"`
	File      string `json:"File"`
	StartLine int    `json:"StartLine"`
	Commit    string `json:"Commit"`
	Secret    string `json:"Secret"`
	Author    string `json:"Author"`
}

func runGitleaks(ctx context.Context, repoPath string) ([]rawFinding, error) {
	reportPath := repoPath + "-gitleaks.json"
	defer os.Remove(reportPath)

	cmd := exec.CommandContext(ctx, "gitleaks", "detect",
		"--source", repoPath,
		"--report-format", "json",
		"--report-path", reportPath,
		"--exit-code", "3",
		"--max-target-megabytes", "25",
		"--no-banner",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GOMAXPROCS=3")

	err := cmd.Run()

	var exitErr *exec.ExitError

	if err != nil && (!errors.As(err, &exitErr) || exitErr.ExitCode() != 3) {
		return nil, fmt.Errorf("gitleaks: %w", err)
	}

	data, err := os.ReadFile(reportPath)

	if err != nil {
		return nil, fmt.Errorf("read gitleaks report: %w", err)
	}

	var raw []gitleaksFinding

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse gitleaks report: %w", err)
	}

	findings := make([]rawFinding, 0, len(raw))

	for _, f := range raw {
		findings = append(findings, rawFinding{
			scanFinding: scanFinding{
				Rule:   f.RuleID,
				File:   f.File,
				Line:   f.StartLine,
				Commit: f.Commit,
				Author: f.Author,
			},
			raw: f.Secret,
		})
	}

	return findings, nil
}

type trufflehogFinding struct {
	DetectorName   string `json:"DetectorName"`
	Raw            string `json:"Raw"`
	Verified       bool   `json:"Verified"`
	SourceMetadata struct {
		Data struct {
			Git struct {
				Commit string `json:"commit"`
				File   string `json:"file"`
				Line   int    `json:"line"`
				Email  string `json:"email"`
			} `json:"Git"`
		} `json:"Data"`
	} `json:"SourceMetadata"`
}

func runTrufflehog(ctx context.Context, repoPath string) ([]rawFinding, error) {
	cmd := exec.CommandContext(ctx, "trufflehog", "git", "file://"+repoPath,
		"--json",
		"--no-update",
		"--fail",
		"--concurrency=2",
		"--results=verified,unknown",
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	var exitErr *exec.ExitError

	if err != nil && (!errors.As(err, &exitErr) || exitErr.ExitCode() != 183) {
		return nil, fmt.Errorf("trufflehog: %w", err)
	}

	var findings []rawFinding
	scanner := bufio.NewScanner(&stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()

		var f trufflehogFinding

		if err := json.Unmarshal(line, &f); err != nil || f.DetectorName == "" {
			continue
		}

		findings = append(findings, rawFinding{
			scanFinding: scanFinding{
				Rule:     f.DetectorName,
				File:     f.SourceMetadata.Data.Git.File,
				Line:     f.SourceMetadata.Data.Git.Line,
				Commit:   f.SourceMetadata.Data.Git.Commit,
				Author:   f.SourceMetadata.Data.Git.Email,
				Verified: f.Verified,
			},
			raw: f.Raw,
		})
	}

	return findings, nil
}

func maskSecret(secret string) string {
	secret = strings.TrimSpace(secret)

	if len(secret) <= 8 {
		return "••••••••"
	}

	return secret[:4] + strings.Repeat("•", 8) + fmt.Sprintf(" (%d chars)", len(secret))
}

func pushReport(ctx context.Context, config ScanConfig, report scanReport) error {
	data, err := json.Marshal(report)

	if err != nil {
		return err
	}

	layer := static.NewLayer(data, gcrtypes.MediaType(reportMediaType))
	img := mutate.MediaType(empty.Image, gcrtypes.OCIManifestSchema1)
	img = mutate.ConfigMediaType(img, gcrtypes.MediaType(reportConfigMediaType))
	img, err = mutate.AppendLayers(img, layer)

	if err != nil {
		return err
	}

	shortCommit := report.Commit

	if len(shortCommit) > 7 {
		shortCommit = shortCommit[:7]
	}

	for _, tag := range []string{reportTagPrefix + "-" + shortCommit, reportTagPrefix + "-latest"} {
		ref, err := name.ParseReference(config.ReportRepo+":"+tag, name.Insecure)

		if err != nil {
			return err
		}

		if err := remote.Write(ref, img, remote.WithContext(ctx), remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
			return fmt.Errorf("push %s: %w", ref.String(), err)
		}
	}

	slog.Info("scan report pushed", "repo", config.ReportRepo, "findings", len(report.Findings))

	return nil
}

func writeTerminationSummary(findings []scanFinding) {
	verified := 0

	for _, finding := range findings {
		if finding.Verified {
			verified++
		}
	}

	summary, err := json.Marshal(map[string]any{"findings": len(findings), "verified": verified})

	if err != nil {
		return
	}

	if err := os.WriteFile("/dev/termination-log", summary, 0o644); err != nil {
		slog.Warn("failed to write termination summary", "error", err)
	}
}
