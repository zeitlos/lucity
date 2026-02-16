package handler

// Client holds all dependencies for the gateway's business logic.
// In this phase, it serves mock data. In future phases, it will hold
// gRPC client connections to builder, packager, and deployer.
type Client struct{}

func New() *Client {
	return &Client{}
}
