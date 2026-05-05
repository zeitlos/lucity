package softserve

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer/argo/gitops"
	"golang.org/x/crypto/ssh"
)

const repoSuffix = "-gitops"
const sshUser = "admin"

type Forge struct {
	sshAddr string
	sshKey  ssh.Signer

	httpAddr  string
	httpToken string
}

// New initializes a new client to interacts with the softserve git forge.
func New(sshAddr string, sshKey ssh.Signer, httpAddr, httpToken string) *Forge {
	return &Forge{
		sshAddr:   sshAddr,
		sshKey:    sshKey,
		httpAddr:  httpAddr,
		httpToken: httpToken,
	}
}

// CreateRepo creates a GitOps repo on Soft-serve and populates it.
func (f *Forge) CreateRepo(ctx context.Context, name, workspace, displayName string) (string, error) {
	slug := buildRepoSlug(name, workspace)
	cloneURL := f.repoHTTPURL(slug)

	// Create the repo via SSH (idempotent: handle "already exists")
	if _, err := f.sshCmd("repo", "create", slug); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return "", fmt.Errorf("failed to create repo %s: %w", slug, err)
		}

		slog.Info("repo already exists, checking state", "repo", slug)
	}

	// Make it private (idempotent)
	if _, err := f.sshCmd("repo", "private", slug, "true"); err != nil {
		slog.Warn("failed to set repo private", "repo", slug, "error", err)
	}

	// Set display name as Soft-serve project-name metadata
	if displayName != "" {
		if _, err := f.sshCmd("repo", "project-name", slug, displayName); err != nil {
			slog.Warn("failed to set repo project-name", "repo", slug, "error", err)
		}
	}

	repo, err := cloneRepo(f.repoHTTPURL(slug), f.httpToken)

	if err != nil {
		return "", fmt.Errorf("failed to initialize repo %s: %w", slug, err)
	}
	repo.SetWorkspace(workspace)

	if repo.isInitialized() {
		slog.Info("repo already initialized", "repo", slug)
		return cloneURL, nil
	}

	slog.Info("initializing softserve repo", "repo", slug, "url", cloneURL)

	if err := repo.initialize(ctx); err != nil {
		return "", fmt.Errorf("failed to initialize repo contents: %w", err)
	}

	return cloneURL, nil
}

// Repos lists all GitOps repos on Soft-serve.
func (f *Forge) Repos(ctx context.Context, workspace string) ([]gitops.RepositoryEntry, error) {
	output, err := f.sshCmd("repo", "list")

	if err != nil {
		return nil, fmt.Errorf("failed to list repos: %w", err)
	}

	var result []gitops.RepositoryEntry

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		slug := strings.TrimSpace(line)

		if slug == "" || !strings.HasSuffix(slug, repoSuffix) {
			continue
		}

		if workspace != "" && !strings.HasPrefix(slug, workspace+"-") {
			continue
		}

		entry := gitops.RepositoryEntry{
			Slug:    slug,
			HTTPURL: f.repoHTTPURL(slug),
		}

		displayName, err := f.sshCmd("repo", "project-name", slug)

		if err == nil {
			entry.DisplayName = strings.TrimSpace(displayName)
		} else {
			slog.Warn("failed to fetch display name", "repo", slug, "error", err)
		}

		result = append(result, entry)
	}

	return result, nil
}

// Repo reads a single project's metadata.
func (f *Forge) Repo(ctx context.Context, name, workspace string) (*gitops.RepositoryEntry, error) {
	slug := buildRepoSlug(name, workspace)

	if !strings.HasSuffix(slug, repoSuffix) {
		return nil, fmt.Errorf("invalid repo slug: %s", slug)
	}

	displayName, err := f.sshCmd("repo", "project-name", slug)

	if err != nil {
		return nil, fmt.Errorf("failed to get repo display name: %w", err)
	}

	return &gitops.RepositoryEntry{
		Slug:        slug,
		HTTPURL:     f.repoHTTPURL(slug),
		DisplayName: strings.TrimSpace(displayName),
	}, nil
}

// DeleteRepo removes a repo from Soft-serve.
func (f *Forge) DeleteRepo(ctx context.Context, name, workspace string) error {
	slug := buildRepoSlug(name, workspace)

	if _, err := f.sshCmd("repo", "delete", slug); err != nil {
		return fmt.Errorf("failed to delete repo %s: %w", slug, err)
	}

	slog.Info("deleted softserve repo", "repo", slug)
	return nil
}

func (f *Forge) CloneRepo(ctx context.Context, httpURL string) (gitops.Repository, error) {
	return cloneRepo(httpURL, f.httpToken)
}

// sshCmd executes a command on the Soft-serve SSH server.
func (f *Forge) sshCmd(args ...string) (string, error) {
	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(f.sshKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", f.sshAddr, sshConfig)
	if err != nil {
		return "", fmt.Errorf("failed to connect to soft-serve: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create ssh session: %w", err)
	}
	defer session.Close()

	cmd := strings.Join(args, " ")
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("ssh command %q failed: %w (stderr: %s)", cmd, err, stderr.String())
	}

	return stdout.String(), nil
}

// repoHTTPURL returns the HTTP clone URL for a repo.
func (f *Forge) repoHTTPURL(slug string) string {
	return strings.TrimSuffix(f.httpAddr, "/") + "/" + slug + ".git"
}

func buildRepoSlug(name, workspace string) string {
	return workspace + "-" + name + repoSuffix
}

var _ gitops.Forge = (*Forge)(nil)
