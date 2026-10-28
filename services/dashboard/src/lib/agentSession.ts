import {
  AgentApprovalStatus,
  AgentSessionStatus,
  AgentTurnEndReason,
  type AgentToolCallStatus,
} from '@/gql/graphql';

export interface SessionMember {
  name?: string | null;
  email: string;
}

export interface UserMessageEvent {
  __typename: 'AgentUserMessage';
  id: string;
  time: string;
  text: string;
  author?: SessionMember | null;
}

export interface MessageEvent {
  __typename: 'AgentMessage';
  id: string;
  time: string;
  text: string;
}

export interface ToolCallEvent {
  __typename: 'AgentToolCall';
  id: string;
  time: string;
  callId: string;
  connector?: string | null;
  tool: string;
  title: string;
  status: AgentToolCallStatus;
  input?: string | null;
  output?: string | null;
}

export interface ApprovalEvent {
  __typename: 'AgentApproval';
  id: string;
  time: string;
  approval: string;
  toolCallId?: string | null;
  title: string;
  approvalStatus: AgentApprovalStatus;
  decidedBy?: SessionMember | null;
}

export interface TurnEndEvent {
  __typename: 'AgentTurnEnd';
  id: string;
  time: string;
  reason: AgentTurnEndReason;
  error?: string | null;
}

export type SessionEvent = UserMessageEvent | MessageEvent | ToolCallEvent | ApprovalEvent | TurnEndEvent;

export type TranscriptItem =
  | { kind: 'message'; key: string; event: MessageEvent }
  | { kind: 'tools'; key: string; calls: ToolCallEvent[] }
  | { kind: 'approval'; key: string; event: ApprovalEvent; call: ToolCallEvent | null }
  | { kind: 'end'; key: string; event: TurnEndEvent };

export type TranscriptBlock =
  | { kind: 'user'; key: string; event: UserMessageEvent }
  | { kind: 'agent'; key: string; time: string; items: TranscriptItem[] };

export function memberName(member?: SessionMember | null): string {
  return member?.name || member?.email || 'A former member';
}

function eventKey(event: SessionEvent): string {
  switch (event.__typename) {
    case 'AgentToolCall':
      return `call:${event.callId}`;
    case 'AgentApproval':
      return `approval:${event.approval}`;
    default:
      return `event:${event.id}`;
  }
}

export function collapseEvents(events: SessionEvent[]): SessionEvent[] {
  const order: string[] = [];
  const latest = new Map<string, SessionEvent>();
  for (const event of events) {
    const key = eventKey(event);
    if (!latest.has(key)) order.push(key);
    latest.set(key, event);
  }
  return order.map(key => latest.get(key)!);
}

export function pendingApprovals(events: SessionEvent[]): ApprovalEvent[] {
  return collapseEvents(events).filter(
    (event): event is ApprovalEvent => event.__typename === 'AgentApproval' && event.approvalStatus === AgentApprovalStatus.Pending,
  );
}

export function sessionActivity(events: SessionEvent[]): AgentSessionStatus {
  if (pendingApprovals(events).length > 0) return AgentSessionStatus.WaitingForApproval;
  const last = events.at(-1);
  if (!last || last.__typename === 'AgentTurnEnd') return AgentSessionStatus.Idle;
  return AgentSessionStatus.Working;
}

export function buildTranscript(events: SessionEvent[]): TranscriptBlock[] {
  const collapsed = collapseEvents(events);
  const calls = new Map<string, ToolCallEvent>();
  const awaiting = new Set<string>();
  for (const event of collapsed) {
    if (event.__typename === 'AgentToolCall') calls.set(event.callId, event);
    if (event.__typename === 'AgentApproval' && event.approvalStatus === AgentApprovalStatus.Pending && event.toolCallId) {
      awaiting.add(event.toolCallId);
    }
  }

  const blocks: TranscriptBlock[] = [];

  function agentItems(event: SessionEvent): TranscriptItem[] {
    const last = blocks.at(-1);
    if (last?.kind === 'agent') return last.items;
    const block: TranscriptBlock = { kind: 'agent', key: `block:${eventKey(event)}`, time: event.time, items: [] };
    blocks.push(block);
    return block.items;
  }

  for (const event of collapsed) {
    const key = eventKey(event);
    switch (event.__typename) {
      case 'AgentUserMessage':
        blocks.push({ kind: 'user', key, event });
        break;
      case 'AgentMessage':
        agentItems(event).push({ kind: 'message', key, event });
        break;
      case 'AgentToolCall': {
        if (awaiting.has(event.callId)) break;
        const items = agentItems(event);
        const last = items.at(-1);
        if (last?.kind === 'tools') last.calls.push(event);
        else items.push({ kind: 'tools', key, calls: [event] });
        break;
      }
      case 'AgentApproval':
        agentItems(event).push({ kind: 'approval', key, event, call: event.toolCallId ? calls.get(event.toolCallId) ?? null : null });
        break;
      case 'AgentTurnEnd':
        if (event.reason !== AgentTurnEndReason.Completed) agentItems(event).push({ kind: 'end', key, event });
        break;
    }
  }

  return blocks;
}

export function prettyJson(value?: string | null): string {
  if (!value) return '';
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch {
    return value;
  }
}
