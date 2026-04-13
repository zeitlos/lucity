---
title: "The Software Deployment Trilemma: Scalable, Simple, Sovereign — pick two"
description: "I was frustrated of the status-quo in software deployment for many years. Lucity leverages established cloud native patterns and technologies to build a PaaS that feels like Railway but allows incremental adoption of Kubernetes."
date: "2026-04-13"
author: "Christian"
---

Be it a simple website or a full blown enterprise CRM solution: Nowadays, basically all ventures require some software to be deployed. To achieve this, there is no shortage of options to choose from.

Over the years of working as a software engineer I've seen many solutions in action. I've experienced how the industry migrated from VMs to container and from Docker to Kubernetes. I've witnessed Heroku transitioning from beloved platform to [slowly being killed](https://www.heroku.com/blog/an-update-on-heroku/){target="_blank"}.

This experience — along with my passion for building tools for efficient software development — eventually lead to me building my own software deployment platform: Lucity. While building Lucity, I realized there seems to be a software deployment trilemma that says: Scalable, Simple, Sovereign — pick two. Ideally we want a solution that excels in all three aspects, but the trilemma dictates that a compromise in one aspect is inevitable. This is akin to [Zooko's triangle](https://en.wikipedia.org/wiki/Zooko%27s_triangle){target="_blank"} or the [CAP theorem](https://en.wikipedia.org/wiki/CAP_theorem){target="_blank"}.

I've categorized existing software deployment solutions into three groups and evaluated each according to the software deployment trilemma.


## 1. The Antichrist of Sovereignty: Platform as a Service

Platform as a Service (PaaS) solutions like Vercel, Heroku and Railway are the gold standard in terms of simplicity, allowing you to deploy a simple web service within minutes. Over the years most of these platforms also established a solid quiver of scalability features: horizontal and vertical scaling, deployments across multiple geographical regions and CDN integrations ([some scale so well, they even show your private user data to unauthenticated users](https://news.ycombinator.com/item?id=47581721){target="_blank"} 🙃).

Now as a surprise to no one, PaaS solutions fall short in terms of sovereignty. If you want to migrate off a PaaS you have to start from scratch. Like totally from scratch. This means writing Dockerfiles and building your CI/CD pipelines. It also means evaluating dozens of tools for managing deployment, observability, logs, containers, build pipelines, vulnerability scanning, and connectivity. But that's just where the real work starts. Once you have defined your new stack, you'll need to train your teams on the technical concepts that were previously abstracted away by the PaaS. This includes things like VM lifecycles, OCI containers, networking, and VPN tunnels. 

This is not to say there is no use-case for proprietary PaaS solutions. I think they work well for small or temporary projects — like personal websites or education projects. I'm also no opposed to renting someone else's server to run your workload (aka "cloud computing"). My main point of critique revolves around the opaque nature how the underlying infrastructure is hidden from developers and completely proprietary. Besides the obvious lock-in, this decision hinders curious developers to learn more about how the workload is actually deployed and learning new skills. 

## 2. "We have PaaS at home": Self-hosting Tools

Remember the good ol' days where you just used ftp to copy some html and php to your apache web server?

```bash
ftp yourserver.com
put index.html
put contact.php
```

Self hosting tools like Coolify feel like the more polished and automated conceptual successor of this workflow. You own a server and the tool uses SSH to connect to the server to orchestrate containers using Docker. It's simple and foolproof.

These tools are not aimed at software engineering teams but at individuals to just get a website or database up and running. Coolify itself says their teams feature "isn't fully polished for production use", their [security track record is a bit bumpy](https://www.aikido.dev/blog/ai-pentesting-coolify-cves){target="_blank"}, and they rely on the now deprecated Docker Swarm for multi-server deployments. Despite being open-source, tools like that create some form of vendor lock-in. The abstractions you build within Coolify are specific to that tool and not exportable.

Tools like Coolify are incredibly... _cool_ and I like the simple no-nonsense approach they take towards software deployment. Projects like that enable more people to run applications on servers they control. 

## 3. Embrace the YAML: Kubernetes

Kubernetes is probably one of the most misunderstood technologies in the modern software engineering era.

As Daddy Hightower famously said:

> Kubernetes is a platform for building platforms. It's a better place to start; not the endgame.

Kubernetes is a wildly customizable abstraction layer for running containers across many servers. If you ever manually deployed software across multiple VMs, you probably know quickly this can get messy. Kubernetes solves this problem by providing you with a unified API for managing your containers, storage, and configuration. It's an extensible state machine backed by a distributed key-value store. If you ever had to build your own virtualization platform, you'll appreciate the flexibility it provides.

Here's where the misconception comes into play and the reason for many frustrated LinkedIn and Reddit posts. Kubernetes was never designed to be used directly by application developers. Granting a full-stack developer access to a Kubernetes cluster and asking them to deploy their app is like passing a painter raw cotton, wooden planks, animal hair, and mineral pigments and asking them to paint a portrait. Everything they need is technically there, but they expected a canvas, brushes and paint. Someone was supposed to assemble the tools before the artist sat down to work.

So yes, technically teams can deploy their apps directly to Kubernetes. But for 95% of teams this is the wrong abstraction layer.

## Conclusion

Conclusively we can say, that some options are incredibly simple and scalable, but fall short in respect to sovereignty. Others give you more control over your infrastructure and data, but are limiting when it comes to scale your deployment. And lastly, some give you a very scalable and sovereign solution at the expense of complexity.

What makes Lucity unique is it's flexibility. You can decide yourself whether you want to self host it or use the managed cloud offering or eject your project completely and use Kubernetes directly if the additional customizability is required.
