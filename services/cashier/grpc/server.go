package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	gostripe "github.com/stripe/stripe-go/v82"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/cashier"
	"github.com/zeitlos/lucity/pkg/deployer"
	stripelib "github.com/zeitlos/lucity/services/cashier/stripe"
)

type Server struct {
	cashier.UnimplementedCashierServiceServer
	stripe   *stripelib.Client
	deployer deployer.DeployerServiceClient
}

func NewServer(stripeClient *stripelib.Client, deployerClient deployer.DeployerServiceClient) *Server {
	return &Server{
		stripe:   stripeClient,
		deployer: deployerClient,
	}
}

// deployerCtx creates a gRPC context for calling the deployer as a system-level caller.
func deployerCtx(ctx context.Context) context.Context {
	ctx = auth.WithClaims(ctx, &auth.Claims{
		Subject: "cashier",
		Roles:   []auth.Role{auth.RoleUser},
	})
	return auth.OutgoingContext(ctx)
}

func (s *Server) CreateCustomer(ctx context.Context, req *cashier.CreateCustomerRequest) (*cashier.CreateCustomerResponse, error) {
	if req.Workspace == "" {
		return nil, fmt.Errorf("workspace required")
	}

	customerID, err := s.stripe.CreateCustomer(ctx, req.Workspace, req.Name, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Store customer ID in workspace ConfigMap via deployer
	_, err = s.deployer.UpdateWorkspaceMetadata(deployerCtx(ctx), &deployer.UpdateWorkspaceMetadataRequest{
		Workspace:        req.Workspace,
		StripeCustomerId: customerID,
	})
	if err != nil {
		slog.Error("failed to store stripe customer ID", "workspace", req.Workspace, "error", err)
		// Don't fail — customer was created in Stripe, we can recover
	}

	slog.Info("stripe customer created", "workspace", req.Workspace, "customer_id", customerID)
	return &cashier.CreateCustomerResponse{CustomerId: customerID}, nil
}

func (s *Server) CreateSubscription(ctx context.Context, req *cashier.CreateSubscriptionRequest) (*cashier.CreateSubscriptionResponse, error) {
	if req.Workspace == "" || req.CustomerId == "" {
		return nil, fmt.Errorf("workspace and customer_id required")
	}

	planPriceID := s.stripe.PlanPriceID(planToString(req.Plan))

	subID, err := s.stripe.CreateSubscription(ctx, req.CustomerId, req.Workspace, planPriceID, int(req.TrialDays))
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Store subscription ID in workspace ConfigMap via deployer
	_, err = s.deployer.UpdateWorkspaceMetadata(deployerCtx(ctx), &deployer.UpdateWorkspaceMetadataRequest{
		Workspace:            req.Workspace,
		StripeSubscriptionId: subID,
	})
	if err != nil {
		slog.Error("failed to store stripe subscription ID", "workspace", req.Workspace, "error", err)
	}

	slog.Info("stripe subscription created", "workspace", req.Workspace, "subscription_id", subID)
	return &cashier.CreateSubscriptionResponse{SubscriptionId: subID}, nil
}

func (s *Server) ChangePlan(ctx context.Context, req *cashier.ChangePlanRequest) (*cashier.ChangePlanResponse, error) {
	meta, err := s.deployer.WorkspaceMetadata(deployerCtx(ctx), &deployer.WorkspaceMetadataRequest{
		Workspace: req.Workspace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace metadata: %w", err)
	}
	if meta.StripeSubscriptionId == "" {
		return nil, fmt.Errorf("no subscription found for workspace %q", req.Workspace)
	}

	newPriceID := s.stripe.PlanPriceID(planToString(req.Plan))

	// Determine old plan price ID (the other one)
	oldPriceID := s.stripe.Prices.HobbyPriceID
	if newPriceID == s.stripe.Prices.HobbyPriceID {
		oldPriceID = s.stripe.Prices.ProPriceID
	}

	sub, err := s.stripe.ChangePlan(ctx, meta.StripeSubscriptionId, newPriceID, oldPriceID)
	if err != nil {
		return nil, fmt.Errorf("failed to change plan: %w", err)
	}

	hasPM, _ := s.stripe.HasPaymentMethod(ctx, meta.StripeCustomerId)

	slog.Info("plan changed", "workspace", req.Workspace, "plan", planToString(req.Plan))
	resp := subscriptionToResponse(sub, s.stripe.Prices)
	resp.HasPaymentMethod = hasPM
	return resp, nil
}

func (s *Server) Subscription(ctx context.Context, req *cashier.SubscriptionRequest) (*cashier.SubscriptionResponse, error) {
	meta, err := s.deployer.WorkspaceMetadata(deployerCtx(ctx), &deployer.WorkspaceMetadataRequest{
		Workspace: req.Workspace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace metadata: %w", err)
	}
	if meta.StripeSubscriptionId == "" {
		return nil, fmt.Errorf("no subscription found for workspace %q", req.Workspace)
	}

	sub, err := s.stripe.Subscription(ctx, meta.StripeSubscriptionId)
	if err != nil {
		return nil, err
	}

	hasPM, _ := s.stripe.HasPaymentMethod(ctx, meta.StripeCustomerId)

	resp := subscriptionToResponse(sub, s.stripe.Prices)
	return &cashier.SubscriptionResponse{
		Plan:              resp.Plan,
		Status:            resp.Status,
		CurrentPeriodEnd:  resp.CurrentPeriodEnd,
		CreditAmountCents: resp.CreditAmountCents,
		TrialEnd:          resp.TrialEnd,
		HasPaymentMethod:  hasPM,
	}, nil
}

func (s *Server) BillingPortalURL(ctx context.Context, req *cashier.BillingPortalURLRequest) (*cashier.BillingPortalURLResponse, error) {
	meta, err := s.deployer.WorkspaceMetadata(deployerCtx(ctx), &deployer.WorkspaceMetadataRequest{
		Workspace: req.Workspace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace metadata: %w", err)
	}
	if meta.StripeCustomerId == "" {
		return nil, fmt.Errorf("no billing customer found for workspace %q", req.Workspace)
	}

	url, err := s.stripe.BillingPortalURL(ctx, meta.StripeCustomerId, req.ReturnUrl)
	if err != nil {
		return nil, err
	}

	return &cashier.BillingPortalURLResponse{Url: url}, nil
}

func (s *Server) UsageSummary(ctx context.Context, req *cashier.UsageSummaryRequest) (*cashier.UsageSummaryResponse, error) {
	meta, err := s.deployer.WorkspaceMetadata(deployerCtx(ctx), &deployer.WorkspaceMetadataRequest{
		Workspace: req.Workspace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace metadata: %w", err)
	}
	if meta.StripeCustomerId == "" {
		return nil, fmt.Errorf("no billing customer found for workspace %q", req.Workspace)
	}

	inv, err := s.stripe.UpcomingInvoice(ctx, meta.StripeCustomerId, meta.StripeSubscriptionId)
	if err != nil {
		// No upcoming invoice (no usage yet) is fine
		return &cashier.UsageSummaryResponse{}, nil
	}

	var resourceCost int64
	var planCost int64
	for _, line := range inv.Lines.Data {
		if line.Pricing == nil || line.Pricing.PriceDetails == nil {
			continue
		}
		priceID := line.Pricing.PriceDetails.Price
		if priceID == s.stripe.Prices.HobbyPriceID || priceID == s.stripe.Prices.ProPriceID {
			planCost = line.Amount
		} else {
			resourceCost += line.Amount
		}
	}

	// Credits are handled by Stripe Credit Grants — query the credit balance.
	creditBalance, err := s.stripe.CreditBalanceCents(ctx, meta.StripeCustomerId)
	if err != nil {
		slog.Warn("failed to get credit balance", "workspace", req.Workspace, "error", err)
	}
	// creditBalance is the remaining balance; credits applied = plan credit - remaining.
	planCredit := stripelib.PlanCreditCents(s.stripe.Prices.HobbyPriceID, s.stripe.Prices)
	if meta.StripeSubscriptionId != "" {
		sub, subErr := s.stripe.Subscription(ctx, meta.StripeSubscriptionId)
		if subErr == nil {
			for _, item := range sub.Items.Data {
				if item.Price.ID == s.stripe.Prices.ProPriceID {
					planCredit = stripelib.PlanCreditCents(s.stripe.Prices.ProPriceID, s.stripe.Prices)
					break
				}
			}
		}
	}
	creditsApplied := planCredit - creditBalance
	if creditsApplied < 0 {
		creditsApplied = 0
	}

	// Estimated total = plan + resources - credits applied
	estimated := planCost + resourceCost - creditsApplied

	return &cashier.UsageSummaryResponse{
		ResourceCostCents:   int32(resourceCost),
		CreditsCents:        int32(creditsApplied),
		EstimatedTotalCents: int32(estimated),
	}, nil
}

// HandleStripeEvent processes webhook events that need workspace ConfigMap updates.
func (s *Server) HandleStripeEvent(event gostripe.Event) {
	switch event.Type {
	case "invoice.payment_failed":
		s.handlePaymentFailed(event)
	case "invoice.payment_succeeded":
		s.handlePaymentSucceeded(event)
	case "customer.subscription.deleted":
		s.handleSubscriptionDeleted(event)
	}
}

func (s *Server) handlePaymentFailed(event gostripe.Event) {
	var inv gostripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		slog.Error("failed to unmarshal invoice for payment_failed", "error", err)
		return
	}

	workspace := ""
	if inv.Parent != nil && inv.Parent.SubscriptionDetails != nil && inv.Parent.SubscriptionDetails.Subscription != nil {
		workspace = inv.Parent.SubscriptionDetails.Subscription.Metadata["workspace"]
	}
	if workspace == "" {
		slog.Warn("payment failed but no workspace in subscription metadata", "invoice", inv.ID)
		return
	}

	slog.Error("payment failed, suspending workspace", "workspace", workspace, "invoice", inv.ID)
	s.suspendWorkspace(workspace, true)
}

func (s *Server) handlePaymentSucceeded(event gostripe.Event) {
	var inv gostripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		slog.Error("failed to unmarshal invoice for payment_succeeded", "error", err)
		return
	}

	workspace := ""
	if inv.Parent != nil && inv.Parent.SubscriptionDetails != nil && inv.Parent.SubscriptionDetails.Subscription != nil {
		workspace = inv.Parent.SubscriptionDetails.Subscription.Metadata["workspace"]
	}
	if workspace == "" {
		return
	}

	// Check if workspace is suspended; if so, resume it.
	ctx := deployerCtx(context.Background())
	meta, err := s.deployer.WorkspaceMetadata(ctx, &deployer.WorkspaceMetadataRequest{Workspace: workspace})
	if err != nil {
		slog.Warn("failed to check workspace suspension on payment success", "workspace", workspace, "error", err)
		return
	}
	if !meta.Suspended {
		return
	}

	slog.Info("payment succeeded, resuming workspace", "workspace", workspace, "invoice", inv.ID)
	s.suspendWorkspace(workspace, false)
}

func (s *Server) handleSubscriptionDeleted(event gostripe.Event) {
	var sub gostripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		slog.Error("failed to unmarshal subscription for deletion", "error", err)
		return
	}

	workspace := sub.Metadata["workspace"]
	if workspace == "" {
		slog.Warn("subscription deleted but no workspace in metadata", "subscription", sub.ID)
		return
	}

	slog.Warn("subscription deleted, suspending workspace", "workspace", workspace, "subscription", sub.ID)
	s.suspendWorkspace(workspace, true)
}

func (s *Server) suspendWorkspace(workspace string, suspended bool) {
	ctx := deployerCtx(context.Background())
	_, err := s.deployer.SuspendWorkspace(ctx, &deployer.SuspendWorkspaceRequest{
		Workspace: workspace,
		Suspended: suspended,
	})
	if err != nil {
		action := "suspend"
		if !suspended {
			action = "resume"
		}
		slog.Error("failed to "+action+" workspace", "workspace", workspace, "error", err)
	}
}

// Conversion helpers

func subscriptionToResponse(sub *gostripe.Subscription, prices stripelib.PriceConfig) *cashier.ChangePlanResponse {
	plan := cashier.Plan_PLAN_HOBBY
	var currentPeriodEnd int64
	for _, item := range sub.Items.Data {
		if item.Price.ID == prices.ProPriceID {
			plan = cashier.Plan_PLAN_PRO
		}
		// All items share the same period — grab from first one
		if currentPeriodEnd == 0 && item.CurrentPeriodEnd > 0 {
			currentPeriodEnd = item.CurrentPeriodEnd
		}
	}

	creditCents := stripelib.PlanCreditCents(prices.HobbyPriceID, prices)
	if plan == cashier.Plan_PLAN_PRO {
		creditCents = stripelib.PlanCreditCents(prices.ProPriceID, prices)
	}

	return &cashier.ChangePlanResponse{
		Plan:              plan,
		Status:            mapSubscriptionStatus(sub.Status),
		CurrentPeriodEnd:  currentPeriodEnd,
		CreditAmountCents: int32(creditCents),
		TrialEnd:          sub.TrialEnd,
	}
}

func mapSubscriptionStatus(s gostripe.SubscriptionStatus) cashier.SubscriptionStatus {
	switch s {
	case gostripe.SubscriptionStatusActive:
		return cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE
	case gostripe.SubscriptionStatusPastDue:
		return cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_PAST_DUE
	case gostripe.SubscriptionStatusCanceled:
		return cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED
	case gostripe.SubscriptionStatusIncomplete:
		return cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_INCOMPLETE
	case gostripe.SubscriptionStatusTrialing:
		return cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIALING
	default:
		return cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_UNSPECIFIED
	}
}

func planToString(p cashier.Plan) string {
	switch p {
	case cashier.Plan_PLAN_PRO:
		return "PRO"
	default:
		return "HOBBY"
	}
}
