package conductor

import (
	"bufio"
	"context"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type Build = buildjob.Job
type BuildID = buildjob.BuildID

var _ platform.WorkspaceScoped = BuildID{}

func (c *Client) Builds(ctx context.Context, workspace, repoURL, contextPath string) ([]Build, error) {
	return c.buildjob.List(ctx, workspace, repoURL, contextPath)
}

func (c *Client) Build(ctx context.Context, id BuildID) (*Build, error) {
	return c.buildjob.Get(ctx, id)
}

func (c *Client) BuildLogs(ctx context.Context, id BuildID) (<-chan string, error) {
	reader, err := c.buildjob.Logs(ctx, id)

	if err != nil {
		return nil, err
	}

	out := make(chan string, 128)

	go func() {
		defer close(out)
		defer reader.Close()

		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			select {
			case out <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

func (c *Client) Deploy(ctx context.Context, serviceID ServiceID, gitRef string) (*Build, error) {
	service, err := c.platform.Service(ctx, serviceID)

	if err != nil {
		return nil, err
	}

	ref := gitRef

	if ref == "" {
		ref = service.Branch
	}

	commit, err := c.source.CommitSHA(ctx, service.SourceURL, ref)

	if err != nil {
		return nil, err
	}

	token, err := c.source.Token(ctx, service.SourceURL)

	if err != nil {
		return nil, err
	}

	imageName := service.ID.Workspace + "/" + service.ID.Project + "/" + service.Name

	build, err := c.buildjob.Start(ctx, buildjob.StartOptions{
		Workspace:        service.ID.Workspace,
		RepoURL:          service.SourceURL,
		Commit:           commit,
		ContextPath:      service.ContextPath,
		TargetImageNames: []string{imageName},
		Token:            token,
		BuildVars:        service.Variables,
	})

	if err != nil {
		return nil, err
	}

	claims, _ := auth.FromContext(ctx)

	go c.runDeploy(claims, service.ID, build.ID)

	return build, nil
}
