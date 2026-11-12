import type { LucityGroup } from './types';

export const lucityGroups: LucityGroup[] = [
  {
    title: 'Builds',
    rows: [
      {
        key: 'auto',
        name: 'Builds without a Dockerfile',
        lucity: { mark: 'yes', text: 'framework detected from your files' },
      },
      {
        key: 'dockerfile',
        name: 'Builds from your Dockerfile',
        lucity: { mark: 'no', text: 'a Dockerfile is ignored; publish the image yourself and run it as a prebuilt image' },
      },
      {
        key: 'images',
        name: 'Prebuilt images',
        lucity: { mark: 'partly', text: 'public registries only, and the tag is fixed once created' },
      },
      {
        key: 'cache',
        name: 'Build cache',
        lucity: { mark: 'yes' },
      },
      {
        key: 'monorepo',
        name: 'Monorepos',
        lucity: { mark: 'yes', text: 'a folder per service' },
      },
      {
        key: 'settings',
        name: 'Build settings',
        lucity: { mark: 'yes', text: 'install, build and start commands, system packages' },
      },
      {
        key: 'templates',
        name: 'One-click templates',
        lucity: { mark: 'no' },
      },
    ],
  },
  {
    title: 'Deploys',
    rows: [
      {
        key: 'push',
        name: 'Push to deploy',
        lucity: { mark: 'yes' },
      },
      {
        key: 'ci',
        name: 'Deploys from GitHub Actions',
        lucity: { mark: 'yes', text: 'no stored token, the job proves who it is' },
      },
      {
        key: 'zero',
        name: 'Zero-downtime rollouts',
        lucity: { mark: 'yes' },
      },
      {
        key: 'config',
        name: 'Config changes without a rebuild',
        lucity: { mark: 'yes', text: 'variables, size, start command' },
      },
      {
        key: 'reasons',
        name: 'Rollout failure reasons',
        lucity: { mark: 'yes', text: 'out of memory, crash loop, failed health check' },
      },
      {
        key: 'rollback',
        name: 'Rollbacks',
        lucity: { mark: 'yes' },
      },
      {
        key: 'envs',
        name: 'Environments',
        lucity: { mark: 'yes', text: 'separate copies of one project' },
      },
      {
        key: 'previews',
        name: 'Preview environments per pull request',
        lucity: { mark: 'no' },
      },
      {
        key: 'cron',
        name: 'Cron jobs',
        lucity: { mark: 'no' },
      },
      {
        key: 'release',
        name: 'Commands before a release',
        lucity: { mark: 'no', text: 'chain migrations in front of the start command' },
      },
    ],
  },
  {
    title: 'Services',
    rows: [
      {
        key: 'size',
        name: 'CPU and memory',
        lucity: { mark: 'yes', text: 'set each separately, up to 4 vCPU and 16 GB on Pro' },
      },
      {
        key: 'replicas',
        name: 'Replicas',
        lucity: { mark: 'yes', text: 'up to 3 on Hobby, 10 on Pro' },
      },
      {
        key: 'autoscale',
        name: 'Autoscaling',
        lucity: { mark: 'yes', text: 'on CPU' },
      },
      {
        key: 'sleep',
        name: 'Sleep when idle',
        lucity: { mark: 'no', text: 'idle services cost little, since CPU and memory are billed on use' },
      },
      {
        key: 'private',
        name: 'Private networking',
        lucity: { mark: 'yes', text: 'services reach each other by name' },
      },
      {
        key: 'vars',
        name: 'Variables and shared variables',
        lucity: { mark: 'yes' },
      },
      {
        key: 'refs',
        name: 'References to databases',
        lucity: { mark: 'yes', text: 'picked from a list, never copied' },
      },
      {
        key: 'health',
        name: 'Health checks',
        lucity: { mark: 'yes', text: 'TCP by default, or HTTP' },
      },
      {
        key: 'tcp',
        name: 'Public TCP ports for services',
        lucity: { mark: 'no', text: 'HTTP only; databases have their own public access' },
      },
      {
        key: 'egressip',
        name: 'Static outbound IP',
        lucity: { mark: 'no' },
      },
    ],
  },
  {
    title: 'Domains',
    rows: [
      {
        key: 'platform',
        name: 'Platform domains with HTTPS',
        lucity: { mark: 'yes' },
      },
      {
        key: 'custom',
        name: 'Custom domains',
        lucity: { mark: 'yes', text: '1 on Hobby, unlimited on Pro' },
      },
      {
        key: 'certs',
        name: 'Automatic certificates',
        lucity: { mark: 'yes' },
      },
      {
        key: 'https',
        name: 'HTTPS only',
        lucity: { mark: 'yes', text: 'plain HTTP is redirected' },
      },
      {
        key: 'redirects',
        name: 'Domain redirects',
        lucity: { mark: 'yes', text: 'www to apex, old domain to new' },
      },
      {
        key: 'edge',
        name: 'Global edge',
        lucity: { mark: 'yes', text: 'visitors connect to a location near them' },
      },
      {
        key: 'waf',
        name: 'Firewall rules for web traffic',
        lucity: { mark: 'no' },
      },
    ],
  },
  {
    title: 'Databases',
    rows: [
      {
        key: 'pg',
        name: 'Managed PostgreSQL',
        lucity: { mark: 'yes' },
      },
      {
        key: 'ha',
        name: 'High availability',
        lucity: { mark: 'yes', text: 'a standby on another machine, on every plan' },
      },
      {
        key: 'pitr',
        name: 'Point-in-time recovery',
        lucity: { mark: 'yes', text: 'any second of the last 30 days, every plan' },
      },
      {
        key: 'pgsize',
        name: 'Flexible database sizing',
        lucity: { mark: 'yes', text: 'CPU, memory and storage set one by one' },
      },
      {
        key: 'allow',
        name: 'Allowed addresses for public access',
        lucity: { mark: 'yes' },
      },
      {
        key: 'explorer',
        name: 'Table browser and SQL editor',
        lucity: { mark: 'yes' },
      },
      {
        key: 'topq',
        name: 'Top queries',
        lucity: { mark: 'yes' },
      },
      {
        key: 'pgvector',
        name: 'pgvector',
        lucity: { mark: 'yes' },
      },
      {
        key: 'kv',
        name: 'Redis-compatible store',
        lucity: { mark: 'yes' },
      },
      {
        key: 'otherdb',
        name: 'MySQL, MongoDB and others',
        lucity: { mark: 'partly', text: 'as prebuilt images with a volume, without managed backups' },
      },
    ],
  },
  {
    title: 'Storage',
    rows: [
      {
        key: 'buckets',
        name: 'S3-compatible buckets',
        lucity: { mark: 'yes' },
      },
      {
        key: 'public',
        name: 'Public buckets on a CDN',
        lucity: { mark: 'yes' },
      },
      {
        key: 'files',
        name: 'Bucket file explorer',
        lucity: { mark: 'yes' },
      },
      {
        key: 'volumes',
        name: 'Volumes',
        lucity: { mark: 'yes', text: 'resized while running' },
      },
      {
        key: 'volbackup',
        name: 'Volume backups',
        lucity: { mark: 'no', text: 'copy important files to a bucket on a schedule' },
      },
    ],
  },
  {
    title: 'Logs and metrics',
    rows: [
      {
        key: 'livelogs',
        name: 'Live logs',
        lucity: { mark: 'yes' },
      },
      {
        key: 'history',
        name: 'Log history and search',
        lucity: { mark: 'yes', text: '14 days, with surrounding lines' },
      },
      {
        key: 'events',
        name: 'Platform events in the logs',
        lucity: { mark: 'yes', text: 'restarts and failed health checks between your lines' },
      },
      {
        key: 'metrics',
        name: 'CPU and memory metrics',
        lucity: { mark: 'yes' },
      },
      {
        key: 'dbmetrics',
        name: 'Database metrics',
        lucity: { mark: 'yes', text: 'connections, cache hits, replica lag' },
      },
      {
        key: 'alerts',
        name: 'Alerts and notifications',
        lucity: { mark: 'no' },
      },
      {
        key: 'dash',
        name: 'Custom dashboards and tracing',
        lucity: { mark: 'no' },
      },
    ],
  },
  {
    title: 'Tools',
    rows: [
      {
        key: 'cli',
        name: 'CLI',
        lucity: { mark: 'yes' },
      },
      {
        key: 'mcp',
        name: 'MCP server for agents',
        lucity: { mark: 'yes', text: 'agents see variable names, never values' },
      },
      {
        key: 'api',
        name: 'Public API',
        lucity: { mark: 'yes', text: 'the same API the dashboard uses' },
      },
      {
        key: 'tokens',
        name: 'API tokens',
        lucity: { mark: 'yes', text: 'each with its own role' },
      },
      {
        key: 'shell',
        name: 'Shell into a running service',
        lucity: { mark: 'no' },
      },
      {
        key: 'iac',
        name: 'Config as code in your repository',
        lucity: { mark: 'no', text: 'Lucity never writes to or reads config from your repository; you can download the deployed config' },
      },
    ],
  },
  {
    title: 'Team',
    rows: [
      {
        key: 'ws',
        name: 'Workspaces and members',
        lucity: { mark: 'yes', text: 'admins and members, on Pro' },
      },
      {
        key: 'sso',
        name: 'Single sign-on',
        lucity: { mark: 'no' },
      },
      {
        key: 'audit',
        name: 'Audit log',
        lucity: { mark: 'no' },
      },
    ],
  },
  {
    title: 'Billing',
    rows: [
      {
        key: 'usage',
        name: 'Billed on actual use',
        lucity: { mark: 'yes', text: 'CPU and memory as used' },
      },
      {
        key: 'credit',
        name: 'Plan fee comes back as credit',
        lucity: { mark: 'yes', text: '€5 on Hobby, €25 on Pro' },
      },
      {
        key: 'spend',
        name: 'Spending limits',
        lucity: { mark: 'no' },
      },
      {
        key: 'trial',
        name: 'Free trial',
        lucity: { mark: 'yes', text: '14 days or €5 of usage, no card' },
      },
      {
        key: 'free',
        name: 'Free tier',
        lucity: { mark: 'no' },
      },
    ],
  },
  {
    title: 'Security',
    rows: [
      {
        key: 'secrets',
        name: 'Secret scanning',
        lucity: { mark: 'yes', text: 'every deploy checks the repository and its history' },
      },
      {
        key: 'cve',
        name: 'Vulnerability checks',
        lucity: { mark: 'yes', text: 'running apps rechecked every two hours' },
      },
      {
        key: 'isolation',
        name: 'Workspace isolation',
        lucity: { mark: 'yes' },
      },
      {
        key: 'envisolation',
        name: 'Environment isolation',
        lucity: { mark: 'yes', text: "staging can't reach production" },
      },
      {
        key: 'soc2',
        name: 'SOC 2 report',
        lucity: { mark: 'no' },
      },
    ],
  },
  {
    title: 'Where it runs and how you leave',
    rows: [
      {
        key: 'regions',
        name: 'Regions',
        lucity: { mark: 'partly', label: 'One', text: 'Germany, with files and backups in France' },
      },
      {
        key: 'eu',
        name: 'Hosting in the EU',
        lucity: { mark: 'yes', text: 'apps and databases in Germany' },
      },
      {
        key: 'swiss',
        name: 'Hosting in Switzerland',
        lucity: { mark: 'no', text: 'Germany and France' },
      },
      {
        key: 'company',
        name: 'Company based in',
        lucity: { mark: 'info', label: 'Switzerland' },
      },
      {
        key: 'oss',
        name: 'Open source',
        lucity: { mark: 'yes', text: 'AGPL-3.0' },
      },
      {
        key: 'selfhost',
        name: 'Self-hosting',
        lucity: { mark: 'yes', text: 'the whole platform, on your own Kubernetes' },
      },
      {
        key: 'eject',
        name: 'Export a project to run elsewhere',
        lucity: { mark: 'yes', text: 'a standard Helm chart and values, per environment' },
      },
    ],
  },
];
