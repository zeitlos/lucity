package kubernetes

import (
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Client struct {
	kubernetes kubernetes.Interface
	dynamic    dynamic.Interface
}

func New(kubernetes kubernetes.Interface, dynamic dynamic.Interface) (*Client, error) {
	return &Client{}, nil
}

var _ platform.Interface = (*Client)(nil)
