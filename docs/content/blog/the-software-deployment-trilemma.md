---
title: "The Software Deployment Trilemma"
description: "There seems to be a trilemma in software deployment solutions that says: Scalable, Simple, Sovereign — pick two. "
date: "2026-04-14"
author: "Christian Blättler"
---


![Trilemma evaluation overview](/img/blog/trilemma-evaluation-overview.jpg "An overview of all categories in respect to the trilemma.")
*An overview of all categories in respect to the trilemma.*

<!-- While building Lucity, I realized there seems to be a software deployment trilemma that says: Scalable, Simple, Sovereign — pick two. Ideally we want a solution that excels in all three aspects, but the trilemma dictates that a compromise in one aspect is inevitable. This is akin to [Zooko's triangle](https://en.wikipedia.org/wiki/Zooko%27s_triangle){target="_blank"} or the [CAP theorem](https://en.wikipedia.org/wiki/CAP_theorem){target="_blank"}. -->

While building Lucity, I evaluated today's software deployment solutions and realized how the following trilemma applies to all of them: Scalable, Simple, Sovereign — pick two.

Ideally we want a solution that excels in all three aspects, but the trilemma dictates that a compromise in one aspect is inevitable. This is akin to [Zooko's triangle](https://en.wikipedia.org/wiki/Zooko%27s_triangle){target="_blank"} or the [CAP theorem](https://en.wikipedia.org/wiki/CAP_theorem){target="_blank"}.

In this article, I will categorize existing software deployment solutions into three groups: PaaS, self-hosting tools, and Kubernetes. Each group will be evaluated according to the aspects of the trilemma.

At the end I'll come back to Lucity and show where it sits within this triangle. The ejecatble architecture of Lucity means you can move around the triangle instead of being pinned into one fixed position.

Before we dive into the categories, here's my definition of the three dimensions we're examining.

**Scalability** does not apply to compute resources such as memory and CPU cores, but also geographical distribution and organization size. A truly scalable solution can support a large team working on a globally distributed deployment.

**Simplicitry** measures the greadient of the learning curve as well as the extent of ongoing maintenance efforts for operating a production deployment. This includes initial setup, maintenance work and security updates.

**Sovereignity** is a measurement how independent a solution is from third parties. This is not to be confused with self-hosting. Self-hostable software can also put you in a position where you are at mercy of a third-party (see: Broadcom VMWare license pricing debacle). Truly sovereign software builds upon widly

## 1. The Antichrist of Sovereignty: Platform as a Service

Platform as a Service (PaaS) solutions like Vercel, Heroku and Railway are the gold standard in terms of simplicity, allowing you to deploy a web service within minutes.

Over the years these platforms also established a solid quiver of scalability features: horizontal and vertical scaling, multiple geographical regions and CDN integrations ([but some don't seem to be able to manage their CDN properly](https://news.ycombinator.com/item?id=47581721){target="_blank"}).

Now as a surprise to no one, PaaS solutions fall short in terms of sovereignty. If you want to migrate off a PaaS you have to start from scratch. Like totally from scratch. This does not only mean writing Dockerfiles and building your CI/CD pipelines. It also means evaluating dozens of tools for managing deployments, observability, logs, containers, build pipelines, vulnerability scanning, and connectivity.

But that's just where the real work starts. Once you have defined your new stack, you'll need to train your teams on the technical concepts that were previously abstracted away by the PaaS. This includes things like VM lifecycles, OCI containers, networking, and VPN tunnels.

<!-- TODO: Improve next paragraph. -->

This is not to say there is no use-case for proprietary PaaS solutions. I think they work well for small or temporary projects — like personal websites or education projects. I'm also not opposed to renting someone else's server to run your workload (aka "cloud computing"). My main point of critique revolves around the opaque nature how the underlying infrastructure is hidden from developers and completely proprietary. Besides the obvious lock-in, this decision hinders curious developers to learn more about how the workload is actually deployed and learning new skills.

![Trilemma evaluation for PaaS](/img/blog/trilemma-evaluation-paas.jpg)
*PaaS represent the gold-standard in terms of simplicity but fall short on sovereignity.*

## 2. "We have PaaS at home": Self-hosting Tools

Remember the good ol' days where you just used ftp to copy some html and php to your apache web server?

```bash
ftp yourserver.com
put index.html
put contact.php
```

Self hosting tools like Coolify feel like the more polished and automated conceptual successor of this workflow. You own a server and the tool uses SSH to connect to the server to orchestrate containers using Docker. It's simple and foolproof.

<!-- TODO: I probably have to add some better structure here. Extend on the point that these tools are not really made for teams. Then add a second paragraph that touches on the "readiness" / quality of these tools. Most are solo dev efforts with uncertain futures. -->

These tools are not aimed at software engineering teams but at individuals looking for a solution to get a website or database up and running. Coolify itself says their teams feature "isn't fully polished for production use", their [security track record is a bit bumpy](https://www.aikido.dev/blog/ai-pentesting-coolify-cves){target="_blank"}, and they rely on the now deprecated Docker Swarm for multi-server deployments. Despite being open-source, tools like that create some form of vendor lock-in. The abstractions you build within Coolify are specific to that tool and not exportable.

<!-- TODO: This paragraph can probably be integrated in the first or second of this section. A separate paragraph might be justified to acknowledge the trends towards self-hosting and more awareness for digital sovereignity due to current geopolitical events. -->

Tools like Coolify are incredibly... _cool_ and I like the simple no-nonsense approach they take towards software deployment. Projects like that enable more people to run applications on servers they control.

![Trilemma evaluation for Self-hosting Tools](/img/blog/trilemma-evaluation-self-hosting-tools.jpg)
*Self-hosting tools are a solid sovereign option if you can live with the limited scalability they offer.*

## 3. Embrace the YAML: Kubernetes

Kubernetes is probably one of the most misunderstood technologies in the modern software engineering era.

As Kelsey Hightower famously said:

![Tweek by Kelsey Hightower: "Kubernetes is a platform for building platforms. It's a better place to start; not the endgame."](/img/blog/trilemma-kelsey-hightower-quote.jpg)
*Tweet by Kelsey Hightower, an icon developer within the cloud-native ecosystem.*

Kubernetes is a wildly customizable abstraction layer for running containers across many servers. If you ever manually deployed software across multiple VMs, you probably know quickly this can get messy. Kubernetes solves this problem by providing you with a unified API for managing your containers, storage, and configuration.

<!-- TODO: Improve the structure of this. -->

It's an extensible state machine backed by a distributed key-value store. If you ever had to build your own virtualization platform, you'll appreciate the flexibility it provides.

Here's where the misconception comes into play and the reason for many frustrated LinkedIn and Reddit posts. Kubernetes was never designed to be used directly by application developers.

Granting a full-stack developer access to a Kubernetes cluster and asking them to deploy their app is like passing a painter raw cotton, wooden planks, animal hair, and mineral pigments and asking them to paint a portrait. Everything they need is technically there, but they expected a canvas, brushes and paint. Someone was supposed to assemble the tools before the artist sat down to work.

So yes, technically teams can deploy their apps directly to Kubernetes. But for 95% of teams this is the wrong abstraction layer.

<!-- TODO: Go on to explain the incredible scalability and ecosystem of kubernetes. Then add another paragraph on sovereignity by explaining the architecture and also elaborate the sovereeignity pitfall of using cloud platform specific features. -->

![Trilemma evaluation for Kubernetes](/img/blog/trilemma-evaluation-kubernetes.jpg)

<!-- TODO: Better title -->

## Conclusion

<!-- TODO: This conclustion is dog shit. Improve this by speficially explaining how Lucity bends the trilemma triangle by allowing users to eject from a PaaS to a full-fledged Kubernetes setup. And how this integrates seamlessly for organizations already using Kubernetes or how this can be used as an on-ramp for organizations new to Kubernetes. -->

Conclusively we can say, that some options are incredibly simple and scalable, but fall short in respect to sovereignty. Others give you more control over your infrastructure and data, but are limiting when it comes to scale your deployment. And lastly, some give you a very scalable and sovereign solution at the expense of complexity.

What makes Lucity unique is its flexibility. You can decide yourself whether you want to self host it or use the managed cloud offering or eject your project completely and use Kubernetes directly if the additional customizability is required.

![Trilemma evaluation for Lucity](/img/blog/trilemma-evaluation-lucity.jpg)

<!-- Maybe add a note on me as the author and how to reach me. Also that this blog post was written by a human and AI was soley used to catch typos. -->