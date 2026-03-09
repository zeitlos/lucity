import gql from 'graphql-tag';

export const WorkspacesQuery = gql`
  query Workspaces {
    workspaces {
      id
      name
      personal
      githubLinked
    }
  }
`;

export const WorkspaceQuery = gql`
  query Workspace {
    workspace {
      id
      name
      personal
      githubLinked
      githubInstallUrl
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
      githubLinked
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

export const LinkGithubInstallationMutation = gql`
  mutation LinkGithubInstallation($installationId: ID!) {
    linkGithubInstallation(installationId: $installationId) {
      id
      githubLinked
    }
  }
`;

export const UnlinkGithubInstallationMutation = gql`
  mutation UnlinkGithubInstallation {
    unlinkGithubInstallation {
      id
      githubLinked
      githubInstallUrl
    }
  }
`;
