package planner

import "context"

type Interface interface {
	Plan(ctx context.Context, repoURL, ref, token string) ([]Plan, error)
}

type Plan struct {
	Name         string
	ContextPath  string
	Providers    []string          // ["node", "next"] — language + framework
	Versions     map[string]string // {"node": "20"}
	StartCommand string
}
