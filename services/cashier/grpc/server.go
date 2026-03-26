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
	"github.com/zeitlos/lucity/pkg/logto"
	stripelib "github.com/zeitlos/lucity/services/cashier/stripe"
)

type Server struct {
	cashier.UnimplementedCashierServiceServer
	stripe   *stripelib.Client
	deployer deployer.DeployerServiceClient
	logto    *logto.Client
	issuer   *auth.Issuer
}

func NewServer(stripeClient *stripelib.Client, deployerClient deployer.DeployerServiceClient, logtoClient *logto.Client, issuer *auth.Issuer) *Server {
	return &Server{
		stripe:   stripeClient,
		deployer: deployerClient,
		logto:    logtoClient,
		issuer:   issuer,
	}
}

func (s *Server) CreateCustomer(ctx context.Context, req *cashier.CreateCustomerRequest) (*cashier.CreateCustomerResponse, error) {
	if req.Workspace == "" {
		return nil, fmt.Errorf("workspace required")
	}

	// Idempotent: check if Logto org already has a stripeCustomerId.
	org, err := s.logto.OrganizationByName(ctx, req.Workspace)
	if err != nil {
		slog.Warn("failed to look up logto org for customer check", "workspace", req.Workspace, "error", err)
	}
	if org != nil && org.CustomData != nil {
		if existingID, ok := org.CustomData["stripeCustomerId"].(string); ok && existingID != "" {
			slog.Info("stripe customer already exists", "workspace", req.Workspace, "customer_id", existingID)
			return &cashier.CreateCustomerResponse{CustomerId: existingID}, nil
		}
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
	}
	if existingSubID != "" {
		slog.Info("stripe subscription already exists", "workspace", req.Workspace, "subscription_id", existingSubID)
		return &cashier.CreateSubscriptionResponse{SubscriptionId: existingSubID}, nil
	}

	subID, err := s.stripe.CreateSubscription(ctx, req.CustomerId, req.Workspace, int(req.CreditDays))
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

	// Determine old plan price ID (the other one).
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

	// Use the actual credit balance from Stripe instead of the plan entitlement.
	// This correctly reflects trial credits for users without a plan, and is always
	// accurate for users with a plan (plan credits are issued as credit grants too).
	creditBalance, err := s.stripe.CreditBalanceCents(ctx, req.CustomerId)
	if err != nil {
		slog.Warn("failed to get credit balance", "customer", req.CustomerId, "error", err)
	}

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
		CreditAmountCents: int32(creditBalance),
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

	var resourceCost int64
	var planCost int64

	inv, err := s.stripe.UpcomingInvoice(ctx, req.CustomerId, req.SubscriptionId)
	if err == nil {
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
	}

	// Include open invoices (e.g. threshold-triggered invoices that are pending
	// payment). Their usage is no longer on the upcoming preview but the credits
	// haven't been deducted yet until the invoice is paid.
	openTotal, err := s.stripe.UnpaidInvoiceTotal(ctx, req.SubscriptionId)
	if err != nil {
		slog.Warn("failed to get open invoice total", "workspace", req.Workspace, "error", err)
	}
	resourceCost += openTotal

	// Credits: use the credit grant total and cap applied credits at resource cost.
	// Stripe doesn't deduct from credit balance until invoice finalization, so we
	// calculate applied credits ourselves from the total outstanding resource cost.
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
// Returns an error if the event could not be processed (triggers Stripe retry).
func (s *Server) HandleStripeEvent(event gostripe.Event) error {
	switch event.Type {
	case "invoice.payment_failed":
		return s.handlePaymentFailed(event)
	case "invoice.payment_succeeded":
		return s.handlePaymentSucceeded(event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(event)
	}
	return nil
}

func (s *Server) handlePaymentFailed(event gostripe.Event) error {
	var inv gostripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	workspace := ""
	if inv.Parent != nil && inv.Parent.SubscriptionDetails != nil {
		workspace = inv.Parent.SubscriptionDetails.Metadata["workspace"]
	}
	if workspace == "" {
		return fmt.Errorf("no workspace in subscription metadata for invoice %s", inv.ID)
	}

	slog.Error("payment failed, suspending workspace", "workspace", workspace, "invoice", inv.ID)
	s.suspendWorkspace(workspace, true)
	return nil
}

func (s *Server) handlePaymentSucceeded(event gostripe.Event) error {
	var inv gostripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	workspace := ""
	var subscriptionID, customerID string
	if inv.Parent != nil && inv.Parent.SubscriptionDetails != nil {
		workspace = inv.Parent.SubscriptionDetails.Metadata["workspace"]
		if inv.Parent.SubscriptionDetails.Subscription != nil {
			subscriptionID = inv.Parent.SubscriptionDetails.Subscription.ID
		}
	}
	if inv.Customer != nil {
		customerID = inv.Customer.ID
	}
	if workspace == "" {
		// Payment succeeded but no workspace — not actionable, don't retry.
		return nil
	}

	// For trial subscriptions (no plan), check if the user has converted.
	// If they haven't added a payment method, suspend: the trial is over.
	if subscriptionID != "" {
		ctx := context.Background()
		sub, err := s.stripe.Subscription(ctx, subscriptionID)
		if err != nil {
			slog.Error("failed to fetch subscription for trial check", "error", err, "subscription", subscriptionID)
			// Fall through to resume (safe default for paying customers)
		} else if !s.subscriptionHasPlan(sub) {
			// Trial subscription: check for payment method
			hasPM, _ := s.stripe.HasPaymentMethod(ctx, customerID)
			if !hasPM {
				slog.Info("trial ended, no plan or payment method, suspending workspace",
					"workspace", workspace, "invoice", inv.ID)
				s.suspendWorkspace(workspace, true)
				return nil
			}
		}
	}

	slog.Info("payment succeeded, resuming workspace", "workspace", workspace, "invoice", inv.ID)
	s.suspendWorkspace(workspace, false)
	return nil
}

// subscriptionHasPlan returns true if the subscription has a Hobby or Pro plan item.
func (s *Server) subscriptionHasPlan(sub *gostripe.Subscription) bool {
	for _, item := range sub.Items.Data {
		if item.Price.ID == s.stripe.Prices.HobbyPriceID ||
			item.Price.ID == s.stripe.Prices.ProPriceID {
			return true
		}
	}
	return false
}

func (s *Server) handleSubscriptionDeleted(event gostripe.Event) error {
	var sub gostripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	workspace := sub.Metadata["workspace"]
	if workspace == "" {
		return fmt.Errorf("no workspace in subscription metadata for %s", sub.ID)
	}

	slog.Warn("subscription deleted, suspending workspace", "workspace", workspace, "subscription", sub.ID)
	s.suspendWorkspace(workspace, true)
	return nil
}

func (s *Server) suspendWorkspace(workspace string, suspended bool) {
	ctx := auth.WithClaims(context.Background(), &auth.Claims{
		Subject: "cashier",
		Roles:   []auth.Role{auth.RoleUser},
	})
	ctx = auth.WithIssuer(ctx, s.issuer)
	ctx = auth.OutgoingContext(ctx)

	action := "suspend"
	if !suspended {
		action = "resume"
	}

	// 1. Write suspended flag to GitOps repo via deployer -> packager.
	_, err := s.deployer.SuspendWorkspace(ctx, &deployer.SuspendWorkspaceRequest{
		Workspace: workspace,
		Suspended: suspended,
	})
	if err != nil {
		slog.Error("failed to "+action+" workspace", "workspace", workspace, "error", err)
	}

	// 2. Persist suspension state in Logto org customData for dashboard visibility.
	org, err := s.logto.OrganizationByName(context.Background(), workspace)
	if err != nil {
		slog.Error("failed to look up logto org for suspension", "workspace", workspace, "error", err)
		return
	}
	customData := org.CustomData
	if customData == nil {
		customData = make(map[string]interface{})
	}
	if suspended {
		customData["suspended"] = true
	} else {
		delete(customData, "suspended")
	}
	if err := s.logto.UpdateOrganizationCustomData(context.Background(), org.ID, customData); err != nil {
		slog.Error("failed to update logto suspension state", "workspace", workspace, "suspended", suspended, "error", err)
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

func (s *Server) CreatePlanCheckoutSession(ctx context.Context, req *cashier.CreatePlanCheckoutSessionRequest) (*cashier.CreatePlanCheckoutSessionResponse, error) {
	if req.CustomerId == "" || req.SuccessUrl == "" || req.CancelUrl == "" {
		return nil, fmt.Errorf("customer_id, success_url, and cancel_url required")
	}

	plan := planToString(req.Plan)
	planPriceID := s.stripe.PlanPriceID(plan)
	planName := "Hobby"
	if plan == "PRO" {
		planName = "Pro"
	}

	url, sessionID, err := s.stripe.CreateSetupCheckoutSession(ctx, req.CustomerId, planPriceID, planName, req.SuccessUrl, req.CancelUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan checkout session: %w", err)
	}

	slog.Info("plan checkout session created", "workspace", req.Workspace, "session_id", sessionID)
	return &cashier.CreatePlanCheckoutSessionResponse{Url: url, SessionId: sessionID}, nil
}

func (s *Server) AddPlan(ctx context.Context, req *cashier.AddPlanRequest) (*cashier.AddPlanResponse, error) {
	if req.CustomerId == "" || req.SubscriptionId == "" || req.CheckoutSessionId == "" {
		return nil, fmt.Errorf("customer_id, subscription_id, and checkout_session_id required")
	}

	// Retrieve the setup checkout session to get the payment method.
	paymentMethodID, planPriceID, err := s.stripe.RetrieveSetupCheckoutSession(ctx, req.CheckoutSessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve checkout session: %w", err)
	}

	// Override plan from session metadata if the request specifies one.
	if req.Plan != cashier.Plan_PLAN_UNSPECIFIED {
		planPriceID = s.stripe.PlanPriceID(planToString(req.Plan))
	}

	// Set the payment method as the customer's default for invoices.
	if err := s.stripe.SetDefaultPaymentMethod(ctx, req.CustomerId, paymentMethodID); err != nil {
		return nil, fmt.Errorf("failed to set payment method: %w", err)
	}

	// Add the plan to the subscription.
	sub, err := s.stripe.AddPlanToSubscription(ctx, req.SubscriptionId, planPriceID)
	if err != nil {
		return nil, fmt.Errorf("failed to add plan: %w", err)
	}

	// Clear trial billing params (thresholds, interval) and reset billing cycle
	// anchor so the paid billing period starts fresh from now.
	if err := s.stripe.ClearTrialBillingParams(ctx, req.SubscriptionId); err != nil {
		slog.Warn("failed to clear trial billing params", "workspace", req.Workspace, "error", err)
	}

	// Resume workspace in case it was suspended at trial end.
	s.suspendWorkspace(req.Workspace, false)

	slog.Info("plan added to subscription", "workspace", req.Workspace, "plan", planPriceID)
	resp := subscriptionToResponse(sub, s.stripe.Prices)
	return &cashier.AddPlanResponse{
		Plan:              resp.Plan,
		Status:            resp.Status,
		CurrentPeriodEnd:  resp.CurrentPeriodEnd,
		CreditAmountCents: resp.CreditAmountCents,
		HasPaymentMethod:  true,
	}, nil
}

// Conversion helpers

func subscriptionToResponse(sub *gostripe.Subscription, prices stripelib.PriceConfig) *cashier.ChangePlanResponse {
	plan := cashier.Plan_PLAN_UNSPECIFIED
	var currentPeriodEnd int64
	for _, item := range sub.Items.Data {
		if item.Price.ID == prices.HobbyPriceID {
			plan = cashier.Plan_PLAN_HOBBY
		} else if item.Price.ID == prices.ProPriceID {
			plan = cashier.Plan_PLAN_PRO
		}
		// All items share the same period — grab from first one
		if currentPeriodEnd == 0 && item.CurrentPeriodEnd > 0 {
			currentPeriodEnd = item.CurrentPeriodEnd
		}
	}

	var creditCents int64
	if plan != cashier.Plan_PLAN_UNSPECIFIED {
		creditCents = stripelib.PlanCreditCents(prices.HobbyPriceID, prices)
		if plan == cashier.Plan_PLAN_PRO {
			creditCents = stripelib.PlanCreditCents(prices.ProPriceID, prices)
		}
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
