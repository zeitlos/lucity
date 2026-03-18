import gql from 'graphql-tag';

export const SubscriptionQuery = gql`
  query Subscription {
    subscription {
      plan
      status
      currentPeriodEnd
      creditAmountCents
      creditExpiry
      hasPaymentMethod
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
      creditExpiry
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

export const CreatePlanCheckoutMutation = gql`
  mutation CreatePlanCheckout($plan: Plan!) {
    createPlanCheckout(plan: $plan) {
      url
    }
  }
`;

export const CompletePlanCheckoutMutation = gql`
  mutation CompletePlanCheckout($sessionId: String!) {
    completePlanCheckout(sessionId: $sessionId) {
      plan
      status
      currentPeriodEnd
      creditAmountCents
      hasPaymentMethod
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
