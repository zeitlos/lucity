// Package customdomain runs the periodic reconciliation loop that
// turns user-declared custom domains into Certificates and HTTPRoutes
// in the cluster. It is unrelated to the deployer.Backend abstraction
// — this is platform-side state that follows user intent rather than
// being applied through the deployment pipeline.
//
// Populated in phase 2 from services/deployer/grpc/custom_domain.go.
package customdomain
