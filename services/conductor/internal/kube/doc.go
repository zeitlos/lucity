// Package kube holds small Kubernetes API helpers shared across the
// conductor's internal packages: namespace lifecycle, label
// application, workspace-label verification.
//
// This is intentionally not a wrapper around client-go — it is a thin
// set of functions that takes a typed client and does one specific
// thing. Higher layers compose these helpers; the deployer.Backend
// implementations may also use them directly.
//
// Populated in phase 2.
package kube
