import gql from 'graphql-tag';

export const SubscriptionQuery = gql`
  query Subscription {
    subscription {
      plan
      status
      currentPeriodEnd
      creditAmountCents
    }
  }
`;

export const UsageSummaryQuery = gql`
  query UsageSummary {
    usageSummary {
      resourceCostCents
      creditsCents
      estimatedTotalCents
    }
  }
`;

export const ChangePlanMutation = gql`
  mutation ChangePlan($plan: Plan!) {
    changePlan(plan: $plan) {
      plan
      status
      currentPeriodEnd
      creditAmountCents
    }
  }
`;

export const BillingPortalUrlMutation = gql`
  mutation BillingPortalUrl {
    billingPortalUrl {
      url
    }
  }
`;

export const EnvironmentResourcesQuery = gql`
  query EnvironmentResources($projectId: ID!, $environment: String!) {
    environmentResources(projectId: $projectId, environment: $environment) {
      tier
      allocation {
        cpuMillicores
        memoryMB
        diskMB
      }
    }
  }
`;

export const SetEnvironmentResourcesMutation = gql`
  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {
    setEnvironmentResources(input: $input) {
      tier
      allocation {
        cpuMillicores
        memoryMB
        diskMB
      }
    }
  }
`;
