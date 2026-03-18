package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

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

func (s *Server) CreateCustomer(ctx context.Context, req *cashier.CreateCustomerRequest) (*cashier.CreateCustomerResponse, error) {
	if req.Workspace == "" {
		return nil, fmt.Errorf("workspace required")
	}

	// Idempotent: check if a customer already exists for this workspace in Stripe.
	existing, err := s.stripe.CustomerByWorkspace(ctx, req.Workspace)
	if err != nil {
		slog.Warn("failed to search for existing customer", "workspace", req.Workspace, "error", err)
		// Fall through to create — search failure shouldn't block signup
	}
	if existing != "" {
		slog.Info("stripe customer already exists", "workspace", req.Workspace, "customer_id", existing)
		return &cashier.CreateCustomerResponse{CustomerId: existing}, nil
	}

	customerID, err := s.stripe.CreateCustomer(ctx, req.Workspace, req.Name, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	slog.Info("stripe customer created", "workspace", req.Workspace, "customer_id", customerID)
	return &cashier.CreateCustomerResponse{CustomerId: customerID}, nil
}

func (s *Server) CreateSubscription(ctx context.Context, req *cashier.CreateSubscriptionRequest) (*cashier.CreateSubscriptionResponse, error) {
	if req.Workspace == "" || req.CustomerId == "" {
		return nil, fmt.Errorf("workspace and customer_id required")
	}

	// Idempotent: check if an active subscription already exists for this customer.
	existingSubID, err := s.stripe.ActiveSubscriptionForCustomer(ctx, req.CustomerId, req.Workspace)
	if err != nil {
		slog.Warn("failed to check for existing subscription", "workspace", req.Workspace, "error", err)
		// Fall through to create — list failure shouldn't block signup
	}
	if existingSubID != "" {
		slog.Info("stripe subscription already exists", "workspace", req.Workspace, "subscription_id", existingSubID)
		return &cashier.CreateSubscriptionResponse{SubscriptionId: existingSubID}, nil
	}

	planPriceID := s.stripe.PlanPriceID(planToString(req.Plan))

	subID, err := s.stripe.CreateSubscription(ctx, req.CustomerId, req.Workspace, planPriceID, int(req.CreditDays))
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	slog.Info("stripe subscription created", "workspace", req.Workspace, "subscription_id", subID)
	return &cashier.CreateSubscriptionResponse{SubscriptionId: subID}, nil
}

func (s *Server) ChangePlan(ctx context.Context, req *cashier.ChangePlanRequest) (*cashier.ChangePlanResponse, error) {
	if req.CustomerId == "" || req.SubscriptionId == "" {
		return nil, fmt.Errorf("customer_id and subscription_id required")
	}

	newPriceID := s.stripe.PlanPriceID(planToString(req.Plan))

	// Determine old plan price ID (the other one)
	oldPriceID := s.stripe.Prices.HobbyPriceID
	if newPriceID == s.stripe.Prices.HobbyPriceID {
		oldPriceID = s.stripe.Prices.ProPriceID
	}

	sub, err := s.stripe.ChangePlan(ctx, req.SubscriptionId, newPriceID, oldPriceID)
	if err != nil {
		return nil, fmt.Errorf("failed to change plan: %w", err)
	}

	hasPM, _ := s.stripe.HasPaymentMethod(ctx, req.CustomerId)

	slog.Info("plan changed", "workspace", req.Workspace, "plan", planToString(req.Plan))
	resp := subscriptionToResponse(sub, s.stripe.Prices)
	resp.HasPaymentMethod = hasPM
	return resp, nil
}

func (s *Server) Subscription(ctx context.Context, req *cashier.SubscriptionRequest) (*cashier.SubscriptionResponse, error) {
	if req.CustomerId == "" || req.SubscriptionId == "" {
		return nil, fmt.Errorf("customer_id and subscription_id required")
	}

	sub, err := s.stripe.Subscription(ctx, req.SubscriptionId)
	if err != nil {
		return nil, err
	}

	hasPM, _ := s.stripe.HasPaymentMethod(ctx, req.CustomerId)

	resp := subscriptionToResponse(sub, s.stripe.Prices)

	// If the customer has no payment method, check for an active credit grant.
	// The dashboard uses creditExpiry to show a countdown badge prompting
	// the user to add a payment method before credits expire.
	var creditExpiry int64
	if !hasPM {
		grantExpiry, _ := s.stripe.CreditGrantExpiry(ctx, req.CustomerId)
		if grantExpiry > time.Now().Unix() {
			creditExpiry = grantExpiry
		}
	}

	return &cashier.SubscriptionResponse{
		Plan:              resp.Plan,
		Status:            resp.Status,
		CurrentPeriodEnd:  resp.CurrentPeriodEnd,
		CreditAmountCents: resp.CreditAmountCents,
		CreditExpiry:      creditExpiry,
		HasPaymentMethod:  hasPM,
	}, nil
}

func (s *Server) BillingPortalURL(ctx context.Context, req *cashier.BillingPortalURLRequest) (*cashier.BillingPortalURLResponse, error) {
	if req.CustomerId == "" {
		return nil, fmt.Errorf("customer_id required")
	}

	url, err := s.stripe.BillingPortalURL(ctx, req.CustomerId, req.ReturnUrl)
	if err != nil {
		return nil, err
	}

	return &cashier.BillingPortalURLResponse{Url: url}, nil
}

func (s *Server) UsageSummary(ctx context.Context, req *cashier.UsageSummaryRequest) (*cashier.UsageSummaryResponse, error) {
	if req.CustomerId == "" || req.SubscriptionId == "" {
		return nil, fmt.Errorf("customer_id and subscription_id required")
	}

	inv, err := s.stripe.UpcomingInvoice(ctx, req.CustomerId, req.SubscriptionId)
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

	// Credits: use the credit grant total and cap applied credits at resource cost.
	// Stripe doesn't deduct from credit balance until invoice finalization, so we
	// calculate applied credits ourselves from the upcoming invoice's resource cost.
	creditBalance, err := s.stripe.CreditBalanceCents(ctx, req.CustomerId)
	if err != nil {
		slog.Warn("failed to get credit balance", "workspace", req.Workspace, "error", err)
	}
	creditsApplied := min(creditBalance, resourceCost)

	// Estimated total = plan + resources - credits applied
	estimated := planCost + resourceCost - creditsApplied

	return &cashier.UsageSummaryResponse{
		ResourceCostCents:   int32(resourceCost),
		CreditsCents:        int32(creditsApplied),
		EstimatedTotalCents: int32(estimated),
	}, nil
}

// HandleStripeEvent processes Stripe webhook events for billing lifecycle.
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

	// Unconditionally resume. If the workspace is not suspended, the deployer
	// will find no pre-suspend annotations and the operation is a no-op.
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
	ctx := auth.WithClaims(context.Background(), &auth.Claims{
		Subject: "cashier",
		Roles:   []auth.Role{auth.RoleUser},
	})
	ctx = auth.OutgoingContext(ctx)
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

func (s *Server) CreateCheckoutSession(ctx context.Context, req *cashier.CreateCheckoutSessionRequest) (*cashier.CreateCheckoutSessionResponse, error) {
	if req.Workspace == "" || req.Name == "" || req.Email == "" {
		return nil, fmt.Errorf("workspace, name, and email required")
	}

	planPriceID := s.stripe.PlanPriceID(planToString(req.Plan))

	url, sessionID, err := s.stripe.CreateCheckoutSession(ctx, req.Workspace, req.Name, planPriceID, req.Email, req.SuccessUrl, req.CancelUrl, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	slog.Info("checkout session created", "workspace", req.Workspace, "session_id", sessionID)
	return &cashier.CreateCheckoutSessionResponse{Url: url, SessionId: sessionID}, nil
}

func (s *Server) RetrieveCheckoutSession(ctx context.Context, req *cashier.RetrieveCheckoutSessionRequest) (*cashier.RetrieveCheckoutSessionResponse, error) {
	if req.SessionId == "" {
		return nil, fmt.Errorf("session_id required")
	}

	result, err := s.stripe.RetrieveCheckoutSession(ctx, req.SessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve checkout session: %w", err)
	}

	plan := cashier.Plan_PLAN_HOBBY
	// Plan is stored in subscription metadata, but we can also infer from the session metadata
	// The gateway will pass the correct plan based on what was requested

	return &cashier.RetrieveCheckoutSessionResponse{
		SessionId:      result.SessionID,
		Status:         result.Status,
		Workspace:      result.Workspace,
		Name:           result.Name,
		Plan:           plan,
		UserId:         result.UserID,
		Email:          result.Email,
		CustomerId:     result.CustomerID,
		SubscriptionId: result.SubscriptionID,
	}, nil
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
