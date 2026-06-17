// Package grpc serves the conductor's external gRPC surface. After
// the merge, the only remaining external caller is cashier, and the
// only remaining RPC is SuspendWorkspace.
//
// Internal-JWT verification (ES256) is preserved unchanged from the
// previous deployer service so cashier's existing auth wiring keeps
// working.
//
// Populated in phase 3.
package grpc
