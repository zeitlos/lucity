// Package deploy holds the in-memory Tracker that follows individual
// deploy runs (build → push → apply → rollout) so the dashboard can
// surface live progress for a triggered deploy.
//
// This is distinct from the deployer package: the Backend in
// internal/deployer is concerned with cluster state; this package is
// concerned with the user-visible journey of a single deploy.
//
// Populated in phase 3 from services/gateway/deploy/.
package deploy
