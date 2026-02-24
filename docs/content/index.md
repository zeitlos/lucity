---
title: "Lucity — The PaaS you can leave"
description: "Open-source PaaS on Kubernetes. Git push to deploy, eject to standard Helm charts and ArgoCD configs at any time. Self-hostable, no lock-in."
---

::u-page-hero
#title
The PaaS you can leave.

#description
Open-source PaaS on Kubernetes. Git push to deploy, environments out of the box, and a real exit door — `lucity eject` gives you standard Helm charts and ArgoCD configs.

#links
  :::u-button
  ---
  to: /getting-started/quick-start
  color: neutral
  size: xl
  trailing-icon: i-lucide-arrow-right
  ---
  Get Started
  :::

  :::u-button
  ---
  to: https://github.com/zeitlos/lucity
  target: _blank
  variant: outline
  icon: i-simple-icons-github
  size: xl
  color: neutral
  ---
  View on GitHub
  :::
::

::u-page-section
#title
What you can do

#features
  :::u-page-feature
  ---
  icon: i-lucide-rocket
  to: /features/builds
  title: Push to deploy
  description: "Connect your GitHub repo. Lucity detects your framework, builds a container image, and deploys it. Zero Dockerfiles required."
  ---
  :::

  :::u-page-feature
  ---
  icon: i-lucide-layers
  to: /features/environments
  title: Multi-environment
  description: "Development, staging, production, and PR preview environments. Promote images between them without rebuilding."
  ---
  :::

  :::u-page-feature
  ---
  icon: i-lucide-door-open
  to: /features/eject
  title: Eject anytime
  description: "Export your entire setup as standard Helm charts and ArgoCD configs. No lock-in. Your infrastructure, always."
  ---
  :::

  :::u-page-feature
  ---
  icon: i-lucide-database
  to: /infrastructure/databases
  title: Batteries included
  description: "PostgreSQL via CloudNativePG, Redis, cron jobs, and HTTP routing via Gateway API. Everything your app needs."
  ---
  :::

  :::u-page-feature
  ---
  icon: i-lucide-git-branch
  to: /architecture/gitops
  title: GitOps native
  description: "Every deployment is a Git commit. ArgoCD syncs your workloads. Full audit trail, real rollbacks."
  ---
  :::

  :::u-page-feature
  ---
  icon: i-lucide-heart
  to: https://github.com/zeitlos/lucity
  title: Open source
  description: "AGPL-3.0 licensed. Self-host on your own Kubernetes cluster. Built on ArgoCD, Helm, CloudNativePG, and friends."
  ---
  :::
::
