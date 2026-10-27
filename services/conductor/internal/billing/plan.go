package billing

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v87"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/usererror"
)

// Selection is the outcome of choosing a plan: the subscription now on it,
// or a checkout to complete first because the customer has no card on file.
type Selection struct {
	Subscription *Subscription
	CheckoutURL  string
}

// Checkout configures the Stripe Checkout a plan selection may start.
type Checkout struct {
	SuccessURL string
	CancelURL  string
	Metadata   map[string]string
}

// Session is a completed or pending Stripe Checkout.
type Session struct {
	Complete   bool
	CustomerID string
	Metadata   map[string]string
}

// SelectPlan puts the customer on a plan. It switches the plan of a paid
// subscription, adds the plan to a trial, or starts a fresh subscription when
// there is none or it ended. Without a card on file it returns a checkout to
// collect one instead.
func (c *Client) SelectPlan(ctx context.Context, customerID, workspace string, plan billing.Plan, checkout Checkout) (*Selection, error) {
	price, err := c.planPrice(plan)

	if err != nil {
		return nil, err
	}

	current, err := c.billing.Subscription(ctx, customerID)

	if err != nil {
		return nil, err
	}

	live := current != nil && !billing.Ended(current)

	if live && c.billing.Plan(current) == plan {
		return c.selection(ctx, customerID)
	}

	if live && c.billing.Plan(current).Paid() {
		if err := c.swapPlan(ctx, current, price); err != nil {
			return nil, err
		}

		return c.selection(ctx, customerID)
	}

	hasPaymentMethod, err := c.hasPaymentMethod(ctx, customerID, current)

	if err != nil {
		return nil, err
	}

	if !hasPaymentMethod {
		url, err := c.checkout(ctx, customerID, workspace, price, live, checkout)

		if err != nil {
			return nil, err
		}

		return &Selection{CheckoutURL: url}, nil
	}

	if live {
		err = c.addPlan(ctx, current.ID, price)
	} else {
		err = c.subscribe(ctx, customerID, workspace, price)
	}

	if err != nil {
		return nil, err
	}

	return c.selection(ctx, customerID)
}

// CompleteCheckout finishes a checkout that SelectPlan started for the
// customer. A checkout that collected a card makes it the invoice default and
// then puts the customer on the plan it was started for.
func (c *Client) CompleteCheckout(ctx context.Context, customerID, workspace, sessionID string) (*Subscription, error) {
	session, err := c.stripe.V1CheckoutSessions.Retrieve(ctx, sessionID, &stripe.CheckoutSessionRetrieveParams{
		Expand: []*string{stripe.String("setup_intent.payment_method")},
	})

	if err != nil {
		return nil, fmt.Errorf("read checkout session: %w", err)
	}

	if session.Customer == nil || session.Customer.ID != customerID {
		return nil, usererror.New("this checkout does not belong to the workspace")
	}

	if session.Status != stripe.CheckoutSessionStatusComplete {
		return nil, usererror.Errorf("the checkout is not complete (status: %s)", session.Status)
	}

	if session.Mode == stripe.CheckoutSessionModeSetup {
		if session.SetupIntent == nil || session.SetupIntent.PaymentMethod == nil {
			return nil, usererror.New("the checkout collected no payment method")
		}

		plan, err := c.planOf(session.Metadata["plan_price_id"])

		if err != nil {
			return nil, err
		}

		if _, err := c.stripe.V1Customers.Update(ctx, customerID, &stripe.CustomerUpdateParams{
			InvoiceSettings: &stripe.CustomerUpdateInvoiceSettingsParams{
				DefaultPaymentMethod: stripe.String(session.SetupIntent.PaymentMethod.ID),
			},
		}); err != nil {
			return nil, fmt.Errorf("set default payment method: %w", err)
		}

		if _, err := c.SelectPlan(ctx, customerID, workspace, plan, Checkout{}); err != nil {
			return nil, err
		}
	}

	return c.Subscription(ctx, customerID)
}

func (c *Client) Session(ctx context.Context, sessionID string) (*Session, error) {
	session, err := c.stripe.V1CheckoutSessions.Retrieve(ctx, sessionID, nil)

	if err != nil {
		return nil, fmt.Errorf("read checkout session: %w", err)
	}

	result := &Session{
		Complete: session.Status == stripe.CheckoutSessionStatusComplete,
		Metadata: session.Metadata,
	}

	if session.Customer != nil {
		result.CustomerID = session.Customer.ID
	}

	return result, nil
}

func (c *Client) selection(ctx context.Context, customerID string) (*Selection, error) {
	subscription, err := c.Subscription(ctx, customerID)

	if err != nil {
		return nil, err
	}

	return &Selection{Subscription: subscription}, nil
}

func (c *Client) swapPlan(ctx context.Context, subscription *stripe.Subscription, price string) error {
	for _, item := range subscription.Items.Data {
		if item.Price == nil || (item.Price.ID != c.prices.Hobby && item.Price.ID != c.prices.Pro) {
			continue
		}

		if _, err := c.stripe.V1SubscriptionItems.Update(ctx, item.ID, &stripe.SubscriptionItemUpdateParams{
			Price:             stripe.String(price),
			ProrationBehavior: stripe.String("create_prorations"),
		}); err != nil {
			return fmt.Errorf("change plan: %w", err)
		}

		return nil
	}

	return usererror.New("the subscription has no plan to change")
}

func (c *Client) addPlan(ctx context.Context, subscriptionID, price string) error {
	params := &stripe.SubscriptionUpdateParams{
		Items:             []*stripe.SubscriptionUpdateItemParams{{Price: stripe.String(price)}},
		ProrationBehavior: stripe.String("create_prorations"),
	}
	params.AddExtra("billing_cycle_anchor[type]", "now")
	params.AddExtra("billing_thresholds", "")
	params.AddExtra("pending_invoice_item_interval", "")

	if _, err := c.stripe.V1Subscriptions.Update(ctx, subscriptionID, params); err != nil {
		return fmt.Errorf("add plan: %w", err)
	}

	return nil
}

func (c *Client) subscribe(ctx context.Context, customerID, workspace, price string) error {
	if _, err := c.stripe.V1Subscriptions.Create(ctx, &stripe.SubscriptionCreateParams{
		Customer:        stripe.String(customerID),
		PaymentBehavior: stripe.String("error_if_incomplete"),
		BillingMode:     &stripe.SubscriptionCreateBillingModeParams{Type: stripe.String(string(stripe.SubscriptionBillingModeTypeClassic))},
		Items:           c.items(price),
		Metadata:        map[string]string{"workspace": workspace},
	}); err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}

	return nil
}

func (c *Client) checkout(ctx context.Context, customerID, workspace, price string, live bool, checkout Checkout) (string, error) {
	params := &stripe.CheckoutSessionCreateParams{
		Customer:   stripe.String(customerID),
		SuccessURL: stripe.String(checkout.SuccessURL),
		CancelURL:  stripe.String(checkout.CancelURL),
		Metadata:   checkout.Metadata,
	}

	if live {
		params.Mode = stripe.String(string(stripe.CheckoutSessionModeSetup))
		params.Currency = stripe.String(string(stripe.CurrencyEUR))
		params.CustomText = &stripe.CheckoutSessionCreateCustomTextParams{
			Submit: &stripe.CheckoutSessionCreateCustomTextSubmitParams{
				Message: stripe.String(fmt.Sprintf("You're signing up for the %s plan. Your card will be charged at the end of each billing cycle.", planName(price, c.prices))),
			},
		}

		if params.Metadata == nil {
			params.Metadata = map[string]string{}
		}

		params.Metadata["plan_price_id"] = price
	} else {
		params.Mode = stripe.String(string(stripe.CheckoutSessionModeSubscription))
		params.PaymentMethodCollection = stripe.String("always")
		params.LineItems = []*stripe.CheckoutSessionCreateLineItemParams{{Price: stripe.String(price), Quantity: stripe.Int64(1)}}

		for _, metered := range c.prices.Metered {
			params.LineItems = append(params.LineItems, &stripe.CheckoutSessionCreateLineItemParams{Price: stripe.String(metered)})
		}

		params.SubscriptionData = &stripe.CheckoutSessionCreateSubscriptionDataParams{
			BillingMode: &stripe.CheckoutSessionCreateSubscriptionDataBillingModeParams{Type: stripe.String(string(stripe.SubscriptionBillingModeTypeClassic))},
			Metadata:    map[string]string{"workspace": workspace},
		}
	}

	session, err := c.stripe.V1CheckoutSessions.Create(ctx, params)

	if err != nil {
		return "", fmt.Errorf("create checkout session: %w", err)
	}

	return session.URL, nil
}

func (c *Client) planPrice(plan billing.Plan) (string, error) {
	switch plan {
	case billing.PlanHobby:
		return c.prices.Hobby, nil
	case billing.PlanPro:
		return c.prices.Pro, nil
	}

	return "", usererror.Errorf("unknown plan %q", plan)
}

func (c *Client) planOf(price string) (billing.Plan, error) {
	switch price {
	case c.prices.Hobby:
		return billing.PlanHobby, nil
	case c.prices.Pro:
		return billing.PlanPro, nil
	}

	return billing.PlanNone, usererror.New("the checkout is not for a known plan")
}

func planName(price string, prices Prices) string {
	if price == prices.Pro {
		return "Pro"
	}

	return "Hobby"
}
