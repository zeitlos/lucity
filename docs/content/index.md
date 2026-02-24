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

  :::u-page-grid
    ::::u-page-card
    ---
    spotlight: true
    class: col-span-2
    to: /features/builds
    variant: outline
    ---
    ```sh
    $ git push origin main
    → Detected: Node.js 22 + Next.js
    → Building image…
    → Pushed to registry
    → Deployed to development ✓
    ```

    #title
    :u-icon{name="i-lucide-rocket" class="text-primary"} Push to deploy

    #description
    Connect your GitHub repo. Lucity detects your framework, builds a container image, and deploys it. Zero Dockerfiles required.
    ::::

    ::::u-page-card
    ---
    spotlight: true
    class: col-span-2 lg:col-span-1
    to: /features/environments
    variant: outline
    ---

    #title
    :u-icon{name="i-lucide-layers" class="text-primary"} Multi-environment

    #description
    Development, staging, production, and PR preview environments out of the box. Promote images between them without rebuilding.
    ::::

    ::::u-page-card
    ---
    spotlight: true
    class: col-span-2 lg:col-span-1
    to: /infrastructure/databases
    variant: outline
    ---

    #title
    :u-icon{name="i-lucide-database" class="text-primary"} Batteries included

    #description
    PostgreSQL via CloudNativePG, Redis, cron jobs, and HTTP routing via Gateway API. Everything your app needs.
    ::::

    ::::u-page-card
    ---
    spotlight: true
    class: col-span-2
    to: /features/eject
    variant: outline
    ---
    ```
    $ lucity eject --project myapp
    Ejecting project "myapp"…
    ✓ Helm chart        → ejected/chart/
    ✓ ArgoCD apps       → ejected/argocd/
    ✓ Environment values → ejected/environments/
    ✓ README            → ejected/README.md
    Done. Your infrastructure is yours.
    ```

    #title
    :u-icon{name="i-lucide-door-open" class="text-primary"} Eject anytime

    #description
    Export your entire setup as standard Helm charts and ArgoCD configs. No lock-in. Your infrastructure, always.
    ::::

    ::::u-page-card
    ---
    spotlight: true
    class: col-span-2 lg:col-span-1
    to: /architecture/gitops
    variant: outline
    ---

    #title
    :u-icon{name="i-lucide-git-branch" class="text-primary"} GitOps native

    #description
    Every deployment is a Git commit. ArgoCD syncs your workloads. Full audit trail, real rollbacks.
    ::::

    ::::u-page-card
    ---
    spotlight: true
    class: col-span-2 lg:col-span-1
    to: https://github.com/zeitlos/lucity
    variant: outline
    target: _blank
    ---

    #title
    :u-icon{name="i-lucide-heart" class="text-primary"} Open source

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
    Lucity Cloud is the managed version — same open-source platform, hosted in Switzerland, zero infrastructure to maintain.

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
