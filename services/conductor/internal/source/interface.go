package source

import "context"

type Interface interface {
	CommitSHA(ctx context.Context, repoURL, ref string) (string, error)
	Token(ctx context.Context, repoURL string) (string, error)
}
