package conductor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	pkgbilling "github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/billing"
)

const capabilitiesTTL = time.Minute

var (
	minVolumeSize        = resource.MustParse("10Gi")
	minDatabaseStorage   = resource.MustParse("1Gi")
	minKeyValueStoreSize = resource.MustParse("1Gi")
)

// ErrPrepaidCard is returned when a card checkout saved a prepaid card.
var ErrPrepaidCard = billing.ErrPrepaidCard

type Availability string

const (
	Available             Availability = "AVAILABLE"
	PlanRequired          Availability = "PLAN_REQUIRED"
	PaymentMethodRequired Availability = "PAYMENT_METHOD_REQUIRED"
	Unavailable           Availability = "UNAVAILABLE"
)

// Gate is whether a workspace may use something, and the cheapest plan that
// unlocks it when its plan doesn't.
type Gate struct {
	Availability Availability
	Plan         pkgbilling.Plan
}

type QuantityUpgrade struct {
	Plan pkgbilling.Plan
	Max  resource.Quantity
}

type QuantityLimit struct {
	Min      resource.Quantity
	Max      resource.Quantity
	Upgrades []QuantityUpgrade
}

type CountUpgrade struct {
	Plan  pkgbilling.Plan
	Limit int
}

type ComputeCapability struct {
	Gate
	CPU              QuantityLimit
	Memory           QuantityLimit
	Replicas         int
	ReplicasUpgrades []CountUpgrade
}

type DatabaseCapability struct {
	Gate
	CPU     QuantityLimit
	Memory  QuantityLimit
	Storage QuantityLimit
}

type StorageCapability struct {
	Gate
	Size QuantityLimit
}

type CountCapability struct {
	Gate
	Limit    int
	Upgrades []CountUpgrade
}

// Capabilities is what a workspace may create and how large, with what
// unlocks the rest. The dashboard shows it and every guard checks it.
type Capabilities struct {
	Services       ComputeCapability
	Databases      DatabaseCapability
	KeyValueStores StorageCapability
	Volumes        StorageCapability
	Buckets        Gate
	Agents         Gate
	Members        CountCapability
	CustomDomains  CountCapability
}

// GateError refuses something the workspace's plan or payment method doesn't
// allow yet.
type GateError struct {
	Gate
	Message string
}

func (e *GateError) Error() string {
	return e.Message
}

type cachedCapabilities struct {
	capabilities Capabilities
	expires      time.Time
}

// PlanOffer is what a plan includes and costs, for comparing plans.
type PlanOffer struct {
	Plan   pkgbilling.Plan
	Price  billing.Price
	Limits pkgbilling.Limits
}

// Capabilities reports what a workspace can use. Every guard checks the same
// answer, so a client never offers what the conductor would refuse.
func (c *Client) Capabilities(ctx context.Context, workspaceID string) (Capabilities, error) {
	c.capabilitiesMu.Lock()
	cached, ok := c.capabilitiesCache[workspaceID]
	c.capabilitiesMu.Unlock()

	if ok && time.Now().Before(cached.expires) {
		return cached.capabilities, nil
	}

	capabilities, err := c.computeCapabilities(ctx, workspaceID)

	if err != nil {
		slog.ErrorContext(ctx, "capabilities lookup failed", "workspace", workspaceID, "error", err)
		return Capabilities{}, usererror.New("couldn't check the workspace's plan, try again")
	}

	c.capabilitiesMu.Lock()
	c.capabilitiesCache[workspaceID] = cachedCapabilities{capabilities: capabilities, expires: time.Now().Add(capabilitiesTTL)}
	c.capabilitiesMu.Unlock()

	return capabilities, nil
}

// PlanOffers returns every plan with what it includes and costs, the free
// tier first. It is empty without billing.
func (c *Client) PlanOffers(ctx context.Context) ([]PlanOffer, error) {
	if c.billing == nil {
		return []PlanOffer{}, nil
	}

	offers := make([]PlanOffer, 0, len(pkgbilling.Plans))

	for _, plan := range pkgbilling.Plans {
		offer := PlanOffer{Plan: plan, Limits: plan.Limits()}

		if plan.Paid() {
			price, err := c.billing.PlanPrice(ctx, plan)

			if err != nil {
				return nil, err
			}

			offer.Price = price
		}

		offers = append(offers, offer)
	}

	for i := range offers {
		if offers[i].Price.Currency == "" && len(offers) > 1 {
			offers[i].Price.Currency = offers[len(offers)-1].Price.Currency
		}
	}

	return offers, nil
}

func (c *Client) forgetCapabilities(workspaceID string) {
	c.capabilitiesMu.Lock()
	delete(c.capabilitiesCache, workspaceID)
	c.capabilitiesMu.Unlock()
}

func (c *Client) freshCapabilities(ctx context.Context, workspaceID string) (Capabilities, error) {
	c.forgetCapabilities(workspaceID)

	return c.Capabilities(ctx, workspaceID)
}

func (c *Client) agentsEnabled() bool {
	return c.config.AgentsEnabled
}

func (c *Client) computeCapabilities(ctx context.Context, workspaceID string) (Capabilities, error) {
	orgID, err := c.orgID(ctx, workspaceID)

	if err != nil {
		return Capabilities{}, fmt.Errorf("resolve org: %w", err)
	}

	org, err := c.logto.Organization(ctx, orgID)

	if err != nil {
		return Capabilities{}, fmt.Errorf("read org: %w", err)
	}

	personal, _ := org.CustomData["personal"].(bool)

	if c.billing == nil {
		capabilities := capabilitiesFor(pkgbilling.InstanceLimits, nil)
		capabilities.Agents = c.agentsGate()

		if personal {
			capabilities.Members.Gate = Gate{Availability: Unavailable}
		}

		return capabilities, nil
	}

	plan := pkgbilling.PlanNone
	customerID, _ := org.CustomData["stripeCustomerId"].(string)

	if customerID != "" {
		if plan, err = c.billing.Plan(ctx, customerID); err != nil {
			return Capabilities{}, fmt.Errorf("read plan: %w", err)
		}
	}

	capabilities := capabilitiesFor(plan.Limits(), higherPlans(plan))
	capabilities.Agents = c.agentsGate()

	if err := c.countMembers(ctx, orgID, personal, &capabilities.Members); err != nil {
		return Capabilities{}, err
	}

	if err := c.countCustomDomains(ctx, workspaceID, &capabilities.CustomDomains); err != nil {
		return Capabilities{}, err
	}

	if personal && !plan.Paid() {
		required, err := c.cardRequired(ctx, orgID, customerID)

		if err != nil {
			return Capabilities{}, err
		}

		if required {
			requireCard(&capabilities)
		}
	}

	return capabilities, nil
}

func (c *Client) agentsGate() Gate {
	if !c.agentsEnabled() {
		return Gate{Availability: Unavailable}
	}

	return Gate{Availability: Available}
}

func (c *Client) countMembers(ctx context.Context, orgID string, personal bool, members *CountCapability) error {
	if personal {
		members.Gate = Gate{Availability: Unavailable}
		return nil
	}

	if members.Limit == pkgbilling.Unlimited {
		return nil
	}

	current, err := c.logto.OrganizationMembers(ctx, orgID)

	if err != nil {
		return fmt.Errorf("read members: %w", err)
	}

	members.Gate = countGate(len(current), members.Limit, members.Upgrades)

	return nil
}

func (c *Client) countCustomDomains(ctx context.Context, workspaceID string, domains *CountCapability) error {
	if domains.Limit == pkgbilling.Unlimited {
		return nil
	}

	environments, err := c.platform.EnvironmentsByWorkspace(ctx, workspaceID)

	if err != nil {
		return fmt.Errorf("read environments: %w", err)
	}

	count := 0

	for _, environment := range environments {
		services, err := c.platform.Services(ctx, environment.ID)

		if err != nil {
			return fmt.Errorf("read services: %w", err)
		}

		for _, service := range services {
			for _, endpoint := range service.Endpoints {
				if c.customHost(endpoint.Host) {
					count++
				}
			}
		}
	}

	domains.Gate = countGate(count, domains.Limit, domains.Upgrades)

	return nil
}

func (c *Client) customHost(host string) bool {
	return host != "" && !c.hostname.IsPlatform(host) && !c.hostname.IsInternal(host)
}

func (c *Client) cardRequired(ctx context.Context, orgID, customerID string) (bool, error) {
	members, err := c.logto.OrganizationMembers(ctx, orgID)

	if err != nil {
		return false, fmt.Errorf("read members: %w", err)
	}

	if !youngAccount(members, c.config.CardGateAccountAge) {
		return false, nil
	}

	if customerID == "" {
		return true, nil
	}

	usable, err := c.billing.UsableCard(ctx, customerID)

	if err != nil {
		return false, fmt.Errorf("read cards: %w", err)
	}

	return !usable, nil
}

func youngAccount(members []logto.OrganizationMember, threshold time.Duration) bool {
	if threshold <= 0 || len(members) == 0 {
		return false
	}

	owner := members[0]

	for _, identity := range owner.Identities {
		created, ok := identity.CreatedAt()

		if !ok || owner.CreatedAt == 0 {
			continue
		}

		signedUp := time.UnixMilli(owner.CreatedAt)

		return signedUp.Sub(created) < threshold
	}

	return false
}

func requireCard(capabilities *Capabilities) {
	for _, gate := range []*Gate{
		&capabilities.Services.Gate,
		&capabilities.Databases.Gate,
		&capabilities.KeyValueStores.Gate,
		&capabilities.Volumes.Gate,
		&capabilities.Buckets,
		&capabilities.Agents,
	} {
		if gate.Availability == Available {
			*gate = Gate{Availability: PaymentMethodRequired}
		}
	}
}

func capabilitiesFor(limits pkgbilling.Limits, higher []pkgbilling.Plan) Capabilities {
	cpu := quantityLimit(minServiceCPU, limits.CPU, higher, func(l pkgbilling.Limits) resource.Quantity { return l.CPU })
	memory := quantityLimit(minServiceMemory, limits.Memory, higher, func(l pkgbilling.Limits) resource.Quantity { return l.Memory })

	volumes := StorageCapability{
		Gate: Gate{Availability: Available},
		Size: quantityLimit(minVolumeSize, limits.VolumeSize, higher, func(l pkgbilling.Limits) resource.Quantity {
			if !l.Volumes {
				return resource.Quantity{}
			}

			return l.VolumeSize
		}),
	}

	if !limits.Volumes {
		volumes.Gate = featureGate(higher, func(l pkgbilling.Limits) bool { return l.Volumes })
	}

	return Capabilities{
		Services: ComputeCapability{
			Gate:             Gate{Availability: Available},
			CPU:              cpu,
			Memory:           memory,
			Replicas:         limits.Replicas,
			ReplicasUpgrades: countUpgrades(limits.Replicas, higher, func(l pkgbilling.Limits) int { return l.Replicas }),
		},
		Databases: DatabaseCapability{
			Gate:    Gate{Availability: Available},
			CPU:     cpu,
			Memory:  memory,
			Storage: quantityLimit(minDatabaseStorage, limits.DatabaseStorage, higher, func(l pkgbilling.Limits) resource.Quantity { return l.DatabaseStorage }),
		},
		KeyValueStores: StorageCapability{
			Gate: Gate{Availability: Available},
			Size: quantityLimit(minKeyValueStoreSize, limits.KeyValueStoreSize, higher, func(l pkgbilling.Limits) resource.Quantity { return l.KeyValueStoreSize }),
		},
		Volumes: volumes,
		Buckets: Gate{Availability: Available},
		Members: CountCapability{
			Gate:     Gate{Availability: Available},
			Limit:    limits.Members,
			Upgrades: countUpgrades(limits.Members, higher, func(l pkgbilling.Limits) int { return l.Members }),
		},
		CustomDomains: CountCapability{
			Gate:     Gate{Availability: Available},
			Limit:    limits.CustomDomains,
			Upgrades: countUpgrades(limits.CustomDomains, higher, func(l pkgbilling.Limits) int { return l.CustomDomains }),
		},
	}
}

func higherPlans(plan pkgbilling.Plan) []pkgbilling.Plan {
	for i, candidate := range pkgbilling.Plans {
		if candidate == plan {
			return pkgbilling.Plans[i+1:]
		}
	}

	return nil
}

func quantityLimit(min, max resource.Quantity, higher []pkgbilling.Plan, of func(pkgbilling.Limits) resource.Quantity) QuantityLimit {
	limit := QuantityLimit{Min: min.DeepCopy(), Max: max.DeepCopy(), Upgrades: []QuantityUpgrade{}}
	ceiling := max.DeepCopy()

	for _, plan := range higher {
		value := of(plan.Limits())

		if value.Cmp(ceiling) > 0 {
			limit.Upgrades = append(limit.Upgrades, QuantityUpgrade{Plan: plan, Max: value.DeepCopy()})
			ceiling = value.DeepCopy()
		}
	}

	return limit
}

func countUpgrades(limit int, higher []pkgbilling.Plan, of func(pkgbilling.Limits) int) []CountUpgrade {
	upgrades := []CountUpgrade{}
	ceiling := limit

	for _, plan := range higher {
		if ceiling == pkgbilling.Unlimited {
			break
		}

		value := of(plan.Limits())

		if value == pkgbilling.Unlimited || value > ceiling {
			upgrades = append(upgrades, CountUpgrade{Plan: plan, Limit: value})
			ceiling = value
		}
	}

	return upgrades
}

func featureGate(higher []pkgbilling.Plan, included func(pkgbilling.Limits) bool) Gate {
	for _, plan := range higher {
		if included(plan.Limits()) {
			return Gate{Availability: PlanRequired, Plan: plan}
		}
	}

	return Gate{Availability: Unavailable}
}

func countGate(count, limit int, upgrades []CountUpgrade) Gate {
	if limit == pkgbilling.Unlimited || count < limit {
		return Gate{Availability: Available}
	}

	for _, upgrade := range upgrades {
		if upgrade.Limit == pkgbilling.Unlimited || count < upgrade.Limit {
			return Gate{Availability: PlanRequired, Plan: upgrade.Plan}
		}
	}

	return Gate{Availability: Unavailable}
}

func (g Gate) allow(what string) error {
	switch g.Availability {
	case Available:
		return nil
	case PlanRequired:
		return &GateError{Gate: g, Message: fmt.Sprintf("%s needs the %s plan", what, g.Plan.Title())}
	case PaymentMethodRequired:
		return &GateError{Gate: g, Message: "Add a card to the workspace to start creating. Nothing is charged until you choose a plan."}
	}

	return usererror.Errorf("%s isn't available here", what)
}

func (l QuantityLimit) check(what string, value resource.Quantity, current *resource.Quantity) error {
	if value.Cmp(l.Min) < 0 {
		return usererror.Errorf("%s must be at least %s", what, l.Min.String())
	}

	if value.Cmp(l.Max) <= 0 || (current != nil && value.Cmp(*current) <= 0) {
		return nil
	}

	for _, upgrade := range l.Upgrades {
		if value.Cmp(upgrade.Max) <= 0 {
			return &GateError{
				Gate:    Gate{Availability: PlanRequired, Plan: upgrade.Plan},
				Message: fmt.Sprintf("%s of %s needs the %s plan", what, value.String(), upgrade.Plan.Title()),
			}
		}
	}

	return usererror.Errorf("%s exceeds the maximum of %s", what, l.Max.String())
}

func (c ComputeCapability) checkReplicas(value, current int) error {
	if value <= c.Replicas || value <= current {
		return nil
	}

	for _, upgrade := range c.ReplicasUpgrades {
		if upgrade.Limit == pkgbilling.Unlimited || value <= upgrade.Limit {
			return &GateError{
				Gate:    Gate{Availability: PlanRequired, Plan: upgrade.Plan},
				Message: fmt.Sprintf("%d replicas need the %s plan", value, upgrade.Plan.Title()),
			}
		}
	}

	return usererror.Errorf("replicas exceed the maximum of %d", c.Replicas)
}

// AsGateError returns the gate err refuses on, if it is one.
func AsGateError(err error) (*GateError, bool) {
	var gate *GateError
	ok := errors.As(err, &gate)

	return gate, ok
}
