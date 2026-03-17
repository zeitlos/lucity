import gql from 'graphql-tag';

export const WorkspacesQuery = gql`
  query Workspaces {
    workspaces {
      id
      name
      personal
    }
  }
`;

export const WorkspaceQuery = gql`
  query Workspace {
    workspace {
      id
      name
      personal
      suspended
      members {
        id
        email
        name
        role
      }
    }
  }
`;

export const CreateWorkspaceMutation = gql`
  mutation CreateWorkspace($input: CreateWorkspaceInput!) {
    createWorkspace(input: $input) {
      id
      name
      personal
    }
  }
`;

export const CreateWorkspaceCheckoutMutation = gql`
  mutation CreateWorkspaceCheckout($input: CreateWorkspaceCheckoutInput!) {
    createWorkspaceCheckout(input: $input) {
      url
    }
  }
`;

export const CompleteWorkspaceCheckoutMutation = gql`
  mutation CompleteWorkspaceCheckout($sessionId: String!) {
    completeWorkspaceCheckout(sessionId: $sessionId) {
      id
      name
      personal
    }
  }
`;

export const UpdateWorkspaceMutation = gql`
  mutation UpdateWorkspace($input: UpdateWorkspaceInput!) {
    updateWorkspace(input: $input) {
      id
      name
    }
  }
`;

export const DeleteWorkspaceMutation = gql`
  mutation DeleteWorkspace {
    deleteWorkspace
  }
`;

export const InviteMemberMutation = gql`
  mutation InviteMember($input: InviteMemberInput!) {
    inviteMember(input: $input) {
      id
      email
      name
      role
    }
  }
`;

export const RemoveMemberMutation = gql`
  mutation RemoveMember($userId: ID!) {
    removeMember(userId: $userId)
  }
`;

export const UpdateMemberRoleMutation = gql`
  mutation UpdateMemberRole($input: UpdateMemberRoleInput!) {
    updateMemberRole(input: $input) {
      id
      email
      name
      role
    }
  }
`;
