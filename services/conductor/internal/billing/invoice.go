package billing

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v87"

	"github.com/zeitlos/lucity/pkg/billing"
)

const pastPeriods = 2

// Meter is a metered resource a subscription bills.
type Meter string

const (
	MeterCPU     Meter = "cpu"
	MeterMemory  Meter = "memory"
	MeterDisk    Meter = "disk"
	MeterStorage Meter = "storage"
	MeterEgress  Meter = "egress"
)

// Meters lists every meter in the order an invoice shows them.
var Meters = []Meter{MeterCPU, MeterMemory, MeterDisk, MeterStorage, MeterEgress}

// Period is one billing period of a subscription. End is exclusive.
type Period struct {
	Start   time.Time
	End     time.Time
	Current bool
}

// Rates are the cents one unit of a meter costs: a vCPU-minute, a GB-minute
// of memory, a GB-hour of disk or storage, a GB of egress.
type Rates map[Meter]float64

type InvoiceStatus string

const (
	InvoiceStatusPreview InvoiceStatus = "PREVIEW"
	InvoiceStatusOpen    InvoiceStatus = "OPEN"
	InvoiceStatusPaid    InvoiceStatus = "PAID"
)

// InvoiceLine is one line of an invoice: a plan when Plan is set, a meter
// otherwise. Quantity and UnitAmountCents are in the meter's Stripe unit.
type InvoiceLine struct {
	Plan            billing.Plan
	Meter           Meter
	Quantity        float64
	UnitAmountCents float64
	AmountCents     int64
}

type Party struct {
	Name    string
	Address []string
}

type Invoice struct {
	Number        string
	Status        InvoiceStatus
	IssuedAt      time.Time
	PaidAt        *time.Time
	PDFURL        string
	Issuer        Party
	BilledTo      Party
	Lines         []InvoiceLine
	SubtotalCents int64
	CreditCents   int64
	TotalCents    int64
}

// Periods returns the customer's current billing period and up to two
// before it, newest first, or nil when it never had a subscription.
func (c *Client) Periods(ctx context.Context, customerID string) ([]Period, error) {
	current, err := c.billing.Subscription(ctx, customerID)

	if err != nil || current == nil {
		return nil, err
	}

	var periods []Period

	if start, end := itemPeriod(current); !start.IsZero() {
		now := time.Now()
		periods = append(periods, Period{Start: start, End: end, Current: !now.Before(start) && now.Before(end)})
	}

	for invoice, err := range c.stripe.V1Invoices.List(ctx, &stripe.InvoiceListParams{Subscription: stripe.String(current.ID)}).All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list invoices: %w", err)
		}

		if invoice.BillingReason != stripe.InvoiceBillingReasonSubscriptionCycle || invoice.Status == stripe.InvoiceStatusVoid {
			continue
		}

		periods = append(periods, Period{Start: time.Unix(invoice.PeriodStart, 0), End: time.Unix(invoice.PeriodEnd, 0)})

		if len(periods) > pastPeriods {
			break
		}
	}

	return periods, nil
}

// Rates returns what one unit of each meter costs on the customer's
// subscription, or nil when it never had one.
func (c *Client) Rates(ctx context.Context, customerID string) (Rates, error) {
	current, err := c.billing.Subscription(ctx, customerID)

	if err != nil || current == nil {
		return nil, err
	}

	rates := Rates{}

	if current.Items == nil {
		return rates, nil
	}

	for _, item := range current.Items.Data {
		if item.Price == nil {
			continue
		}

		if meter, ok := c.meter(item.Price.ID); ok {
			rates[meter] = item.Price.UnitAmountDecimal.InexactFloat64()
		}
	}

	return rates, nil
}

// Invoice returns the invoice that closes a billing period: a preview of the
// next one for the current period, the finalized one for a past period. It
// returns nil when the period has no invoice.
func (c *Client) Invoice(ctx context.Context, customerID string, period Period) (*Invoice, error) {
	current, err := c.billing.Subscription(ctx, customerID)

	if err != nil || current == nil {
		return nil, err
	}

	var invoice *stripe.Invoice

	if period.Current {
		invoice, err = c.stripe.V1Invoices.CreatePreview(ctx, &stripe.InvoiceCreatePreviewParams{
			Customer:     stripe.String(customerID),
			Subscription: stripe.String(current.ID),
		})

		if err != nil {
			slog.DebugContext(ctx, "billing: no invoice preview", "customer", customerID, "error", err)
			return nil, nil
		}
	} else {
		invoice, err = c.periodInvoice(ctx, current.ID, period)

		if err != nil || invoice == nil {
			return nil, err
		}
	}

	lines, err := c.invoiceLines(ctx, invoice)

	if err != nil {
		return nil, err
	}

	view := &Invoice{
		Number:        invoice.Number,
		Status:        invoiceStatus(invoice, period),
		IssuedAt:      period.End,
		Lines:         lines,
		SubtotalCents: invoice.Subtotal,
		TotalCents:    invoice.Total,
		BilledTo:      Party{Name: invoice.CustomerName, Address: addressLines(invoice.CustomerAddress)},
		PDFURL:        invoice.InvoicePDF,
	}

	if transitions := invoice.StatusTransitions; transitions != nil {
		if transitions.FinalizedAt > 0 {
			view.IssuedAt = time.Unix(transitions.FinalizedAt, 0)
		}

		if transitions.PaidAt > 0 {
			paidAt := time.Unix(transitions.PaidAt, 0)
			view.PaidAt = &paidAt
		}
	}

	for _, credit := range invoice.TotalPretaxCreditAmounts {
		view.CreditCents += credit.Amount
	}

	if period.Current && view.CreditCents == 0 {
		balance, err := c.creditBalance(ctx, customerID)

		if err != nil {
			slog.WarnContext(ctx, "billing: failed to read credit balance", "customer", customerID, "error", err)
		}

		view.CreditCents = min(balance, meteredCents(view.Lines))
		view.TotalCents -= view.CreditCents
	}

	issuer, err := c.issuer(ctx)

	if err != nil {
		slog.WarnContext(ctx, "billing: failed to read the issuer", "error", err)
	}

	view.Issuer = issuer

	return view, nil
}

func (c *Client) periodInvoice(ctx context.Context, subscriptionID string, period Period) (*stripe.Invoice, error) {
	for invoice, err := range c.stripe.V1Invoices.List(ctx, &stripe.InvoiceListParams{Subscription: stripe.String(subscriptionID)}).All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list invoices: %w", err)
		}

		if invoice.BillingReason != stripe.InvoiceBillingReasonSubscriptionCycle || invoice.Status == stripe.InvoiceStatusVoid {
			continue
		}

		if invoice.PeriodStart == period.Start.Unix() && invoice.PeriodEnd == period.End.Unix() {
			return invoice, nil
		}

		if invoice.PeriodEnd < period.Start.Unix() {
			return nil, nil
		}
	}

	return nil, nil
}

func (c *Client) allLines(ctx context.Context, invoice *stripe.Invoice) ([]*stripe.InvoiceLineItem, error) {
	if invoice.Lines == nil {
		return nil, nil
	}

	if !invoice.Lines.HasMore {
		return invoice.Lines.Data, nil
	}

	var lines []*stripe.InvoiceLineItem

	for line, err := range c.stripe.V1Invoices.ListLines(ctx, &stripe.InvoiceListLinesParams{Invoice: stripe.String(invoice.ID)}).All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list invoice lines: %w", err)
		}

		lines = append(lines, line)
	}

	return lines, nil
}

func (c *Client) invoiceLines(ctx context.Context, invoice *stripe.Invoice) ([]InvoiceLine, error) {
	items, err := c.allLines(ctx, invoice)

	if err != nil {
		return nil, err
	}

	plans := map[billing.Plan]*InvoiceLine{}
	meters := map[Meter]*InvoiceLine{}
	var planOrder []billing.Plan

	for _, line := range items {
		if line.Pricing == nil || line.Pricing.PriceDetails == nil || line.Pricing.PriceDetails.Price == nil {
			continue
		}

		priceID := line.Pricing.PriceDetails.Price.ID
		quantity := line.QuantityDecimal.InexactFloat64()
		unit := line.Pricing.UnitAmountDecimal.InexactFloat64()

		if plan := c.plan(priceID); plan != billing.PlanNone {
			existing, ok := plans[plan]

			if !ok {
				existing = &InvoiceLine{Plan: plan, Quantity: 1}
				plans[plan] = existing
				planOrder = append(planOrder, plan)
			}

			existing.AmountCents += line.Amount
			existing.UnitAmountCents = float64(existing.AmountCents)
			continue
		}

		meter, ok := c.meter(priceID)

		if !ok {
			continue
		}

		existing, ok := meters[meter]

		if !ok {
			existing = &InvoiceLine{Meter: meter, UnitAmountCents: unit}
			meters[meter] = existing
		}

		existing.Quantity += quantity
		existing.AmountCents += line.Amount
	}

	lines := make([]InvoiceLine, 0, len(plans)+len(meters))

	for _, plan := range planOrder {
		lines = append(lines, *plans[plan])
	}

	for _, meter := range Meters {
		if line, ok := meters[meter]; ok {
			lines = append(lines, *line)
		}
	}

	return lines, nil
}

func (c *Client) issuer(ctx context.Context) (Party, error) {
	account, err := c.stripe.V1Accounts.Retrieve(ctx, nil)

	if err != nil {
		return Party{}, fmt.Errorf("retrieve account: %w", err)
	}

	var party Party

	if account.Settings != nil && account.Settings.Dashboard != nil {
		party.Name = account.Settings.Dashboard.DisplayName
	}

	if account.BusinessProfile != nil {
		if account.BusinessProfile.Name != "" {
			party.Name = account.BusinessProfile.Name
		}

		party.Address = addressLines(account.BusinessProfile.SupportAddress)
	}

	if account.Company != nil {
		if account.Company.Name != "" {
			party.Name = account.Company.Name
		}

		if lines := addressLines(account.Company.Address); len(lines) > 0 {
			party.Address = lines
		}
	}

	return party, nil
}

func (c *Client) meter(priceID string) (Meter, bool) {
	switch priceID {
	case c.prices.CPU:
		return MeterCPU, true
	case c.prices.Memory:
		return MeterMemory, true
	case c.prices.Disk:
		return MeterDisk, true
	case c.prices.Storage:
		return MeterStorage, true
	case c.prices.Egress:
		return MeterEgress, true
	}

	return "", false
}

func (c *Client) plan(priceID string) billing.Plan {
	switch priceID {
	case c.prices.Hobby:
		return billing.PlanHobby
	case c.prices.Pro:
		return billing.PlanPro
	}

	return billing.PlanNone
}

func invoiceStatus(invoice *stripe.Invoice, period Period) InvoiceStatus {
	if period.Current {
		return InvoiceStatusPreview
	}

	if invoice.Status == stripe.InvoiceStatusPaid {
		return InvoiceStatusPaid
	}

	return InvoiceStatusOpen
}

func meteredCents(lines []InvoiceLine) int64 {
	var total int64

	for _, line := range lines {
		if line.Plan == billing.PlanNone {
			total += line.AmountCents
		}
	}

	return total
}

func itemPeriod(subscription *stripe.Subscription) (time.Time, time.Time) {
	if subscription.Items != nil {
		for _, item := range subscription.Items.Data {
			if item.CurrentPeriodStart > 0 && item.CurrentPeriodEnd > 0 {
				return time.Unix(item.CurrentPeriodStart, 0), time.Unix(item.CurrentPeriodEnd, 0)
			}
		}
	}

	return time.Time{}, time.Time{}
}

func addressLines(address *stripe.Address) []string {
	if address == nil {
		return nil
	}

	var lines []string

	for _, line := range []string{address.Line1, address.Line2, strings.TrimSpace(address.PostalCode + " " + address.City), address.State, address.Country} {
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}
