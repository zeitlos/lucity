---
title: "The Software Deployment Trilemma"
description: "There seems to be a trilemma in software deployment solutions that says: Scalable, Simple, Sovereign — pick two."
date: "2026-04-15"
author: "Christian Blättler"
---

![Trilemma evaluation overview](/img/blog/trilemma-evaluation-overview.jpg "An overview of all categories in respect to the trilemma.")
*An overview of all categories in respect to the trilemma.*

While building Lucity, I evaluated today's software deployment solutions and realized how the following trilemma applies to all of them: Scalable, Simple, Sovereign — pick two.

Ideally we want a solution that excels in all three aspects, but the trilemma dictates that a compromise in one aspect is inevitable. This is akin to [Zooko's triangle](https://en.wikipedia.org/wiki/Zooko%27s_triangle){target="_blank"} or the [CAP theorem](https://en.wikipedia.org/wiki/CAP_theorem){target="_blank"}.

In this article, I will categorize existing software deployment solutions into three groups: PaaS, self-hosting tools, and Kubernetes. Each group will be evaluated according to the aspects of the trilemma.

At the end I'll come back to Lucity and show where it sits within this triangle. The ejectable architecture of Lucity means you can move around the triangle instead of being pinned to a fixed position.

Before we dive into the categories, here are my definitions of the three dimensions we're examining.

**Scalability** does not only apply to compute resources such as memory and CPU cores, but also geographical distribution and organization size. A truly scalable solution can support a large team working on a globally distributed deployment.

**Simplicity** measures the gradient of the learning curve as well as the extent of ongoing maintenance efforts for operating a production deployment. This includes initial setup, maintenance work, and security updates.

**Sovereignty** measures independence from third parties. This is not to be confused with self-hosting. Self-hostable software can also put you in a position where you are at the mercy of a third-party (think about the Broadcom VMware license debacle from recent years).

## 1. The Antichrist of Sovereignty: Platform as a Service

Platform as a Service (PaaS) solutions like Vercel, Heroku and Railway are the gold standard in terms of simplicity, allowing you to deploy a web service within minutes.

Over the years these platforms also established a solid quiver of scalability features: horizontal and vertical scaling, multiple geographical regions and CDN integrations ([but some don't seem to be able to manage their CDN properly](https://news.ycombinator.com/item?id=47581721){target="_blank"}).

To no one's surprise, PaaS solutions fall short in terms of sovereignty. If you want to migrate off a PaaS you have to start from scratch. Like totally from scratch. This does not only mean writing Dockerfiles and building your CI/CD pipelines. It also means evaluating dozens of tools for managing deployments, observability, logs, containers, build pipelines, vulnerability scanning, and connectivity.

But that's just where the real work starts. Once you have defined your new stack, you'll need to train your teams on the technical concepts that were previously abstracted away by the PaaS. This includes things like VM lifecycles, OCI containers, networking, and VPN tunnels.

Today's PaaS hide the underlying infrastructure layer from developers. This means curious team members will hit a wall when trying to learn how their deployments work under the hood. PaaS solutions like Vercel, Railway or Heroku are essentially gatekeeping your team from acquiring that knowledge.

These aspects help to demonstrate that the inherent vendor lock-in of PaaS solutions is a very real liability for your business. Depending on your use case, this liability might be acceptable. I believe PaaS solutions are great for short-lived endeavours, such as education projects and experiments.

Platforms like Heroku have established a new paradigm of software deployment that still offers unmatched simplicity. Lucity's user experience aims to replicate that as closely as possible. Lucity also integrates [Railpack](https://railpack.com/getting-started), an open-source project by Railway that builds containerized applications directly from source code.

Lastly, I believe critical software should give you the choice between a managed cloud offering and hosting it yourself. If recent geopolitical developments taught us anything, it's that you should not put business-critical software at the mercy of a third-party without an exit hatch.

![Trilemma evaluation for PaaS](/img/blog/trilemma-evaluation-paas.jpg)
*PaaS solutions represent the gold-standard in terms of simplicity but fall short on sovereignty.*

## 2. "We have PaaS at home": Self-Hosting Tools

Remember the good ol' days where you just used ftp to copy some html and php to your apache web server?

```bash
ftp yourserver.com
put index.html
put contact.php
```

Over the years this workflow grew up, but the shape stayed the same. You own a server, manually install your software on it, and hope nothing breaks. For containerized apps this has evolved into SSHing in, pulling your code, building an image, and running it.

```bash
ssh user@yourserver.com
git pull
docker build -t myapp .
docker run -d --restart=always -p 80:80 myapp
```

Self-hosting tools like Coolify, Cap Rover or Dokku are the conceptual successor: they automate this workflow behind a UI.

Where these tools will start showing their limits is once you want to invite your team and collaborate on projects. Generally these tools are built to be operated by one person and they lack proper team management, SSO and audit features required by a team. Coolify's docs acknowledge that by stating that their teams feature "[isn't fully polished for production use](https://coolify.io/docs/get-started/concepts#teams){target="_blank"}".

Once you look beyond organizational scaling, you will also notice that these tools fall short when it comes to scaling deployments beyond one server. Some of them support Docker Swarm for multi-server deployments, but I would advise against using that for new projects, since it's [officially retired](https://docs.docker.com/retired/#swarmkit){target="_blank"}.

Another limit you might encounter with these tools is that most of them have an unclear mid- to long-term roadmap. Furthermore, development and maintenance is often dependent on one single developer or a small team with limited experience. Recently, security company Aikido found [several high-severity CVEs in Coolify](https://www.aikido.dev/blog/ai-pentesting-coolify-cves){target="_blank"}. After skimming the CVEs I was surprised by the simplicity of these vulnerabilities, especially for an open-source project that has been around for years and accumulated over 50k GitHub stars.

It's good to see self-hosting tools like Coolify, Cap Rover or Dokku making sovereign software deployments more accessible. Just be sure you're aware of the limitations of these tools. I think these tools are great for hobby projects and deploying personal websites, but they really start to show their limits once your team grows beyond just you.

![Trilemma evaluation for Self-hosting Tools](/img/blog/trilemma-evaluation-self-hosting-tools.jpg)
*Self-hosting tools are a solid sovereign option if you can live with the limited scalability they offer.*

## 3. Embrace the YAML: Kubernetes

Kubernetes is probably one of the most misunderstood technologies in the modern software engineering era.

As Kelsey Hightower famously said:

![Tweet by Kelsey Hightower: "Kubernetes is a platform for building platforms. It's a better place to start; not the endgame."](/img/blog/trilemma-kelsey-hightower-quote.jpg)
*Tweet by Kelsey Hightower, an iconic developer in the cloud-native ecosystem.*

Kubernetes is a wildly customizable abstraction layer for running containers across many servers. If you've ever manually deployed software across multiple VMs, you probably know how quickly this can get messy. The config will drift and debugging becomes very difficult. Kubernetes solves this problem by providing you with a unified API for managing your containers, storage, and configuration.

At its core, Kubernetes is essentially an extensible state machine backed by a distributed key-value store. This extensibility becomes incredibly valuable if you want to build your own virtualization platform or integrate it into your very specific data center setup. As a developer of [flex.plane](https://flexplane.io/){target="_blank"}, I appreciate that extensibility.

But here's where the misconception comes in. It's the reason behind all those frustrated LinkedIn and Reddit posts. Kubernetes was never designed to be used directly by application developers.

Granting a full-stack developer access to a Kubernetes cluster and instructing them to deploy their app is like passing a painter raw cotton, wooden planks, animal hair, and mineral pigments and asking them to paint a portrait. Everything they need to do their job is technically there, but they expected a canvas, brushes and paint. Someone was supposed to assemble the tools before the artist sat down to work. The same applies to Kubernetes.

So yes, technically teams can deploy their apps directly to Kubernetes. But for 95% of teams Kubernetes is an abstraction layer that sits too low. Kubernetes was designed to orchestrate containers on a massive scale. This is the reason it's the tool of choice for companies like Google, Amazon and Microsoft. But that does not mean it's unusable for smaller deployments.

Thanks to its open-source nature and the adoption of said companies, a rich ecosystem formed around Kubernetes over the years. This ecosystem includes tools for managing security, observability and compliance. Proven workflows and standards have emerged, and the industry has found consensus on best practices around cloud-native software deployment.

Exactly this expertise is the reason Lucity builds atop Kubernetes. Lucity makes Kubernetes accessible by automating a proven software deployment workflow on pre-configured infrastructure. This allows you to enjoy the strengths of the Kubernetes ecosystem without worrying about its complexity. And you get the peace of mind that the exit door is there when you need it.

![Trilemma evaluation for Kubernetes](/img/blog/trilemma-evaluation-kubernetes.jpg)
*Kubernetes prioritizes scalability and extensibility at the cost of simplicity.*

## How Lucity Bends the Trilemma

Just like Vercel, Railway or Heroku, Lucity is a PaaS and shares most characteristics with this group: It's simple to get started and offers automated build & deploy pipelines, observability, scaling and security scanning.

But this is where Lucity differs from the other solutions: It allows you to export the underlying Helm chart, GitOps repo, and configuration and deploy them on any Kubernetes cluster. This is useful if you ever need customization beyond what's possible within the platform or if you need to use Kubernetes-specific features.

Additionally, Lucity is open-source (AGPLv3 license) and allows users to self-host the platform with the full feature set. While all major PaaS are operated by US companies, Lucity is built by [zeitlos.software](https://zeitlos.software/){target="_blank"}—a Swiss company—and hosted in the EU, governed by strict privacy laws.

This makes Lucity the only solution that can move around the trilemma triangle instead of being pinned to a fixed position.

If that sounds interesting to you, give it a try and [sign up for a free trial](https://lucity.cloud/app/login?utm_source=blog&utm_medium=post&utm_campaign=software-deployment-trilemma){target="_blank"}.

![Trilemma evaluation for Lucity](/img/blog/trilemma-evaluation-lucity.jpg)
*Lucity is the only trilemma shape that can change. Start with an open-source PaaS and eject to a full Kubernetes setup if you outgrow it.*

---

Thanks for reading this post on the software deployment trilemma. I'm interested to hear your questions, comments or feedback. Feel free to reach me at [christian@zeitlos.software](mailto:christian@zeitlos.software).

PS: This blog post was written by a real human.