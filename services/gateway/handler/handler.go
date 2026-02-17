package handler

import (
	"github.com/zeitlos/lucity/pkg/packager"
)

// Client holds all dependencies for the gateway's business logic.
type Client struct {
	Packager packager.PackagerServiceClient
}

func New(packagerClient packager.PackagerServiceClient) *Client {
	return &Client{
		Packager: packagerClient,
	}
}
