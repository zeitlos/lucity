type MinimarkNode = string | [string, Record<string, unknown>, ...MinimarkNode[]];

function stripLinkRel(nodes: MinimarkNode[]) {
  for (const node of nodes) {
    if (typeof node === 'string') continue;
    const [tag, props, ...children] = node;
    if (tag === 'a') delete props?.rel;
    stripLinkRel(children);
  }
}

export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('content:llms:generate:document', (_event, doc) => {
    stripLinkRel((doc.body as unknown as { value: MinimarkNode[] }).value ?? []);
  });
});
