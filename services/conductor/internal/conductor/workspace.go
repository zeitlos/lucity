package conductor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/billing"
	"github.com/zeitlos/lucity/services/conductor/internal/directory"
)

type Workspace = directory.Workspace
type WorkspaceDetails = directory.WorkspaceDetails
type WorkspaceMember = directory.WorkspaceMember

var workspaceIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

func (c *Client) Workspaces(ctx context.Context) ([]Workspace, error) {
	claims, err := auth.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	workspaces, err := c.directory.WorkspacesForUser(ctx, claims.Subject)

	if err != nil {
		return nil, err
	}

	if len(workspaces) == 0 {
		if _, _, ensureErr := c.EnsureAccount(ctx, claims.Subject); ensureErr == nil {
			workspaces, err = c.directory.WorkspacesForUser(ctx, claims.Subject)
			if err != nil {
				return nil, err
			}
		}
	}

	return workspaces, nil
}

// EnsureAccount idempotently provisions a user's Logto username and personal
// workspace. It runs lazily on first authenticated access, since clients now
// sign in directly with the identity provider.
func (c *Client) EnsureAccount(ctx context.Context, userID string) (string, bool, error) {
	login, err := c.logto.SocialLogin(ctx, userID)
	if err != nil {
		return "", false, err
	}

	username, err := c.logto.EnsureUsername(ctx, userID, login)
	if err != nil {
		return "", false, err
	}

	return c.EnsurePersonalWorkspace(ctx, userID, username)
}

func (c *Client) Workspace(ctx context.Context, id string) (*WorkspaceDetails, error) {
	return c.directory.Workspace(ctx, id)
}

func (c *Client) UpdateWorkspace(ctx context.Context, id, name string) (*WorkspaceDetails, error) {
	return c.directory.UpdateWorkspace(ctx, id, name)
}

func (c *Client) DeleteWorkspace(ctx context.Context, id string) (bool, error) {
	projects, err := c.platform.Projects(ctx, id)

	if err != nil {
		return false, err
	}

	if len(projects) > 0 {
		return false, usererror.Errorf("cannot delete workspace: %d projects still exist, delete them first", len(projects))
	}

	if c.billing != nil {
		customerID, err := c.stripeCustomerID(ctx, id)

		if err != nil {
			return false, err
		}

		if customerID != "" {
			renews, err := c.billing.Renews(ctx, customerID)

			if err != nil {
				return false, fmt.Errorf("read subscription: %w", err)
			}

			if renews {
				return false, usererror.New("cannot delete workspace: its plan is still active, cancel it in billing first")
			}
		}
	}

	if err := c.deployer.Workspaces().Delete(ctx, id); err != nil {
		return false, fmt.Errorf("delete agent workspace release: %w", err)
	}

	if err := c.directory.DeleteWorkspace(ctx, id); err != nil {
		return false, err
	}

	if err := c.domains.DeleteZone(ctx, id); err != nil {
		slog.WarnContext(ctx, "edge zone delete failed", "workspace", id, "error", err)
	}

	slog.InfoContext(ctx, "workspace deleted")

	return true, nil
}

func (c *Client) InviteMember(ctx context.Context, workspaceID, email string, role auth.WorkspaceRole) (*WorkspaceMember, error) {
	member, err := c.directory.InviteMember(ctx, workspaceID, email, role)

	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "member invited", "email", email, "role", role)

	return member, nil
}

func (c *Client) RemoveMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	claims, err := auth.FromContext(ctx)

	if err != nil {
		return false, err
	}

	if claims.Subject == userID {
		return false, usererror.New("cannot remove yourself from workspace")
	}

	if err := c.directory.RemoveMember(ctx, workspaceID, userID); err != nil {
		return false, err
	}

	slog.InfoContext(ctx, "member removed", "user_id", userID)

	return true, nil
}

func (c *Client) UpdateMemberRole(ctx context.Context, workspaceID, userID string, role auth.WorkspaceRole) (*WorkspaceMember, error) {
	member, err := c.directory.UpdateMemberRole(ctx, workspaceID, userID, role)

	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "member role updated", "user_id", userID, "role", role)

	return member, nil
}

// EnsurePersonalWorkspace creates a personal workspace for a new user if they have none.
// The workspace ID is derived from the user's username. Idempotent: if the workspace
// already exists and belongs to this user, returns the existing ID.
// On genuine collision (different owner), picks {id}-0, {id}-1, etc.
// Returns the workspace ID and whether it was newly created (true) or restored (false).
func (c *Client) EnsurePersonalWorkspace(ctx context.Context, userID, username string) (string, bool, error) {
	if c.logto == nil {
		return "", false, fmt.Errorf("logto not configured")
	}

	wsID := sanitizeWorkspaceID(username)
	if wsID == "" {
		return "", false, fmt.Errorf("cannot derive workspace ID from username %q", username)
	}

	adminRoleID, memberRoleID, err := c.orgRoleIDs(ctx)
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve org role IDs: %w", err)
	}

	// Check if preferred workspace ID already exists (search by name).
	existing, err := c.logto.OrganizationByName(ctx, wsID)
	if err == nil {
		// Cache the mapping
		c.cacheOrgID(wsID, existing.ID)

		// Workspace exists. Check if this user is already a member.
		members, memErr := c.logto.OrganizationMembers(ctx, existing.ID)
		if memErr == nil {
			for _, m := range members {
				if m.ID == userID {
					// User is already a member. Self-heal billing in the background.
					go func() {
						email := ""
						if user, _ := c.logto.User(context.WithoutCancel(ctx), userID); user != nil {
							email = user.PrimaryEmail
						}
						c.setupBilling(ctx, wsID, username, email, 14)
					}()
					slog.Info("personal workspace restored", "id", wsID, "user", userID)
					return wsID, false, nil
				}
			}
		}

		// Check if this is a personal workspace with customData indicating personal=true.
		// If personal and no owner matched, it might be a pre-migration workspace.
		isPersonal, _ := existing.CustomData["personal"].(bool)
		if isPersonal {
			// Re-add user as admin member
			_ = c.logto.AddOrganizationMember(ctx, existing.ID, userID)
			_ = c.logto.AssignOrganizationRoles(ctx, existing.ID, userID, []string{adminRoleID, memberRoleID})
			slog.Info("personal workspace restored (re-added member)", "id", wsID, "user", userID)
			return wsID, false, nil
		}

		// Genuine collision — someone else owns this ID.
		wsID, err = c.findAvailableWorkspaceID(ctx, wsID)
		if err != nil {
			return "", false, fmt.Errorf("failed to find available workspace ID: %w", err)
		}
	}

	// Create Logto organization with personal=true in customData.
	customData := map[string]interface{}{
		"personal": true,
	}
	org, err := c.logto.CreateOrganization(ctx, wsID, username, customData)
	if err != nil {
		return "", false, fmt.Errorf("failed to create personal workspace: %w", err)
	}

	// Cache the org ID mapping
	c.cacheOrgID(wsID, org.ID)

	// Add user as admin member
	if err := c.logto.AddOrganizationMember(ctx, org.ID, userID); err != nil {
		return "", false, fmt.Errorf("failed to add user to personal workspace: %w", err)
	}
	if err := c.logto.AssignOrganizationRoles(ctx, org.ID, userID, []string{adminRoleID, memberRoleID}); err != nil {
		return "", false, fmt.Errorf("failed to assign roles in personal workspace: %w", err)
	}

	slog.Info("personal workspace created", "id", wsID, "user", userID)

	// Best-effort: set up Stripe customer + subscription with 14-day trial
	user, _ := c.logto.User(ctx, userID)
	email := ""
	if user != nil {
		email = user.PrimaryEmail
	}
	c.setupBilling(ctx, wsID, username, email, 14)

	return wsID, true, nil
}

// findAvailableWorkspaceID tries {base}-0, {base}-1, ... up to {base}-9
// to find an available workspace ID. Returns error if all slots are taken.
func (c *Client) findAvailableWorkspaceID(ctx context.Context, base string) (string, error) {
	for i := 0; i < 10; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		_, err := c.logto.OrganizationByName(ctx, candidate)
		if err != nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("all workspace ID slots exhausted for base %q", base)
}

// setupBilling ensures a Stripe customer and subscription exist for a workspace and
// stores the customer ID in the Logto org customData. Idempotent: cashier reuses the
// customer's subscription and never starts a second trial, so this self-heals on
// every login if a previous attempt partially failed.
//
// creditDays > 0 creates a promotional credit grant that expires after that many days.
// Personal workspaces get 14 days of credits; additional workspaces get 0 (no promo credits).
//
// Uses context.WithoutCancel to detach from the HTTP request lifecycle. Billing setup
// must complete even if the browser navigates away during the OIDC callback redirect.
func (c *Client) setupBilling(ctx context.Context, workspace, name, email string, creditDays int) {
	if c.billing == nil {
		return
	}
	if c.logto == nil {
		return
	}

	// Detach from the HTTP request context. The OIDC callback redirects the browser
	// immediately after this returns, which cancels r.Context(). Billing setup must
	// survive that cancellation.
	billingCtx := context.WithoutCancel(ctx)

	// Resolve workspace ID to Logto org ID for API calls
	orgID, err := c.orgID(billingCtx, workspace)
	if err != nil {
		slog.Warn("failed to resolve org ID for billing setup", "workspace", workspace, "error", err)
		return
	}

	// Read current org customData to check if billing is already (partially) set up.
	org, err := c.logto.Organization(billingCtx, orgID)
	if err != nil {
		slog.Warn("failed to read org for billing setup", "workspace", workspace, "error", err)
		return
	}
	customData := org.CustomData
	if customData == nil {
		customData = make(map[string]interface{})
	}

	// Step 1: Ensure Stripe customer exists.
	customerID, _ := customData["stripeCustomerId"].(string)
	if customerID == "" {
		customerID, err = c.billing.CreateCustomer(billingCtx, workspace, name, email)
		if err != nil {
			slog.Warn("failed to create Stripe customer for workspace", "workspace", workspace, "error", err)
			return // Will retry on next login
		}
		customData["stripeCustomerId"] = customerID
		if _, err := c.logto.UpdateOrganizationCustomData(billingCtx, orgID, customData); err != nil {
			slog.Warn("failed to store billing IDs in org customData", "workspace", workspace, "error", err)
			return // Will retry on next login
		}
	}

	// Step 2: Ensure Stripe subscription exists (metered items only, no plan).
	err = c.billing.StartTrial(billingCtx, customerID, workspace, creditDays)
	if errors.Is(err, billing.ErrSubscriptionEnded) {
		slog.Debug("subscription ended, no new trial", "workspace", workspace)
		return
	}
	if err != nil {
		slog.Warn("failed to create Stripe subscription for workspace", "workspace", workspace, "error", err)
		return // Will retry subscription on next login
	}

	slog.Debug("billing set up", "workspace", workspace, "customer_id", customerID)
}

// CreateWorkspaceCheckout creates a workspace on a paid plan and returns the
// URL of the checkout that pays for it. The workspace exists, suspended, from
// the start; completing the checkout lifts the suspension.
func (c *Client) CreateWorkspaceCheckout(ctx context.Context, id, name, plan string) (string, error) {
	claims, err := auth.FromContext(ctx)

	if err != nil {
		return "", err
	}

	if c.logto == nil || c.billing == nil {
		return "", errors.New("billing not configured")
	}

	if !workspaceIDPattern.MatchString(id) {
		return "", usererror.New("invalid workspace ID: must be 3-63 lowercase alphanumeric characters or hyphens")
	}

	if _, err = c.logto.OrganizationByName(ctx, id); err == nil {
		return "", usererror.Errorf("workspace ID %q is already taken", id)
	}

	customerID, err := c.billing.CreateCustomer(ctx, id, name, claims.Email)

	if err != nil {
		return "", err
	}

	if _, err := c.directory.CreateWorkspace(ctx, id, name, map[string]any{
		"stripeCustomerId": customerID,
		"suspended":        true,
	}); err != nil {
		return "", fmt.Errorf("failed to create workspace: %w", err)
	}

	selection, err := c.billing.SelectPlan(ctx, customerID, id, planOf(plan), billing.Checkout{
		SuccessURL: c.config.DashboardURL + "/checkout/success?session_id={CHECKOUT_SESSION_ID}",
		CancelURL:  c.config.DashboardURL,
		Metadata: map[string]string{
			"workspace_id":   id,
			"workspace_name": name,
			"user_id":        claims.Subject,
		},
	})

	if err != nil {
		return "", err
	}

	if selection.CheckoutURL == "" {
		return "", errors.New("a new workspace got no checkout")
	}

	slog.InfoContext(ctx, "workspace checkout initiated", "workspace", id, "plan", plan, "user", claims.Email)

	return selection.CheckoutURL, nil
}

// CompleteWorkspaceCheckout verifies a completed workspace checkout, lifts the
// workspace's suspension and returns it. Called after the user is redirected
// back from Stripe.
func (c *Client) CompleteWorkspaceCheckout(ctx context.Context, sessionID string) (*WorkspaceDetails, error) {
	claims, err := auth.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	if c.billing == nil {
		return nil, errors.New("billing not configured")
	}

	session, err := c.billing.Session(ctx, sessionID)

	if err != nil {
		return nil, err
	}

	if !session.Complete {
		return nil, usererror.New("the checkout is not complete")
	}

	if session.Metadata["user_id"] != claims.Subject {
		return nil, usererror.New("checkout session does not belong to current user")
	}

	wsID := session.Metadata["workspace_id"]

	customerID, err := c.stripeCustomerID(ctx, wsID)

	if err != nil {
		return nil, err
	}

	if customerID == "" || customerID != session.CustomerID {
		return nil, usererror.New("checkout session does not belong to this workspace")
	}

	c.resume(ctx, wsID, customerID)

	workspace, err := c.directory.Workspace(ctx, wsID)

	if err != nil {
		return nil, err
	}

	for _, member := range workspace.Members {
		if member.ID == claims.Subject {
			slog.InfoContext(ctx, "workspace checkout completed", "workspace", wsID)
			return workspace, nil
		}
	}

	return nil, usererror.Errorf("workspace ID %q is already taken", wsID)
}

func (c *Client) orgID(ctx context.Context, workspaceID string) (string, error) {
	// Check cache under read lock
	c.orgIDCacheMu.RLock()
	if orgID, ok := c.orgIDCache[workspaceID]; ok {
		c.orgIDCacheMu.RUnlock()
		return orgID, nil
	}
	c.orgIDCacheMu.RUnlock()

	// Cache miss: look up by name
	org, err := c.logto.OrganizationByName(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve org ID for workspace %q: %w", workspaceID, err)
	}

	c.cacheOrgID(workspaceID, org.ID)
	return org.ID, nil
}

// cacheOrgID stores a workspace ID to Logto org ID mapping in the cache.
func (c *Client) cacheOrgID(workspaceID, logtoOrgID string) {
	c.orgIDCacheMu.Lock()
	c.orgIDCache[workspaceID] = logtoOrgID
	c.orgIDCacheMu.Unlock()
}

// sanitizeWorkspaceID converts a username to a valid workspace ID.
func sanitizeWorkspaceID(login string) string {
	id := strings.ToLower(login)
	// Replace any non-alphanumeric characters (except hyphens) with hyphens
	var b strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	id = strings.Trim(b.String(), "-")
	if len(id) < 3 {
		return ""
	}
	if len(id) > 63 {
		id = id[:63]
	}
	return id
}
