---
title: Data Processing Agreement
description: How Lucity processes personal data on your behalf.
---

**Effective date:** July 24, 2026
**Last updated:** July 24, 2026

This Data Processing Agreement ("DPA") forms part of the [Terms of Use](/legal/terms-of-use) between you ("Customer", "Controller") and the operator of Lucity ("Processor", "we", "us"):

zeitlos.software Inh. Christian Blättler
CHE‑439.475.468
Mattenhofstrasse 5, 3007 Bern, Switzerland

It governs the processing of personal data that we carry out on your behalf when you use lucity.cloud to build, deploy, and run your applications. It applies where you are subject to the EU General Data Protection Regulation (GDPR), the Swiss Federal Act on Data Protection (FADP/nDSG), or both.

Where you act as a processor for your own customers, references to "Controller" also cover your role as their processor, and we act as your sub-processor.

---

## 1. Roles and Scope

- You are the **controller** of the personal data your applications process (for example, data about your end users). We are the **processor**, acting only on your documented instructions.
- Your use of the platform, including the configuration of your workloads, environment variables, and services, constitutes your documented instructions. Additional instructions must be agreed in writing.
- We process personal data only to provide the service described in the Terms of Use and do not process it for our own purposes.
- We handle your **account and billing data** as a controller in our own right; that processing is described in the [Privacy Policy](/legal/privacy-policy) and is outside the scope of this DPA.

## 2. Subject Matter of Processing

| Item | Description |
|------|-------------|
| **Subject matter** | Hosting and running the Customer's applications and their data on the platform |
| **Duration** | For the term of the Customer's use of the service |
| **Nature and purpose** | Building, deploying, storing, and executing the Customer's workloads, databases, key-value stores, and object storage |
| **Types of personal data** | Determined by the Customer. Any personal data the Customer's applications store or process on the platform |
| **Categories of data subjects** | Determined by the Customer. Typically the Customer's own users and contacts |

We have no control over, and do not inspect, the categories of personal data or data subjects the Customer chooses to process through their workloads.

## 3. Our Obligations

We will:

- **Process only on instructions.** Process personal data solely on your documented instructions, including regarding international transfers, unless required otherwise by applicable law, in which case we will inform you unless the law prohibits it.
- **Confidentiality.** Ensure that personnel authorized to process the data are bound by confidentiality.
- **Security.** Implement appropriate technical and organizational measures as described in Annex 1.
- **Sub-processors.** Use sub-processors only under the conditions in Section 4.
- **Assist with data-subject rights.** Taking into account the nature of the processing, assist you by appropriate measures in responding to requests from data subjects exercising their rights.
- **Assist with compliance.** Assist you in ensuring compliance with your security, breach-notification, and data-protection-impact-assessment obligations, taking into account the information available to us.
- **Breach notification.** Notify you without undue delay after becoming aware of a personal data breach affecting your data, with the information reasonably available to us.
- **Deletion or return.** On termination, delete or return the personal data as described in Section 5.
- **Audits.** Make available the information necessary to demonstrate compliance with this DPA and allow for and contribute to audits, including inspections, conducted by you or an auditor you mandate, subject to reasonable notice, confidentiality, and no undue disruption to our operations.

## 4. Sub-processors

You provide general authorization for us to engage the sub-processors listed below to process personal data on your behalf. We impose data-protection obligations on each sub-processor that are no less protective than those in this DPA, and we remain responsible for their performance.

| Sub-processor | Purpose | Location |
|---------------|---------|----------|
| **Hetzner** | Core infrastructure hosting (compute, workloads, databases) | Germany (EU) |
| **OVHcloud** | Object storage hosting | France (EU) |
| **Bunny** | Content delivery for public buckets and custom-domain TLS | EU company; global edge network |
| **Stripe** | Payment processing (account and billing data only) | EU and international, under Standard Contractual Clauses |

We will give you at least 30 days' notice, by email or platform notification, before adding or replacing a sub-processor. If you reasonably object on data-protection grounds, we will work with you in good faith to address the concern, and if we cannot, you may terminate the affected service.

## 5. Deletion and Return

On termination of the service, or on your request, we will delete the personal data we process on your behalf within 30 days, unless applicable law requires us to retain it. When you delete a project or account, the associated data is removed within 30 days. Because the platform is ejectable, you can export your configuration and data before termination.

## 6. International Transfers

Personal data processed on your behalf is stored in the European Union (see the [Privacy Policy](/legal/privacy-policy)). Where a sub-processor transfers personal data outside the EU or Switzerland, that transfer is covered by an appropriate safeguard, such as the EU Standard Contractual Clauses together with, for data subject to the FADP, the recognition of those clauses by the Swiss Federal Data Protection and Information Commissioner.

## 7. Swiss FADP

Where the FADP applies, references to the GDPR are read as references to the equivalent provisions of the FADP, "personal data" includes data relating to legal entities to the extent protected by the FADP, and the supervisory authority is the Federal Data Protection and Information Commissioner (FDPIC).

## 8. Liability and Precedence

The liability limitations in the Terms of Use apply to this DPA. If there is a conflict between this DPA and the Terms of Use regarding the processing of personal data, this DPA prevails.

## 9. Contact

For data-protection matters and to exercise controller rights under this DPA: privacy@lucity.cloud

---

## Annex 1: Technical and Organizational Measures

We maintain measures appropriate to the risk, including:

- **Encryption in transit** using TLS for all connections to and within the platform.
- **Encryption at rest** for stored data, including databases and object storage.
- **Tenant isolation** through namespace-level separation of workspaces in Kubernetes, with network policies restricting cross-tenant access.
- **Access control** on a least-privilege, role-based basis, with authentication through our self-hosted identity provider.
- **Secrets handling** through Kubernetes secrets, kept separate from application code and logs.
- **Resilience** through replicated databases and backups of managed database services.
- **Monitoring and logging** of platform activity to detect and respond to security events.
- **Self-hosted core services** (identity, container registry, deployment tooling) within our own cluster, limiting exposure to external processors.
