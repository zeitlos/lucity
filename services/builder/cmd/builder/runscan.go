package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
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
	Scanner     string `envconfig:"SCAN_SCANNER" required:"true"`
	SourceURL   string `envconfig:"SCAN_SOURCE_URL" required:"true"`
	Commit      string `envconfig:"SCAN_COMMIT"`
	ReportRepo  string `envconfig:"SCAN_REPORT_REPO" required:"true"`
	GitHubToken string `envconfig:"GITHUB_TOKEN"`
}

const (
	reportMediaType       = "application/vnd.lucity.scan-report.v1+json"
	reportConfigMediaType = "application/vnd.lucity.scan-report.config.v1+json"

	scanHistoryDepth = 200
)

type scanReport struct {
	Scanner   string        `json:"scanner"`
	Commit    string        `json:"commit"`
	ScannedAt time.Time     `json:"scannedAt"`
	Findings  []scanFinding `json:"findings"`
}

type scanFinding struct {
	Rule   string `json:"rule"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Commit string `json:"commit"`
	Secret string `json:"secret"`
	Author string `json:"author,omitempty"`
}

func runScan() {
	logger.Setup("info")

	var config ScanConfig

	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("scan runner starting", "scan_id", config.ScanID, "scanner", config.Scanner, "source_url", config.SourceURL)

	findings, err := executeScan(config)

	if err != nil {
		slog.Error("scan failed", "error", err)
		os.Exit(1)
	}

	writeTerminationSummary(config.Scanner, len(findings))

	if len(findings) > 0 {
		slog.Warn("scan found potential secrets", "count", len(findings))
		return
	}

	slog.Info("scan clean")
}

func executeScan(config ScanConfig) ([]scanFinding, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	repoPath, commit, err := cloneForScan(ctx, config)

	if err != nil {
		return nil, fmt.Errorf("clone failed: %w", err)
	}

	defer os.RemoveAll(repoPath)

	var findings []scanFinding

	switch config.Scanner {
	case "gitleaks":
		findings, err = runGitleaks(ctx, repoPath)
	case "trufflehog":
		findings, err = runTrufflehog(ctx, repoPath)
	default:
		return nil, fmt.Errorf("unknown scanner %q", config.Scanner)
	}

	if err != nil {
		return nil, err
	}

	report := scanReport{
		Scanner:   config.Scanner,
		Commit:    commit,
		ScannedAt: time.Now().UTC(),
		Findings:  findings,
	}

	if err := pushReport(ctx, config, report); err != nil {
		return nil, fmt.Errorf("push report: %w", err)
	}

	return findings, nil
}

func cloneForScan(ctx context.Context, config ScanConfig) (string, string, error) {
	tmpDir, err := os.MkdirTemp("", "scan-*")

	if err != nil {
		return "", "", err
	}

	cloneOpts := &git.CloneOptions{
		URL:          config.SourceURL,
		Depth:        scanHistoryDepth,
		SingleBranch: true,
	}

	if config.GitHubToken != "" {
		cloneOpts.Auth = &githttp.BasicAuth{
			Username: "x-access-token",
			Password: config.GitHubToken,
		}
	}

	repo, err := git.PlainCloneContext(ctx, tmpDir, false, cloneOpts)

	if err != nil {
		os.RemoveAll(tmpDir)
		return "", "", err
	}

	commit := config.Commit

	if commit != "" {
		worktree, err := repo.Worktree()

		if err == nil {
			if err := worktree.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(commit)}); err != nil {
				slog.Warn("checkout failed, scanning default branch head", "commit", commit, "error", err)
			}
		}
	}

	if commit == "" {
		head, err := repo.Head()

		if err != nil {
			os.RemoveAll(tmpDir)
			return "", "", err
		}

		commit = head.Hash().String()
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

func runGitleaks(ctx context.Context, repoPath string) ([]scanFinding, error) {
	reportPath := repoPath + "-gitleaks.json"
	defer os.Remove(reportPath)

	cmd := exec.CommandContext(ctx, "gitleaks", "detect",
		"--source", repoPath,
		"--report-format", "json",
		"--report-path", reportPath,
		"--exit-code", "3",
		"--no-banner",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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

	findings := make([]scanFinding, 0, len(raw))

	for _, f := range raw {
		findings = append(findings, scanFinding{
			Rule:   f.RuleID,
			File:   f.File,
			Line:   f.StartLine,
			Commit: f.Commit,
			Secret: maskSecret(f.Secret),
			Author: f.Author,
		})
	}

	return findings, nil
}

type trufflehogFinding struct {
	DetectorName   string `json:"DetectorName"`
	Raw            string `json:"Raw"`
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

func runTrufflehog(ctx context.Context, repoPath string) ([]scanFinding, error) {
	cmd := exec.CommandContext(ctx, "trufflehog", "git", "file://"+repoPath,
		"--json",
		"--no-update",
		"--fail",
		"--concurrency=1",
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	var exitErr *exec.ExitError

	if err != nil && (!errors.As(err, &exitErr) || exitErr.ExitCode() != 183) {
		return nil, fmt.Errorf("trufflehog: %w", err)
	}

	var findings []scanFinding
	scanner := bufio.NewScanner(&stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()

		var f trufflehogFinding

		if err := json.Unmarshal(line, &f); err != nil || f.DetectorName == "" {
			continue
		}

		findings = append(findings, scanFinding{
			Rule:   f.DetectorName,
			File:   f.SourceMetadata.Data.Git.File,
			Line:   f.SourceMetadata.Data.Git.Line,
			Commit: f.SourceMetadata.Data.Git.Commit,
			Secret: maskSecret(f.Raw),
			Author: f.SourceMetadata.Data.Git.Email,
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

	for _, tag := range []string{config.Scanner + "-" + shortCommit, config.Scanner + "-latest"} {
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

func writeTerminationSummary(scanner string, findings int) {
	summary, err := json.Marshal(map[string]any{"scanner": scanner, "findings": findings})

	if err != nil {
		return
	}

	if err := os.WriteFile("/dev/termination-log", summary, 0o644); err != nil {
		slog.Warn("failed to write termination summary", "error", err)
	}
}
