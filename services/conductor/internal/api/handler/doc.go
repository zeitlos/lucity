// Package handler is the composition layer: it knows Lucity
// vocabulary (workspaces, projects, environments, services) and
// composes the atomic mechanism packages (deployer, builder, chart,
// kube) to implement business logic.
//
// Handler methods take typed values (WorkspaceID, ProjectID, etc.)
// as explicit arguments — workspace context is never read from
// context.Context below this layer. The compiler enforces that all
// downstream calls have the values they need.
//
// Populated in phase 3 from services/gateway/handler/, with signature
// changes to make workspace explicit and to consume in-process
// interfaces instead of gRPC clients.
package handler
