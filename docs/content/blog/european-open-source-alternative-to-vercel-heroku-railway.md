---
title: "The European open-source alternative to Vercel, Heroku, and Railway"
description: "Why I built Lucity: the same developer experience as Vercel, Heroku, and Railway, on an open platform you own and can leave at any time."
date: "2026-07-27"
author: "Christian Blättler"
---

Ever since I deployed my first web app to Vercel (back when it was [still called ZEIT](https://vercel.com/blog/zeit-is-now-vercel)) as a teenager, I was hooked on the developer experience. Simple and intuitive, yet scalable enough for real projects.

At the time I was used to configuring Windows servers by hand, so it felt genuinely magical that infrastructure could be this effortless.

That fascination with infrastructure automation stuck with me. I have spent almost ten years as a software engineer on infrastructure automation projects, much of it on teams running containers at scale with Kubernetes. I saw first hand how a few well-chosen integrations can turn a plain cluster into a platform that is far greater than the sum of its parts.

So I always assumed Kubernetes would be the obvious foundation for a platform like this. That is why it surprised me to learn that Heroku, Vercel, and Railway were all built on proprietary container orchestration ([Heroku appears to have moved to Kubernetes in recent years](https://www.heroku.com/blog/next-generation-heroku-platform/)).

For a long time that was just an interesting technical detail. Then the recent geopolitical instability made something clear: these platforms fall under US jurisdiction, and with it the US CLOUD Act, which is more than a theoretical risk for a European business.

So I finally started building the platform I had been sketching in the back of my mind for years, with one goal:

> Replicate the developer experience of Vercel, Heroku, and Railway, on an open platform that never locks you in.

The result is Lucity. You can [try it for free](https://lucity.cloud/app/login). A few of the highlights:

- Git push to deploy
- Automated builds for your favorite languages and frameworks: Node.js, Next.js, Python, Go, PHP, Ruby, Rust, Java, Elixir, Bun, Deno, .NET, and static sites
- One-click PostgreSQL (replicated), S3-compatible object storage, and Redis-compatible key-value stores
- Tokenless deployments from GitHub Actions
- AI-native deployments powered by an MCP server and a Claude Code plugin
- Dynamic variables, so you reference credentials instead of hardcoding them
- Usage-based billing: pay only for what you use
- Vulnerability and leaked-credential scans on every commit
- Separate environments for development, staging, and production
- Logs, metrics, and observability built in
- Custom domains, private networking, and much more

And naturally, that headline feature is what makes "no lock-in" more than a slogan. When you want out, `lucity eject` hands you the full configuration as standard Helm charts, ready to run on any Kubernetes cluster with Lucity out of the picture. It works because Lucity keeps no database of its own: it derives all of its state from Kubernetes and the systems around it. Your infrastructure stays yours.

Lucity is under active development. Reach out at christian@zeitlos.software or open an issue at [github.com/zeitlos/lucity](https://github.com/zeitlos/lucity) for feature requests and bug reports.

It is also listed as a European alternative on:

- [European Tech Map](https://europeantechmap.eu/company/lucity)
- [euroalternative.eu](https://euroalternative.eu/lucity)
- [European Purpose](https://europeanpurpose.com/tool/lucity)
