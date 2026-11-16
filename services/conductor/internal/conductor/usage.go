package conductor

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	pkgbilling "github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/billing"
	"github.com/zeitlos/lucity/services/conductor/internal/usage"
)

type UsageReport struct {
	Period        billing.Period
	MeasuredUntil time.Time
	Projects      []ProjectUsage
	Invoice       *InvoiceView
}

type ProjectUsage struct {
	ID           platform.ProjectID
	Name         string
	DeletedAt    *time.Time
	Cents        float64
	Daily        []float64
	Environments []EnvironmentUsage
}

type EnvironmentUsage struct {
	ID        platform.EnvironmentID
	Name      string
	DeletedAt *time.Time
	Cents     float64
	Daily     []float64
	Resources []ResourceUsage
}

type ResourceUsage struct {
	ID    string
	Name  string
	Kind  usage.Kind
	Cents float64
	Daily []float64
}

type InvoiceView struct {
	Number        string
	Status        billing.InvoiceStatus
	IssuedAt      time.Time
	PaidAt        *time.Time
	PDFURL        string
	Issuer        billing.Party
	BilledTo      billing.Party
	Lines         []InvoiceLineView
	SubtotalCents int
	CreditCents   int
	TotalCents    int
}

// InvoiceLineView is an invoice line in the unit people read: hours rather
// than the minutes CPU and memory are metered in, months of 720 hours rather
// than the hours volumes and object storage are priced in.
type InvoiceLineView struct {
	Kind            string
	Plan            *pkgbilling.Plan
	Quantity        float64
	Unit            string
	UnitAmountCents float64
	AmountCents     int
}

type lineUnit struct {
	kind   string
	unit   string
	factor float64
}

var lineUnits = map[billing.Meter]lineUnit{
	billing.MeterCPU:     {kind: "CPU", unit: "vCPU-hours", factor: 60},
	billing.MeterMemory:  {kind: "MEMORY", unit: "GB-hours", factor: 60},
	billing.MeterDisk:    {kind: "VOLUMES", unit: "GB-months", factor: 720},
	billing.MeterStorage: {kind: "OBJECT_STORAGE", unit: "GB-months", factor: 720},
	billing.MeterEgress:  {kind: "EGRESS", unit: "GB", factor: 1},
}

// BillingPeriods returns the workspace's current billing period and up to
// two before it, newest first.
func (c *Client) BillingPeriods(ctx context.Context, ws string) ([]billing.Period, error) {
	customerID, err := c.billingCustomer(ctx, ws)

	if err != nil {
		return nil, err
	}

	return c.billing.Periods(ctx, customerID)
}

// Usage returns what the workspace used in the billing period that starts at
// periodStart, split by project, environment and resource, with the invoice
// that closes the period. It returns nil when the workspace never had a
// subscription.
func (c *Client) Usage(ctx context.Context, ws string, periodStart time.Time) (*UsageReport, error) {
	customerID, err := c.billingCustomer(ctx, ws)

	if err != nil {
		return nil, err
	}

	periods, err := c.billing.Periods(ctx, customerID)

	if err != nil {
		return nil, fmt.Errorf("read billing periods: %w", err)
	}

	if len(periods) == 0 {
		return nil, nil
	}

	index := slices.IndexFunc(periods, func(period billing.Period) bool {
		return period.Start.Unix() == periodStart.Unix()
	})

	if index < 0 {
		return nil, usererror.New("unknown billing period")
	}

	period := periods[index]

	rates, err := c.billing.Rates(ctx, customerID)

	if err != nil {
		return nil, fmt.Errorf("read rates: %w", err)
	}

	since, err := c.workspaceCreated(ctx, ws)

	if err != nil {
		return nil, err
	}

	projects, err := c.platform.Projects(ctx, ws)

	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	namespaces := map[string]usage.Location{}

	for _, project := range projects {
		for _, environment := range project.Environments {
			namespaces[environment.ID.Namespace()] = usage.Location{Project: project.ID.Name, Environment: environment.ID.Name}
		}
	}

	var invoice *billing.Invoice
	var invoiceErr error
	var group sync.WaitGroup

	group.Go(func() {
		invoice, invoiceErr = c.billing.Invoice(ctx, customerID, period)
	})

	split, err := c.usage.Split(ctx, usage.Request{
		Workspace:  ws,
		Period:     period,
		Since:      since,
		Rates:      rates,
		Namespaces: namespaces,
	})

	group.Wait()

	if err != nil {
		return nil, err
	}

	if invoiceErr != nil {
		return nil, fmt.Errorf("read invoice: %w", invoiceErr)
	}

	return &UsageReport{
		Period:        period,
		MeasuredUntil: split.MeasuredUntil,
		Projects:      c.projectUsage(ctx, ws, projects, split),
		Invoice:       invoiceView(invoice),
	}, nil
}

func (c *Client) projectUsage(ctx context.Context, ws string, live []platform.Project, split *usage.Split) []ProjectUsage {
	liveProjects := make(map[string]platform.Project, len(live))
	liveEnvironments := map[platform.EnvironmentID]platform.Environment{}

	for _, project := range live {
		liveProjects[project.ID.Name] = project

		for _, environment := range project.Environments {
			liveEnvironments[environment.ID] = environment
		}
	}

	byProject := map[string]*ProjectUsage{}
	var order []string

	for _, environment := range split.Environments {
		projectID := platform.ProjectID{Workspace: ws, Name: environment.Project}
		project, ok := byProject[environment.Project]

		if !ok {
			project = &ProjectUsage{ID: projectID, Name: projectID.Name, Daily: make([]float64, split.Days)}

			if liveProject, ok := liveProjects[environment.Project]; ok {
				project.Name = liveProject.Name
			}

			byProject[environment.Project] = project
			order = append(order, environment.Project)
		}

		environmentID := platform.EnvironmentID{Workspace: ws, Project: environment.Project, Name: environment.ID}
		entry := EnvironmentUsage{
			ID:    environmentID,
			Name:  environmentID.Name,
			Cents: environment.Cents,
			Daily: environment.Daily,
		}

		if liveEnvironment, ok := liveEnvironments[environmentID]; ok {
			entry.Name = liveEnvironment.Name
			entry.Resources = c.resourceUsage(ctx, environmentID, environment.Resources)
		} else {
			lastSeen := environment.LastSeen
			entry.DeletedAt = &lastSeen
			entry.Resources = resourceUsage(environmentID, environment.Resources, nil)
		}

		project.Cents += environment.Cents

		for i, cents := range environment.Daily {
			project.Daily[i] += cents
		}

		project.Environments = append(project.Environments, entry)
	}

	result := make([]ProjectUsage, 0, len(order))

	for _, name := range order {
		project := byProject[name]

		if _, ok := liveProjects[name]; !ok {
			for _, environment := range project.Environments {
				if environment.DeletedAt != nil && (project.DeletedAt == nil || environment.DeletedAt.After(*project.DeletedAt)) {
					project.DeletedAt = environment.DeletedAt
				}
			}
		}

		result = append(result, *project)
	}

	slices.SortStableFunc(result, func(a, b ProjectUsage) int {
		switch {
		case a.Cents > b.Cents:
			return -1
		case a.Cents < b.Cents:
			return 1
		}

		return 0
	})

	return result
}

func (c *Client) resourceUsage(ctx context.Context, environment platform.EnvironmentID, resources []usage.Resource) []ResourceUsage {
	names := map[usage.Kind]map[string]string{}
	add := func(kind usage.Kind, id, name string) {
		if names[kind] == nil {
			names[kind] = map[string]string{}
		}

		names[kind][id] = name
	}

	if services, err := c.cachedPlatform.Services(ctx, environment); err != nil {
		slog.WarnContext(ctx, "usage: failed to list services", "environment", environment, "error", err)
	} else {
		for _, service := range services {
			add(usage.KindService, service.ID.Name, service.Name)
		}
	}

	if databases, err := c.cachedPlatform.Databases(ctx, environment); err != nil {
		slog.WarnContext(ctx, "usage: failed to list databases", "environment", environment, "error", err)
	} else {
		for _, database := range databases {
			add(usage.KindDatabase, database.ID.Name, database.Name)
		}
	}

	if stores, err := c.cachedPlatform.KeyValueStores(ctx, environment); err != nil {
		slog.WarnContext(ctx, "usage: failed to list key-value stores", "environment", environment, "error", err)
	} else {
		for _, store := range stores {
			add(usage.KindKeyValueStore, store.ID.Name, store.Name)
		}
	}

	if volumes, err := c.cachedPlatform.Volumes(ctx, environment); err != nil {
		slog.WarnContext(ctx, "usage: failed to list volumes", "environment", environment, "error", err)
	} else {
		for _, volume := range volumes {
			add(usage.KindVolume, volume.ID.Name, volume.Name)
		}
	}

	return resourceUsage(environment, resources, names)
}

func resourceUsage(environment platform.EnvironmentID, resources []usage.Resource, names map[usage.Kind]map[string]string) []ResourceUsage {
	result := make([]ResourceUsage, 0, len(resources))

	for _, resource := range resources {
		name := resource.ID

		if readable := names[resource.Kind][resource.ID]; readable != "" {
			name = readable
		}

		result = append(result, ResourceUsage{
			ID:    environment.String() + "/" + resource.ID,
			Name:  name,
			Kind:  resource.Kind,
			Cents: resource.Cents,
			Daily: resource.Daily,
		})
	}

	return result
}

func (c *Client) workspaceCreated(ctx context.Context, ws string) (time.Time, error) {
	orgID, err := c.orgID(ctx, ws)

	if err != nil {
		return time.Time{}, fmt.Errorf("failed to resolve org: %w", err)
	}

	org, err := c.logto.Organization(ctx, orgID)

	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get organization: %w", err)
	}

	return time.UnixMilli(int64(org.CreatedAt)), nil
}

func invoiceView(invoice *billing.Invoice) *InvoiceView {
	if invoice == nil {
		return nil
	}

	view := &InvoiceView{
		Number:        invoice.Number,
		Status:        invoice.Status,
		IssuedAt:      invoice.IssuedAt,
		PaidAt:        invoice.PaidAt,
		PDFURL:        invoice.PDFURL,
		Issuer:        invoice.Issuer,
		BilledTo:      invoice.BilledTo,
		SubtotalCents: int(invoice.SubtotalCents),
		CreditCents:   int(invoice.CreditCents),
		TotalCents:    int(invoice.TotalCents),
	}

	for _, line := range invoice.Lines {
		if line.Plan != pkgbilling.PlanNone {
			plan := line.Plan
			view.Lines = append(view.Lines, InvoiceLineView{
				Kind:            "PLAN",
				Plan:            &plan,
				Quantity:        line.Quantity,
				Unit:            "month",
				UnitAmountCents: line.UnitAmountCents,
				AmountCents:     int(line.AmountCents),
			})

			continue
		}

		unit, ok := lineUnits[line.Meter]

		if !ok {
			continue
		}

		view.Lines = append(view.Lines, InvoiceLineView{
			Kind:            unit.kind,
			Quantity:        line.Quantity / unit.factor,
			Unit:            unit.unit,
			UnitAmountCents: line.UnitAmountCents * unit.factor,
			AmountCents:     int(line.AmountCents),
		})
	}

	return view
}
