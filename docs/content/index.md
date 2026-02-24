---
title: "Lucity: The PaaS you can leave"
description: "Open-source PaaS on Kubernetes. Git push to deploy, eject to standard Helm charts and ArgoCD configs at any time. Self-hostable, no lock-in."
---

::u-page-hero
#title
The PaaS you can leave.

#description
Open-source PaaS on Kubernetes. Git push to deploy, environments out of the box, and a real exit door. `lucity eject` gives you standard Helm charts and ArgoCD configs.

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

:hero-demo

::u-page-section
#title
What you can do

  :::u-page-grid
    ::::u-page-card{spotlight class="col-span-2 bento-card-deploy" variant="subtle"}
    :bento-deploy

    #title
    Push to deploy

    #description
    Connect your GitHub repo. Push your code, watch it flow through the pipeline and land on Kubernetes. Zero Dockerfiles required.
    ::::

    ::::u-page-card{spotlight class="col-span-2 lg:col-span-1 bento-card-envs" variant="subtle"}
    :bento-environments

    #title
    Multi-environment

    #description
    Dev, staging, production, and PR previews. Clone environments in seconds. Promote images without rebuilding.
    ::::

    ::::u-page-card{spotlight class="col-span-2 lg:col-span-1 bento-card-batteries" variant="subtle"}
    :bento-batteries

    #title
    Batteries included

    #description
    PostgreSQL via CloudNativePG, Redis, cron jobs, and HTTP routing via Gateway API. Everything your app needs.
    ::::

    ::::u-page-card{spotlight class="col-span-2 bento-card-eject" variant="subtle"}
    :bento-eject

    #title
    Eject anytime

    #description
    One command. Standard Helm charts, ArgoCD configs, environment values, and a README. Your infrastructure is yours.
    ::::

    ::::u-page-card{spotlight class="col-span-2 lg:col-span-1 bento-card-gitops" variant="subtle"}
    :bento-git-ops

    #title
    GitOps native

    #description
    How the big players do it, just cleverly automated. Every deploy is a Git commit. ArgoCD syncs your workloads.
    ::::

    ::::u-page-card{spotlight class="col-span-2 lg:col-span-1 bento-card-oss" variant="subtle"}
    :bento-open-source

    #title
    Open source

    #description
    AGPL-3.0 licensed. Self-host on your own Kubernetes cluster. Built on ArgoCD, Helm, CloudNativePG, and friends.
    ::::
  :::
::

::u-page-section
#title
Built on tools you already know

  :::u-page-logos{.justify-center}
    ::::u-link{to="https://kubernetes.io" target="_blank"}
    Kubernetes
    ::::
    ::::u-link{to="https://argoproj.github.io/cd/" target="_blank"}
    ArgoCD
    ::::
    ::::u-link{to="https://helm.sh" target="_blank"}
    Helm
    ::::
    ::::u-link{to="https://gateway-api.sigs.k8s.io/" target="_blank"}
    Gateway API
    ::::
    ::::u-link{to="https://cloudnative-pg.io" target="_blank"}
    CloudNativePG
    ::::
    ::::u-link{to="https://zotregistry.dev" target="_blank"}
    Zot
    ::::
  :::
::

::u-page-section
  :::u-page-grid
    ::::u-page-card
    ---
    class: col-span-2 lg:col-span-1
    variant: soft
    ---

    #title
    Self-host it

    #description
    Run Lucity on your own Kubernetes cluster. One Helm install, full control. Your infrastructure, your rules.

      :::::u-button
      ---
      to: /getting-started/quick-start
      color: neutral
      trailing-icon: i-lucide-arrow-right
      ---
      Quick Start
      :::::
    ::::

    ::::u-page-card
    ---
    class: col-span-2 lg:col-span-1
    variant: soft
    ---

    #title
    Or let us run it

    #description
    Lucity Cloud is the managed version. Same open-source platform, hosted in Switzerland, zero infrastructure to maintain.

      :::::u-button
      ---
      to: /cloud
      color: neutral
      trailing-icon: i-lucide-arrow-right
      ---
      Join the waitlist
      :::::
    ::::
  :::
::
