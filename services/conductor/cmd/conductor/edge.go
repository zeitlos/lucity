package main

import (
	"strings"

	"github.com/zeitlos/lucity/pkg/bunny"
	"github.com/zeitlos/lucity/pkg/edge"
)

func newEdge(bunnyClient *bunny.Client, config Config) *edge.Client {
	return edge.New(bunnyClient, edge.Config{
		OriginURL:    config.EdgeOriginURL,
		HeaderSecret: config.EdgeHeaderSecret,
		ZonePrefix:   strings.ReplaceAll(config.WorkloadDomain, ".", "-"),
	})
}
