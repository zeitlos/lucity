package stripe

import (
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

	event, err := webhook.ConstructEventWithOptions(body, r.Header.Get("Stripe-Signature"), h.secret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
	)
	if err != nil {
		slog.Warn("webhook signature verification failed", "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	slog.Info("stripe webhook received", "type", event.Type, "id", event.ID)

	switch event.Type {
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
