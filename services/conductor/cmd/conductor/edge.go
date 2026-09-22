package main

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/bunny"
	helmDeployer "github.com/zeitlos/lucity/services/conductor/internal/deployer/helm"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/edge"
)

func newEdge(bunnyClient *bunny.Client, config Config) *edge.Client {
	return edge.New(bunnyClient, edge.Config{
		OriginURL:    config.EdgeOriginURL,
		HeaderSecret: config.EdgeHeaderSecret,
		ZonePrefix:   strings.ReplaceAll(config.WorkloadDomain, ".", "-"),
	})
}

func edgeHeaderOptions(enforced bool) []helmDeployer.Option {
	if !enforced {
		return nil
	}

	return []helmDeployer.Option{helmDeployer.WithHeaderMatches(values.HeaderMatch{
		Name:      edge.HeaderName,
		ValueFrom: &values.SecretKeyRef{SecretName: edge.HeaderSecretName, Key: edge.HeaderSecretKey},
	})}
}

func tlsCertificate(ctx context.Context, k8s kubernetes.Interface, namespace, name string) (*edge.Certificate, error) {
	secret, err := k8s.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return &edge.Certificate{Certificate: secret.Data["tls.crt"], Key: secret.Data["tls.key"]}, nil
}
