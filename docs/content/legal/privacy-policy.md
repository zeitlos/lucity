---
title: Privacy Policy
description: How Lucity handles your data.
---

# Privacy Policy

**Effective date:** March 16, 2026
**Last updated:** March 16, 2026

This privacy policy explains how lucity.cloud ("Lucity", "we", "us") collects, uses, and protects your data.

**Operator:**
zeitlos.software Inh. Christian Blättler
CHE‑439.475.468
Mattenhofstrasse 5, 3007 Bern, Switzerland

**Contact:** privacy@lucity.cloud

---

## 1. What We Collect

### Account Data

When you sign up, we collect:

- **Email address** and **display name** (via our authentication service)
- **Billing information** (name, billing address, payment method) — processed by Stripe

### Usage Data

When you use the platform, we process:

- **Project and deployment metadata** — project names, environment configurations, build logs, deployment status
- **Resource consumption** — CPU, memory, disk, and egress usage for billing

### Analytics

We use [Rybbit](https://rybbit.com) (self-hosted) for website analytics. Rybbit is cookie-free, does not collect personal data, and does not track users across websites. No cookie consent banner is required.

### What We Do Not Collect

- We do not read or store your application source code beyond what is needed during the build process
- We do not use cookies for tracking
- We do not build advertising profiles
- We do not sell your data

## 2. How We Use Your Data

- **Provide the service** — run your workloads, manage deployments, process builds
- **Billing** — calculate resource usage, process payments, send invoices
- **Communication** — service notifications, security alerts, billing updates
- **Improve the platform** — aggregated, anonymized usage patterns

## 3. Data Processing and Storage

All platform data is processed and stored in the **European Union** (Hetzner Cloud, Nuremberg, Germany).

Our infrastructure services (authentication, Git server, container registry, deployment tooling, analytics) are **self-hosted** within our own cluster. No third party has access to your workload data, source code, or deployment configurations.

### Third-Party Processors

| Provider | Purpose | Data Shared | Privacy Info |
|----------|---------|-------------|--------------|
| **Stripe** | Payment processing | Name, email, billing address, payment method | [stripe.com/privacy](https://stripe.com/privacy) |
| **Hetzner** | Infrastructure hosting | Data stored on their servers (encrypted at rest) | [hetzner.com/privacy-policy](https://www.hetzner.com/privacy-policy/) |

Stripe may process data outside the EU. Stripe maintains EU Standard Contractual Clauses for international transfers. We do not share data with any other third parties.

## 4. Data Retention

- **Account data** — retained while your account is active. Deleted within 30 days of account deletion.
- **Build artifacts and logs** — retained while the associated project exists.
- **Billing records** — retained for 10 years as required by Swiss law (OR Art. 958f).
- **Analytics data** — aggregated and anonymous; no personal data retained.

## 5. Your Rights

Under the Swiss Federal Act on Data Protection (nDSG/FADP) and the EU General Data Protection Regulation (GDPR), you have the right to:

- **Access** your personal data
- **Correct** inaccurate data
- **Delete** your account and associated data
- **Export** your data (platform data is ejectable by design)
- **Object** to processing
- **Withdraw consent** where processing is based on consent

To exercise any of these rights, email privacy@lucity.cloud.

## 6. Security

We protect your data with:

- TLS encryption for all data in transit
- Encrypted storage at rest
- Role-based access control
- Isolated tenant workspaces (namespace-level separation in Kubernetes)

## 7. Children

Lucity is not directed at children under 16. We do not knowingly collect data from children under 16. If you believe a child has provided us with personal data, please contact us.

## 8. Changes

We may update this policy. Material changes will be communicated via email or platform notification. Continued use after changes constitutes acceptance.

## 9. Contact and Supervisory Authority

For privacy questions: privacy@lucity.cloud

If you believe your data protection rights have been violated, you may lodge a complaint with:

- **Switzerland:** [Federal Data Protection and Information Commissioner (FDPIC)](https://www.edoeb.admin.ch/)
- **EU:** Your local data protection authority
