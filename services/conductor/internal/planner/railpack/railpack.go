package railpack

import (
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
)

type Client struct{}

func New() *Client {
	return &Client{}
}

var _ planner.Interface = (*Client)(nil)
