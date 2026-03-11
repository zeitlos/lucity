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
  to: https://lucity.cloud/app/login
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

::u-page-section{class="mt-16 sm:mt-24"}
#title
Everything you need to ship

#description
All the building blocks for deploying and running your apps on Kubernetes. Built on standard tools, so you can eject whenever you want.

:bento-grid
::

::u-page-section
#title
Built on tools you already know

#description
The enterprise way to deploy Kubernetes apps, cleverly automated. No proprietary runtime, no black boxes — just standard tools, loosely coupled, with zero platform state to babysit.

:tools-architecture
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
      to: https://lucity.cloud/app/login
      color: neutral
      trailing-icon: i-lucide-arrow-right
      ---
      Start for free
      :::::
    ::::
  :::
::
