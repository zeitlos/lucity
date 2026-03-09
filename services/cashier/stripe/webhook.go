package stripe

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	gostripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

// EventHandler processes Stripe webhook events that require side effects
// (e.g., updating workspace metadata via the deployer).
type EventHandler interface {
	HandleStripeEvent(event gostripe.Event)
}

// WebhookHandler is an HTTP handler for Stripe webhook events.
type WebhookHandler struct {
	secret  string
	client  *Client
	handler EventHandler
}

// NewWebhookHandler creates a webhook handler with signature verification.
func NewWebhookHandler(secret string, client *Client, handler EventHandler) *WebhookHandler {
	return &WebhookHandler{
		secret:  secret,
		client:  client,
		handler: handler,
	}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 65536))
	if err != nil {
		slog.Error("failed to read webhook body", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), h.secret)
	if err != nil {
		slog.Warn("webhook signature verification failed", "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	slog.Info("stripe webhook received", "type", event.Type, "id", event.ID)

	switch event.Type {
	case "invoice.created":
		h.handleInvoiceCreated(event)
	case "invoice.payment_succeeded":
		slog.Info("payment succeeded", "event_id", event.ID)
	case "invoice.payment_failed":
		h.handler.HandleStripeEvent(event)
	case "customer.subscription.deleted":
		h.handler.HandleStripeEvent(event)
	default:
		slog.Debug("unhandled stripe event", "type", event.Type)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"received": true}`))
}

// handleInvoiceCreated applies plan credits as a negative invoice item.
func (h *WebhookHandler) handleInvoiceCreated(event gostripe.Event) {
	var inv gostripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		slog.Error("failed to unmarshal invoice", "error", err)
		return
	}

	// Only apply credits to subscription invoices
	if inv.Parent == nil || inv.Parent.SubscriptionDetails == nil {
		return
	}

	// Find the plan item to determine credit amount
	var planPriceID string
	var resourceCharges int64
	for _, line := range inv.Lines.Data {
		if line.Pricing == nil || line.Pricing.PriceDetails == nil {
			continue
		}
		priceID := line.Pricing.PriceDetails.Price
		if priceID == h.client.Prices.HobbyPriceID || priceID == h.client.Prices.ProPriceID {
			planPriceID = priceID
		} else {
			// Resource line item
			resourceCharges += line.Amount
		}
	}

	if planPriceID == "" {
		slog.Warn("no plan item found on invoice", "invoice", inv.ID)
		return
	}

	creditAmount := PlanCreditCents(planPriceID, h.client.Prices)
	if resourceCharges <= 0 {
		slog.Debug("no resource charges to credit", "invoice", inv.ID)
		return
	}

	// Credit = min(plan_credit, resource_charges)
	if creditAmount > resourceCharges {
		creditAmount = resourceCharges
	}

	if err := h.client.AddInvoiceCredit(nil, inv.Customer.ID, creditAmount, "Plan credit"); err != nil {
		slog.Error("failed to add invoice credit", "invoice", inv.ID, "error", err)
		return
	}

	slog.Info("applied plan credit", "invoice", inv.ID, "credit_cents", creditAmount, "resource_charges_cents", resourceCharges)
}
