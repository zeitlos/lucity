---
title: Terms of Use
description: Terms governing your use of Lucity.
navigation: false
---

**Effective date:** March 16, 2026
**Last updated:** July 24, 2026

These terms govern your use of lucity.cloud ("Lucity", "the platform", "the service"). By creating an account, you agree to these terms. If you use the service on behalf of an organization, you represent that you are authorized to bind that organization, and "you" refers to that organization.

**Operator:**
zeitlos.software Inh. Christian Blättler
CHE‑439.475.468
Mattenhofstrasse 5, 3007 Bern, Switzerland

**Contact:** hello@lucity.cloud

---

## 1. The Service

Lucity is a platform-as-a-service (PaaS) that deploys and runs your applications on Kubernetes. The platform manages builds, deployments, and infrastructure on your behalf.

Lucity is open source under the AGPL-3.0 license. You may self-host it on your own infrastructure free of charge under the terms of that license. These terms apply only to the hosted service at lucity.cloud, not to your own self-hosted deployments.

## 2. Accounts

- You must provide accurate information when creating an account.
- You are responsible for maintaining the security of your account credentials and for all activity under your account.
- You must be at least 16 years old to use the service.
- One person or legal entity per account. Workspaces provide multi-user collaboration.

## 3. Plans and Billing

### Subscription

Lucity offers paid subscription plans (currently Hobby at EUR 5/month and Pro at EUR 25/month). Each plan includes a monthly credit allowance applied toward resource usage.

### Resource Usage

Resource consumption (CPU, memory, disk, egress) is metered continuously. Usage beyond included credits is billed at the end of each billing cycle at the published rates.

### Payment

Payments are processed by Stripe. You authorize recurring charges to your payment method. All prices are in EUR and are exclusive of any applicable taxes (such as VAT/MWST). Where a tax applies to your purchase, it will be added at the applicable rate, and you are responsible for it. If your payment fails, we may suspend the service after reasonable notice until payment is made.

### Cancellation

You may cancel your subscription at any time. The service remains available until the end of the current billing period. We do not offer refunds for partial billing periods.

### Price Changes

We may adjust pricing with 30 days' notice. Continued use after the effective date constitutes acceptance. If you disagree, cancel before the change takes effect.

## 4. Acceptable Use

You agree not to use the platform, and not to allow anyone using your account or workloads to use the platform, to:

- Violate any applicable law or regulation
- Infringe on intellectual property rights or other rights of third parties
- Store, host, or distribute content that is unlawful, including child sexual abuse material
- Distribute malware, spam, or phishing content, or operate botnets, command-and-control infrastructure, or open mail relays
- Originate denial-of-service attacks, port scanning, or intrusion attempts against any system
- Operate anonymizing proxies or services intended to conceal the origin of abusive traffic
- Perform cryptocurrency mining
- Run workloads that consume excessive shared resources in a way that degrades service for others
- Attempt to access other users' data or workloads, or to circumvent tenant isolation
- Reverse-engineer, attack, or probe the platform's infrastructure

We reserve the right to suspend or terminate accounts that violate these terms.

### Reporting Abuse

To report content or activity on the platform that you believe violates these terms, contact **hello@lucity.cloud**. We review reports and, where appropriate, remove content or suspend the responsible account. We may act immediately in cases of severe abuse or legal risk.

## 5. Your Data and Content

- **You own your data.** We claim no intellectual property rights over your source code, configurations, or application data.
- **License to operate.** To provide the service, you grant us a limited, non-exclusive, worldwide license to host, copy, build, store, process, and transmit your content, solely as needed to operate the platform and run your workloads on your behalf. This license ends when the content is deleted from the platform.
- **Ejectability.** You can export your platform configuration (Helm charts and environment values) at any time via the eject feature. The exported output is fully self-contained with no Lucity dependencies.
- **Deletion.** When you delete a project or account, associated data (builds, configurations, deployments) is removed within 30 days.

## 6. Your Responsibilities

You are solely responsible for:

- The content, code, and workloads you deploy, and for ensuring you have all rights necessary to upload and run them.
- Compliance by your applications, and your end users, with all applicable laws, including data-protection law.
- The security and configuration of your own applications, including secrets you set as environment variables.
- Maintaining your own backups of any data you cannot afford to lose. The platform is not a backup service, and we do not guarantee against loss of your application data.

You warrant that your use of the service will comply with these terms and all applicable laws.

## 7. Platform Availability

- We aim for high availability but do not guarantee specific uptime percentages.
- The platform is provided "as is". Scheduled maintenance will be announced in advance when possible.
- Lucity is non-intrusive: platform downtime does not affect your running workloads. Your applications continue to run on Kubernetes independently.

## 8. Limitation of Liability

To the maximum extent permitted by Swiss law:

- Lucity is provided **"as is"** without warranties of any kind, whether express or implied.
- We are not liable for indirect, incidental, or consequential damages (including lost profits, data loss, or business interruption).
- Our total liability is limited to the amount you paid for the service in the 12 months preceding the claim.
- We are not responsible for outages, data loss, or issues caused by third-party services, your application code, or circumstances beyond our control.

Nothing in these terms excludes or limits liability that cannot be excluded or limited under Swiss law, including liability for unlawful intent or gross negligence under Art. 100 of the Swiss Code of Obligations.

## 9. Indemnification

You will defend, indemnify, and hold harmless zeitlos.software and its operator against any third-party claim, demand, loss, or damage (including reasonable legal costs) arising out of or related to:

- Your content, code, or workloads
- Your use of the service in breach of these terms or any applicable law
- Your violation of the rights of a third party

We will notify you of any such claim, and you may control its defense, provided that any settlement that imposes obligations on us requires our prior written consent.

## 10. Intellectual Property

- **Lucity** is open-source software licensed under AGPL-3.0. The source code is available at [github.com/zeitlos/lucity](https://github.com/zeitlos/lucity).
- The Lucity name, logo, and branding are trademarks of zeitlos.software.
- Your use of the platform does not grant you rights to our trademarks.

## 11. Suspension and Termination

- **By you:** Cancel your subscription at any time through the billing portal or by contacting us.
- **By us:** We may suspend or terminate your account for violation of these terms or for non-payment, with notice where reasonably possible. In cases of severe abuse or legal risk, we may act immediately.
- **Effect:** Upon termination, your workloads will be stopped and data deleted within 30 days. We recommend ejecting your configuration before cancellation.

## 12. Data Protection

Our handling of personal data is described in our [Privacy Policy](/legal/privacy-policy). Where you use the platform to process personal data for which you are the controller (for example, data of your own end users), our [Data Processing Agreement](/legal/data-processing-agreement) applies and forms part of these terms.

## 13. Changes to These Terms

We may update these terms. Material changes will be communicated via email or platform notification at least 14 days before taking effect. Continued use after the effective date constitutes acceptance.

## 14. Governing Law and Jurisdiction

These terms are governed by the laws of Switzerland, excluding its conflict-of-law rules and the UN Convention on Contracts for the International Sale of Goods (CISG). The exclusive place of jurisdiction for any dispute is Bern, Switzerland. If you are a consumer, mandatory consumer-protection law may entitle you to bring proceedings at your place of residence, and nothing here removes that right.

## 15. Miscellaneous

- **Force majeure.** We are not liable for any failure or delay caused by events beyond our reasonable control, including outages of third-party infrastructure, network failures, or acts of government.
- **Assignment.** You may not assign these terms without our consent. We may assign them to a successor of our business.
- **Severability.** If any provision is held unenforceable, the remaining provisions stay in effect.
- **Entire agreement.** These terms, together with the Privacy Policy and, where applicable, the Data Processing Agreement, are the entire agreement between you and us regarding the service.

## 16. Contact

Questions about these terms: hello@lucity.cloud
