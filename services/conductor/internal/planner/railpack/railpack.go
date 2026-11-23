package railpack

import (
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
)

const concurrentPlans = 2

type Client struct {
	slots chan struct{}
}

func New() *Client {
	return &Client{slots: make(chan struct{}, concurrentPlans)}
}

var _ planner.Interface = (*Client)(nil)
