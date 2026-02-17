package handler

import (
	gh "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/packager"
)

// Client holds all dependencies for the gateway's business logic.
type Client struct {
	GitHubApp *gh.App
	Packager  packager.PackagerServiceClient
}

func New(githubApp *gh.App, packagerClient packager.PackagerServiceClient) *Client {
	return &Client{
		GitHubApp: githubApp,
		Packager:  packagerClient,
	}
}
