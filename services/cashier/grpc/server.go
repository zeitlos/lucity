package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gostripe "github.com/stripe/stripe-go/v87"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/cashier"
	"github.com/zeitlos/lucity/pkg/logto"
	stripelib "github.com/zeitlos/lucity/services/cashier/stripe"
)

type Server struct {
	cashier.UnimplementedCashierServiceServer
	stripe  *stripelib.Client
	billing *billing.Client
	logto   *logto.Client
}

func NewServer(stripeClient *stripelib.Client, billingClient *billing.Client, logtoClient *logto.Client) *Server {
	return &Server{
		stripe:  stripeClient,
		billing: billingClient,
		logto:   logtoClient,
	}
}

func (s *Server) CreateCustomer(ctx context.Context, req *cashier.CreateCustomerRequest) (*cashier.CreateCustomerResponse, error) {
	if req.Workspace == "" {
		return nil, fmt.Errorf("workspace required")
	}

	// Idempotent: check if Logto org already has a stripeCustomerId.
	org, err := s.logto.OrganizationByName(ctx, req.Workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to look up workspace: %w", err)
	}
	if existingID, _ := org.CustomData["stripeCustomerId"].(string); existingID != "" {
		if err := s.stripe.EnsureCustomerWorkspace(ctx, existingID, req.Workspace); err != nil {
			slog.WarnContext(ctx, "failed to link stripe customer to workspace", "workspace", req.Workspace, "customer_id", existingID, "error", err)
		}
		slog.Debug("stripe customer already exists", "workspace", req.Workspace, "customer_id", existingID)
		return &cashier.CreateCustomerResponse{CustomerId: existingID}, nil
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

	existing, err := s.billing.Subscription(ctx, req.CustomerId)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing subscription: %w", err)
	}
	if existing != nil {
		if billing.Ended(existing) {
			return nil, status.Errorf(codes.FailedPrecondition, "the subscription of workspace %s has ended", req.Workspace)
		}
		slog.Debug("stripe subscription already exists", "workspace", req.Workspace, "subscription_id", existing.ID)
		return &cashier.CreateSubscriptionResponse{SubscriptionId: existing.ID}, nil
	}

	subID, err := s.stripe.CreateSubscription(ctx, req.CustomerId, req.Workspace, int(req.CreditDays))
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	slog.Info("stripe subscription created", "workspace", req.Workspace, "subscription_id", subID)
	return &cashier.CreateSubscriptionResponse{SubscriptionId: subID}, nil
}

func (s *Server) ChangePlan(ctx context.Context, req *cashier.ChangePlanRequest) (*cashier.ChangePlanResponse, error) {
	if req.CustomerId == "" {
		return nil, fmt.Errorf("customer_id required")
	}

	current, err := s.subscription(ctx, req.Workspace, req.CustomerId)
	if err != nil {
		return nil, err
	}

	newPriceID := s.stripe.PlanPriceID(planToString(req.Plan))

	// Determine old plan price ID (the other one).
	oldPriceID := s.stripe.Prices.HobbyPriceID
	if newPriceID == s.stripe.Prices.HobbyPriceID {
		oldPriceID = s.stripe.Prices.ProPriceID
	}

	sub, err := s.stripe.ChangePlan(ctx, current.ID, newPriceID, oldPriceID)
	if err != nil {
		return nil, fmt.Errorf("failed to change plan: %w", err)
	}

	hasPM, _ := s.stripe.HasPaymentMethod(ctx, sub)

	slog.Info("plan changed", "workspace", req.Workspace, "plan", planToString(req.Plan))
	resp := s.subscriptionToResponse(sub, s.stripe.Prices)
	resp.HasPaymentMethod = hasPM
	return resp, nil
}

func (s *Server) Subscription(ctx context.Context, req *cashier.SubscriptionRequest) (*cashier.SubscriptionResponse, error) {
	if req.CustomerId == "" {
		return nil, fmt.Errorf("customer_id required")
	}

	sub, err := s.subscription(ctx, req.Workspace, req.CustomerId)
	if err != nil {
		return nil, err
	}

	hasPM, _ := s.stripe.HasPaymentMethod(ctx, sub)

	resp := s.subscriptionToResponse(sub, s.stripe.Prices)

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
	if req.CustomerId == "" {
		return nil, fmt.Errorf("customer_id required")
	}

	sub, err := s.subscription(ctx, req.Workspace, req.CustomerId)
	if err != nil {
		return nil, err
	}

	var resourceCost int64
	var planCost int64

	inv, err := s.stripe.UpcomingInvoice(ctx, req.CustomerId, sub.ID)
	if err == nil {
		for _, line := range inv.Lines.Data {
			if line.Pricing == nil || line.Pricing.PriceDetails == nil || line.Pricing.PriceDetails.Price == nil {
				continue
			}
			priceID := line.Pricing.PriceDetails.Price.ID
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
	openTotal, err := s.stripe.UnpaidInvoiceTotal(ctx, sub.ID)
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

func (s *Server) subscription(ctx context.Context, workspace, customerID string) (*gostripe.Subscription, error) {
	sub, err := s.billing.Subscription(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, status.Errorf(codes.NotFound, "workspace %s has no subscription", workspace)
	}
	return sub, nil
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
	if req.CustomerId == "" || req.CheckoutSessionId == "" {
		return nil, fmt.Errorf("customer_id and checkout_session_id required")
	}

	current, err := s.subscription(ctx, req.Workspace, req.CustomerId)
	if err != nil {
		return nil, err
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
	sub, err := s.stripe.AddPlanToSubscription(ctx, current.ID, planPriceID)
	if err != nil {
		return nil, fmt.Errorf("failed to add plan: %w", err)
	}

	slog.Info("plan added to subscription", "workspace", req.Workspace, "plan", planPriceID)
	resp := s.subscriptionToResponse(sub, s.stripe.Prices)
	return &cashier.AddPlanResponse{
		Plan:              resp.Plan,
		Status:            resp.Status,
		CurrentPeriodEnd:  resp.CurrentPeriodEnd,
		CreditAmountCents: resp.CreditAmountCents,
		HasPaymentMethod:  true,
	}, nil
}

func (s *Server) subscriptionToResponse(sub *gostripe.Subscription, prices stripelib.PriceConfig) *cashier.ChangePlanResponse {
	plan := cashier.Plan_PLAN_UNSPECIFIED
	var currentPeriodEnd int64
	for _, item := range sub.Items.Data {
		switch item.Price.ID {
		case prices.HobbyPriceID:
			plan = cashier.Plan_PLAN_HOBBY
		case prices.ProPriceID:
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
		Status:            s.mapSubscriptionStatus(sub),
		CurrentPeriodEnd:  currentPeriodEnd,
		CreditAmountCents: int32(creditCents),
	}
}

func (s *Server) mapSubscriptionStatus(sub *gostripe.Subscription) cashier.SubscriptionStatus {
	switch sub.Status {
	case gostripe.SubscriptionStatusActive:
		if !s.subscriptionHasPlan(sub) {
			// Subscriptions wihtout a plan are trials
			return cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIALING
		}

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
