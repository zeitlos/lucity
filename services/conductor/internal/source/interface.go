package source

import "context"

type Commit struct {
	SHA     string
	Message string
}

type Interface interface {
	Commit(ctx context.Context, repoURL, ref string) (Commit, error)
	Token(ctx context.Context, repoURL string) (string, error)
}
