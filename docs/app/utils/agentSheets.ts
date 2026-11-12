export interface AgentSheet {
  slug: string;
  title: string;
  intro: string;
  caption: string;
  sections: { title: string; text: string }[];
  table: {
    title: string;
    head: [string, string];
    rows: { icon: string; name: string; region?: string; text: string }[];
  };
}

export const agentSheets: AgentSheet[] = [
  {
    slug: 'triggers',
    title: 'Triggers',
    intro: 'A trigger starts a conversation with an agent. It can be something that happens on Lucity or in one of your tools, a schedule, a webhook or an email. The agent sleeps until one fires and is at work seconds later.',
    caption: 'Forty restarts, one thread',
    sections: [
      { title: 'One thread per incident', text: 'Events about the same thing land in the same conversation. A service that restarts forty times in ten minutes opens one thread, and every restart after the first is added to it instead of waking the agent again. Lucity events group by service and kind, webhooks and emails set their own key.' },
      { title: 'What the agent is told', text: 'The event arrives as the first message, in plain words with the details attached: the service and environment, the last lines of its logs, the deploy and its commit, and a link back to the dashboard.' },
      { title: 'Quiet hours and limits', text: 'Each trigger can be paused, kept quiet outside the hours you set and limited to a number of runs per hour. A trigger that fires while its limit is spent is added to the thread and waits there.' },
    ],
    table: {
      title: 'Every trigger',
      head: ['Source', 'Events'],
      rows: [
        { icon: 'lucity', name: 'Lucity', text: 'Deploy failed, restart loop, out of memory, disk almost full, new vulnerability, certificate failed, backup failed' },
        { icon: 'github', name: 'GitHub', text: 'Issue opened, pull request opened, review requested, workflow failed' },
        { icon: 'sentry', name: 'Sentry', text: 'New issue, regression, spike in errors' },
        { icon: 'grafana', name: 'Grafana', text: 'Alert firing, alert resolved' },
        { icon: 'linear', name: 'Linear', text: 'Issue assigned to the agent, mention in a comment' },
        { icon: 'clock', name: 'Schedule', text: 'Any schedule, in plain words or cron, in your time zone' },
        { icon: 'webhook', name: 'Webhook', text: 'Any JSON body, signed with HMAC SHA-256' },
        { icon: 'mail', name: 'Email', text: "The agent's own address, attachments included" },
        { icon: 'bell', name: 'Chat', text: 'A message or a mention in a linked chat' },
      ],
    },
  },
  {
    slug: 'connectors',
    title: 'Connectors',
    intro: "A connector gives an agent tools. Any remote MCP server becomes one, and so does your own API. The agent calls a tool and Lucity adds the key on the way out, so no key ever enters the agent's sandbox.",
    caption: 'Allowed, asks first or hidden, tool by tool',
    sections: [
      { title: 'Keys stay outside', text: 'Credentials belong to the workspace, not to an agent. Every request leaves the sandbox through Lucity, which adds the key for that connector and nothing else. Rotate a key once and every agent uses the new one.' },
      { title: 'Three settings per tool', text: 'Allowed runs right away. Asks first waits for a yes in the chat or the dashboard. Hidden is never shown to the model. Tools a server adds later start as asks first.' },
      { title: 'Hosts you allow', text: 'Beyond its connectors, an agent reaches only the hosts you list. Package registries are open by default, so it can install what it needs. Everything else stays closed until you add it.' },
    ],
    table: {
      title: 'Every connector',
      head: ['Connector', 'Tools'],
      rows: [
        { icon: 'lucity', name: 'Lucity', text: 'Logs, metrics, deploys, rollbacks, variables and read-only database queries' },
        { icon: 'github', name: 'GitHub', text: 'Issues, pull requests, code search, Actions runs' },
        { icon: 'linear', name: 'Linear', text: 'Issues, projects, comments' },
        { icon: 'jira', name: 'Jira', text: 'Issues, transitions, comments' },
        { icon: 'sentry', name: 'Sentry', text: 'Issues, events, releases' },
        { icon: 'grafana', name: 'Grafana', text: 'Dashboards, alerts, queries' },
        { icon: 'notion', name: 'Notion', text: 'Pages, databases, search' },
        { icon: 'modelcontextprotocol', name: 'Any MCP server', text: 'Remote servers over HTTP, with a header or an OAuth sign-in' },
        { icon: 'server', name: 'Your own API', text: 'One tool per operation, from its OpenAPI file' },
      ],
    },
  },
  {
    slug: 'models',
    title: 'Model providers',
    intro: 'Agents think with the providers you already pay for. Connect a provider once with your own key. Each conversation picks a model and how hard it thinks, and the bill stays with the provider, at their price.',
    caption: 'One agent, its model and its limit',
    sections: [
      { title: 'No markup on tokens', text: 'Model calls go to the provider on your key. Lucity bills the agent like a service, for the memory and CPU it uses while it works, and nothing for tokens.' },
      { title: 'A daily limit per agent', text: 'Set a spend limit in your currency. An agent that reaches it finishes its current step, says so in the thread and waits for the next day, or for you to raise the limit.' },
      { title: 'Work that stays in Europe', text: 'Pair an agent with a provider in the EU or Switzerland and the whole loop stays there, from the trigger to the last tool call.' },
    ],
    table: {
      title: 'Every model provider',
      head: ['Provider', 'Models'],
      rows: [
        { icon: 'anthropic', name: 'Anthropic', region: 'United States', text: 'Claude Opus, Sonnet and Haiku' },
        { icon: 'openai', name: 'OpenAI', region: 'United States', text: 'GPT models, the reasoning ones included' },
        { icon: 'mistralai', name: 'Mistral', region: 'France', text: 'Mistral Large, Medium and Codestral' },
        { icon: 'infomaniak', name: 'Infomaniak', region: 'Switzerland', text: 'Open models such as Llama, Qwen and Mistral' },
        { icon: 'scaleway', name: 'Scaleway', region: 'France', text: 'Open models such as Llama, Qwen and Mistral' },
        { icon: 'server', name: 'Your own', region: 'Where you run it', text: 'Any OpenAI-compatible API, vLLM and Ollama included' },
      ],
    },
  },
  {
    slug: 'channels',
    title: 'Channels',
    intro: 'An agent answers where your team already talks. Add it to a chat and each conversation becomes a thread there, with approvals as buttons. Everything it does is in the dashboard too.',
    caption: 'An approval, in the thread',
    sections: [
      { title: 'It knows who is asking', text: "Teammates link their chat account to Lucity once. The agent then knows who asked, acts with that person's role and records who said yes." },
      { title: 'A thread per conversation', text: 'Each conversation is a thread in Slack, Mattermost and Teams, and a topic in a Telegram group. Replies in the thread go to the agent. The rest of the channel stays yours.' },
      { title: 'On record in the dashboard', text: 'Every conversation is also in the dashboard and the API, with each step, tool call and approval, whichever chat it started in.' },
    ],
    table: {
      title: 'Every channel',
      head: ['Channel', 'Threads and approvals'],
      rows: [
        { icon: 'telegram', name: 'Telegram', text: 'A topic per conversation in a group, or a direct chat. Approvals as buttons.' },
        { icon: 'mattermost', name: 'Mattermost', text: 'A thread per conversation. Approvals as buttons.' },
        { icon: 'slack', name: 'Slack', text: 'A thread per conversation. Approvals as buttons.' },
        { icon: 'microsoftteams', name: 'Microsoft Teams', text: 'A reply chain per conversation. Approvals as buttons.' },
        { icon: 'mail', name: 'Email', text: 'One mail thread per conversation. Approve by replying yes.' },
        { icon: 'layout-dashboard', name: 'Dashboard and API', text: 'Every conversation, step and approval, from every channel' },
      ],
    },
  },
];
