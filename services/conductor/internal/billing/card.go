package billing

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v87"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/usererror"
)

// ErrPrepaidCard is returned when a card checkout saved a prepaid card, which
// doesn't count as a card on file.
var ErrPrepaidCard = errors.New("prepaid card")

// Price is what a plan costs a month.
type Price struct {
	AmountCents int64
	Currency    string
}

// Plan returns the plan the customer is billed on, PlanNone when it has none.
func (c *Client) Plan(ctx context.Context, customerID string) (billing.Plan, error) {
	current, err := c.billing.Subscription(ctx, customerID)

	if err != nil || current == nil || billing.Ended(current) {
		return billing.PlanNone, err
	}

	return c.billing.Plan(current), nil
}

// UsableCard reports whether the customer has a card attached that isn't
// prepaid. It doesn't need to be the default payment method.
func (c *Client) UsableCard(ctx context.Context, customerID string) (bool, error) {
	cards, err := c.cards(ctx, customerID)

	if err != nil {
		return false, err
	}

	for _, card := range cards {
		if !prepaid(card) {
			return true, nil
		}
	}

	return false, nil
}

// PaymentMethodCheckout starts a checkout that saves a card for the customer
// without charging it and without making it the default.
func (c *Client) PaymentMethodCheckout(ctx context.Context, customerID, workspace string, checkout Checkout) (string, error) {
	if err := c.expireCheckouts(ctx, customerID); err != nil {
		return "", err
	}

	session, err := c.stripe.V1CheckoutSessions.Create(ctx, &stripe.CheckoutSessionCreateParams{
		Customer:                  stripe.String(customerID),
		Mode:                      stripe.String(string(stripe.CheckoutSessionModeSetup)),
		Currency:                  stripe.String(string(stripe.CurrencyEUR)),
		AllowedPaymentMethodTypes: []*string{stripe.String("card")},
		SuccessURL:                stripe.String(checkout.SuccessURL),
		CancelURL:                 stripe.String(checkout.CancelURL),
		SetupIntentData: &stripe.CheckoutSessionCreateSetupIntentDataParams{
			Metadata: map[string]string{"workspace": workspace},
		},
		CustomText: &stripe.CheckoutSessionCreateCustomTextParams{
			Submit: &stripe.CheckoutSessionCreateCustomTextSubmitParams{
				Message: stripe.String("Nothing is charged. You're only billed once you choose a plan."),
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("create card checkout: %w", err)
	}

	return session.URL, nil
}

// CompletePaymentMethodCheckout checks a finished card checkout. A prepaid
// card is detached again and reported with ErrPrepaidCard.
func (c *Client) CompletePaymentMethodCheckout(ctx context.Context, customerID, sessionID string) error {
	session, err := c.stripe.V1CheckoutSessions.Retrieve(ctx, sessionID, &stripe.CheckoutSessionRetrieveParams{
		Expand: []*string{stripe.String("setup_intent.payment_method")},
	})

	if err != nil {
		return fmt.Errorf("read checkout session: %w", err)
	}

	if session.Customer == nil || session.Customer.ID != customerID {
		return usererror.New("this checkout does not belong to the workspace")
	}

	if session.Status != stripe.CheckoutSessionStatusComplete || session.Mode != stripe.CheckoutSessionModeSetup {
		return usererror.New("the card checkout is not complete")
	}

	if session.SetupIntent == nil || session.SetupIntent.PaymentMethod == nil {
		return usererror.New("the checkout collected no card")
	}

	card := session.SetupIntent.PaymentMethod

	if !prepaid(card) {
		return nil
	}

	if _, err := c.stripe.V1PaymentMethods.Detach(ctx, card.ID, nil); err != nil {
		return fmt.Errorf("detach prepaid card: %w", err)
	}

	return ErrPrepaidCard
}

// PlanPrice returns the monthly price of a plan.
func (c *Client) PlanPrice(ctx context.Context, plan billing.Plan) (Price, error) {
	id, err := c.planPrice(plan)

	if err != nil {
		return Price{}, err
	}

	price, err := c.stripe.V1Prices.Retrieve(ctx, id, nil)

	if err != nil {
		return Price{}, fmt.Errorf("read price of plan %s: %w", plan, err)
	}

	return Price{AmountCents: price.UnitAmount, Currency: string(price.Currency)}, nil
}

func (c *Client) cards(ctx context.Context, customerID string) ([]*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String(string(stripe.PaymentMethodTypeCard)),
	}
	params.Limit = stripe.Int64(10)

	var cards []*stripe.PaymentMethod

	for card, err := range c.stripe.V1PaymentMethods.List(ctx, params).All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list cards of customer %s: %w", customerID, err)
		}

		cards = append(cards, card)

		if len(cards) == 10 {
			break
		}
	}

	return cards, nil
}

func (c *Client) promoteCard(ctx context.Context, customerID string) (bool, error) {
	cards, err := c.cards(ctx, customerID)

	if err != nil || len(cards) == 0 {
		return false, err
	}

	if _, err := c.stripe.V1Customers.Update(ctx, customerID, &stripe.CustomerUpdateParams{
		InvoiceSettings: &stripe.CustomerUpdateInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(cards[0].ID),
		},
	}); err != nil {
		return false, fmt.Errorf("set default payment method: %w", err)
	}

	return true, nil
}

func prepaid(card *stripe.PaymentMethod) bool {
	return card.Card != nil && card.Card.Funding == stripe.CardFundingPrepaid
}
