/* eslint-disable */
import * as types from './graphql';
import type { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';

/**
 * Map of all GraphQL operations in the project.
 *
 * This map has several performance disadvantages:
 * 1. It is not tree-shakeable, so it will include all operations in the project.
 * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.
 * 3. It does not support dead code elimination, so it will add unused operations.
 *
 * Therefore it is highly recommended to use the babel or swc plugin for production.
 * Learn more about it here: https://the-guild.dev/graphql/codegen/plugins/presets/preset-client#reducing-bundle-size
 */
type Documents = {
    "\n  query Environment {\n    environment\n  }\n": typeof types.EnvironmentDocument,
    "\n  query EnvironmentLogs($id: ID!) {\n    environmentLogs(id: $id)\n  }\n": typeof types.EnvironmentLogsDocument,
    "\n  query Trends($since: Time!) {\n    signups(since: $since) { start count }\n    workspaceStandings(since: $since) { at good billing abuse }\n  }\n": typeof types.TrendsDocument,
    "\n  fragment WorkspaceFields on Workspace {\n    id\n    createdAt\n    abuse\n    abuseNote\n    suspended\n    billing { plan trialEndsAt goodStanding standingReason hasPaymentMethod }\n    members { name email role account { login createdAt } }\n    signals { kind strong detail }\n    runningPods\n    pendingPods\n    memoryRequested\n    egress7d\n    environments {\n      id\n      project\n      name\n      runningPods\n      pendingPods\n      services { name image commit repository }\n    }\n  }\n": typeof types.WorkspaceFieldsFragmentDoc,
    "\n  query Workspaces {\n    workspaces { loadedAt workspaces { ...WorkspaceFields } }\n  }\n": typeof types.WorkspacesDocument,
    "\n  query CachedWorkspaces {\n    cachedWorkspaces { loadedAt workspaces { ...WorkspaceFields } }\n  }\n": typeof types.CachedWorkspacesDocument,
    "\n  mutation MarkAbuse($workspaceIDs: [ID!]!, $note: String) {\n    markAbuse(workspaceIDs: $workspaceIDs, note: $note) { workspaceID error }\n  }\n": typeof types.MarkAbuseDocument,
    "\n  mutation ClearAbuse($workspaceIDs: [ID!]!) {\n    clearAbuse(workspaceIDs: $workspaceIDs) { workspaceID error }\n  }\n": typeof types.ClearAbuseDocument,
};
const documents: Documents = {
    "\n  query Environment {\n    environment\n  }\n": types.EnvironmentDocument,
    "\n  query EnvironmentLogs($id: ID!) {\n    environmentLogs(id: $id)\n  }\n": types.EnvironmentLogsDocument,
    "\n  query Trends($since: Time!) {\n    signups(since: $since) { start count }\n    workspaceStandings(since: $since) { at good billing abuse }\n  }\n": types.TrendsDocument,
    "\n  fragment WorkspaceFields on Workspace {\n    id\n    createdAt\n    abuse\n    abuseNote\n    suspended\n    billing { plan trialEndsAt goodStanding standingReason hasPaymentMethod }\n    members { name email role account { login createdAt } }\n    signals { kind strong detail }\n    runningPods\n    pendingPods\n    memoryRequested\n    egress7d\n    environments {\n      id\n      project\n      name\n      runningPods\n      pendingPods\n      services { name image commit repository }\n    }\n  }\n": types.WorkspaceFieldsFragmentDoc,
    "\n  query Workspaces {\n    workspaces { loadedAt workspaces { ...WorkspaceFields } }\n  }\n": types.WorkspacesDocument,
    "\n  query CachedWorkspaces {\n    cachedWorkspaces { loadedAt workspaces { ...WorkspaceFields } }\n  }\n": types.CachedWorkspacesDocument,
    "\n  mutation MarkAbuse($workspaceIDs: [ID!]!, $note: String) {\n    markAbuse(workspaceIDs: $workspaceIDs, note: $note) { workspaceID error }\n  }\n": types.MarkAbuseDocument,
    "\n  mutation ClearAbuse($workspaceIDs: [ID!]!) {\n    clearAbuse(workspaceIDs: $workspaceIDs) { workspaceID error }\n  }\n": types.ClearAbuseDocument,
};

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 *
 *
 * @example
 * ```ts
 * const query = graphql(`query GetUser($id: ID!) { user(id: $id) { name } }`);
 * ```
 *
 * The query argument is unknown!
 * Please regenerate the types.
 */
export function graphql(source: string): unknown;

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Environment {\n    environment\n  }\n"): (typeof documents)["\n  query Environment {\n    environment\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query EnvironmentLogs($id: ID!) {\n    environmentLogs(id: $id)\n  }\n"): (typeof documents)["\n  query EnvironmentLogs($id: ID!) {\n    environmentLogs(id: $id)\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Trends($since: Time!) {\n    signups(since: $since) { start count }\n    workspaceStandings(since: $since) { at good billing abuse }\n  }\n"): (typeof documents)["\n  query Trends($since: Time!) {\n    signups(since: $since) { start count }\n    workspaceStandings(since: $since) { at good billing abuse }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  fragment WorkspaceFields on Workspace {\n    id\n    createdAt\n    abuse\n    abuseNote\n    suspended\n    billing { plan trialEndsAt goodStanding standingReason hasPaymentMethod }\n    members { name email role account { login createdAt } }\n    signals { kind strong detail }\n    runningPods\n    pendingPods\n    memoryRequested\n    egress7d\n    environments {\n      id\n      project\n      name\n      runningPods\n      pendingPods\n      services { name image commit repository }\n    }\n  }\n"): (typeof documents)["\n  fragment WorkspaceFields on Workspace {\n    id\n    createdAt\n    abuse\n    abuseNote\n    suspended\n    billing { plan trialEndsAt goodStanding standingReason hasPaymentMethod }\n    members { name email role account { login createdAt } }\n    signals { kind strong detail }\n    runningPods\n    pendingPods\n    memoryRequested\n    egress7d\n    environments {\n      id\n      project\n      name\n      runningPods\n      pendingPods\n      services { name image commit repository }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Workspaces {\n    workspaces { loadedAt workspaces { ...WorkspaceFields } }\n  }\n"): (typeof documents)["\n  query Workspaces {\n    workspaces { loadedAt workspaces { ...WorkspaceFields } }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query CachedWorkspaces {\n    cachedWorkspaces { loadedAt workspaces { ...WorkspaceFields } }\n  }\n"): (typeof documents)["\n  query CachedWorkspaces {\n    cachedWorkspaces { loadedAt workspaces { ...WorkspaceFields } }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation MarkAbuse($workspaceIDs: [ID!]!, $note: String) {\n    markAbuse(workspaceIDs: $workspaceIDs, note: $note) { workspaceID error }\n  }\n"): (typeof documents)["\n  mutation MarkAbuse($workspaceIDs: [ID!]!, $note: String) {\n    markAbuse(workspaceIDs: $workspaceIDs, note: $note) { workspaceID error }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation ClearAbuse($workspaceIDs: [ID!]!) {\n    clearAbuse(workspaceIDs: $workspaceIDs) { workspaceID error }\n  }\n"): (typeof documents)["\n  mutation ClearAbuse($workspaceIDs: [ID!]!) {\n    clearAbuse(workspaceIDs: $workspaceIDs) { workspaceID error }\n  }\n"];

export function graphql(source: string) {
  return (documents as any)[source] ?? {};
}

export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<  infer TType,  any>  ? TType  : never;