package main

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/bunny"
	"github.com/zeitlos/lucity/services/conductor/internal/edge"
)

func newEdge(bunnyClient *bunny.Client, config Config) *edge.Client {
	return edge.New(bunnyClient, edge.Config{
		OriginURL:    config.EdgeOriginURL,
		HeaderSecret: config.EdgeHeaderSecret,
		ZonePrefix:   strings.ReplaceAll(config.WorkloadDomain, ".", "-"),
	})
}

func tlsCertificate(ctx context.Context, k8s kubernetes.Interface, namespace, name string) (*edge.Certificate, error) {
	secret, err := k8s.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return &edge.Certificate{Certificate: secret.Data["tls.crt"], Key: secret.Data["tls.key"]}, nil
}
