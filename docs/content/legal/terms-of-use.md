---
title: Terms of Use
description: Terms governing your use of Lucity.
---

# Terms of Use

**Effective date:** March 16, 2026
**Last updated:** March 16, 2026

These terms govern your use of lucity.cloud ("Lucity", "the platform", "the service"). By creating an account, you agree to these terms.

**Operator:**
zeitlos.software Inh. Christian Blättler
CHE‑439.475.468
Mattenhofstrasse 5, 3007 Bern, Switzerland

**Contact:** hello@lucity.cloud

---

## 1. The Service

Lucity is a platform-as-a-service (PaaS) that deploys and runs your applications on Kubernetes. The platform manages builds, deployments, and infrastructure on your behalf.

Lucity is open source under the AGPL-3.0 license. You may self-host it on your own infrastructure free of charge under the terms of that license.

## 2. Accounts

- You must provide accurate information when creating an account.
- You are responsible for maintaining the security of your account credentials.
- You must be at least 16 years old to use the service.
- One person or legal entity per account. Workspaces provide multi-user collaboration.

## 3. Plans and Billing

### Subscription

Lucity offers paid subscription plans (currently Hobby at EUR 5/month and Pro at EUR 25/month). Each plan includes a monthly credit allowance applied toward resource usage.

### Resource Usage

Resource consumption (CPU, memory, disk, egress) is metered continuously. Usage beyond included credits is billed at the end of each billing cycle at the published rates.

### Payment

Payments are processed by Stripe. You authorize recurring charges to your payment method. All prices are in EUR. No VAT/MWST is charged.

### Cancellation

You may cancel your subscription at any time. The service remains available until the end of the current billing period. We do not offer refunds for partial billing periods.

### Price Changes

We may adjust pricing with 30 days' notice. Continued use after the effective date constitutes acceptance. If you disagree, cancel before the change takes effect.

## 4. Acceptable Use

You agree not to use the platform to:

- Violate any applicable law or regulation
- Infringe on intellectual property rights
- Distribute malware, spam, or phishing content
- Perform cryptocurrency mining
- Run workloads that consume excessive shared resources in a way that degrades service for others
- Attempt to access other users' data or workloads
- Reverse-engineer, attack, or probe the platform's infrastructure

We reserve the right to suspend or terminate accounts that violate these terms.

## 5. Your Data and Content

- **You own your data.** We claim no intellectual property rights over your source code, configurations, or application data.
- **Ejectability.** You can export your platform configuration (Helm charts, ArgoCD manifests, environment values) at any time via the eject feature. The exported output is fully self-contained with no Lucity dependencies.
- **Deletion.** When you delete a project or account, associated data (builds, configurations, deployments) is removed within 30 days.

## 6. Platform Availability

- We aim for high availability but do not guarantee specific uptime percentages.
- The platform is provided "as is". Scheduled maintenance will be announced in advance when possible.
- Lucity is non-intrusive: platform downtime does not affect your running workloads. Your applications continue to run on Kubernetes independently.

## 7. Limitation of Liability

To the maximum extent permitted by Swiss law:

- Lucity is provided **"as is"** without warranties of any kind, whether express or implied.
- We are not liable for indirect, incidental, or consequential damages (including lost profits, data loss, or business interruption).
- Our total liability is limited to the amount you paid for the service in the 12 months preceding the claim.
- We are not responsible for outages, data loss, or issues caused by third-party services, your application code, or circumstances beyond our control.

Nothing in these terms excludes liability for fraud or willful misconduct.

## 8. Intellectual Property

- **Lucity** is open-source software licensed under AGPL-3.0. The source code is available at [github.com/zeitlos/lucity](https://github.com/zeitlos/lucity).
- The Lucity name, logo, and branding are trademarks of zeitlos.software.
- Your use of the platform does not grant you rights to our trademarks.

## 9. Termination

- **By you:** Cancel your subscription at any time through the billing portal or by contacting us.
- **By us:** We may suspend or terminate your account for violation of these terms, with notice where reasonably possible. In cases of severe abuse, we may act immediately.
- **Effect:** Upon termination, your workloads will be stopped and data deleted within 30 days. We recommend ejecting your configuration before cancellation.

## 10. Changes to These Terms

We may update these terms. Material changes will be communicated via email or platform notification at least 14 days before taking effect. Continued use after the effective date constitutes acceptance.

## 11. Governing Law and Jurisdiction

These terms are governed by the laws of Switzerland. Any disputes arising from these terms are subject to the exclusive jurisdiction of the courts of Bern, Switzerland.

## 12. Contact

Questions about these terms: hello@lucity.cloud
