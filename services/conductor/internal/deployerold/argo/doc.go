// Package argo implements deployer.Backend on top of Soft-serve
// (GitOps repo storage) and ArgoCD (sync + reconciliation).
//
// Apply writes per-environment values into a project's GitOps repo on
// Soft-serve, commits, and either triggers an ArgoCD sync or relies on
// auto-sync to converge the cluster. Status aggregates ArgoCD's sync
// state with K8s Pod/Deployment health into the abstract Status shape
// the Backend interface returns.
//
// This implementation will be wired up in phase 2 of the conductor
// refactor by moving services/packager/{chart,gitops,eject} and
// services/deployer/{argocd,environment,database} into its subtree.
package argo
