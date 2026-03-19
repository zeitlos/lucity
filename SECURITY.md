# Security Policy

Lucity manages Kubernetes deployments, build pipelines, OCI registries, and database credentials on behalf of its users. We take security seriously and appreciate responsible disclosure from the community.

## Supported Versions

| Version | Supported |
| ------- | --------- |
| Latest release | Yes |
| Previous minor | Critical and high severity only |
| Older | No |

## Reporting a Vulnerability

**Do not open a public issue for security vulnerabilities.**

Email: **security@lucity.cloud**

Please include:

- Description of the vulnerability
- Steps to reproduce
- Affected component (gateway, builder, packager, deployer, webhook, dashboard)
- Impact assessment (e.g. data exposure, privilege escalation, RCE)
- Any suggested fix or mitigation

You will receive an acknowledgment within 2 business days. We aim to provide a detailed response within 5 business days, including our severity assessment and remediation timeline.

## Scope

The following are in scope:

- Lucity core services (gateway, builder, packager, deployer, webhook)
- GraphQL API authentication and authorization
- Secret and credential handling (environment variables, database credentials)
- Build pipeline isolation (container escapes, supply chain attacks)
- GitOps sync and webhook validation
- RBAC and multi-tenancy boundaries

Out of scope:

- Vulnerabilities in upstream dependencies (report to the upstream project)
- Issues requiring physical access to the host
- Social engineering attacks
- Denial of service attacks

## Safe Harbor

We will not take legal action against researchers who:

- Act in good faith and follow this policy
- Avoid accessing or modifying other users' data
- Do not degrade service availability
- Report findings promptly and allow reasonable time for remediation

## Disclosure

We follow coordinated disclosure. We ask reporters to allow up to 90 days for a fix before public disclosure. Credit will be given in release notes unless anonymity is requested.
