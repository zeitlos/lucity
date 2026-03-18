package handler

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/cashier"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/gateway/logto"
)

// Workspace represents a workspace with metadata and members.
type Workspace struct {
	ID        string
	Name      string
	Personal  bool
	Suspended bool
	Members   []WorkspaceMember
}

// WorkspaceMember represents a user's membership in a workspace.
type WorkspaceMember struct {
	ID    string
	Email string
	Name  string
	Role  auth.WorkspaceRole
}

var workspaceIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

// resolveOrgID resolves a workspace ID (org name) to Logto's internal org ID.
// Uses an in-memory cache to avoid repeated API calls.
func (c *Client) resolveOrgID(ctx context.Context, workspaceID string) (string, error) {
	// Check cache under read lock
	c.orgIDCacheMu.RLock()
	if orgID, ok := c.orgIDCache[workspaceID]; ok {
		c.orgIDCacheMu.RUnlock()
		return orgID, nil
	}
	c.orgIDCacheMu.RUnlock()

	// Cache miss: look up by name
	org, err := c.Logto.OrganizationByName(ctx, workspaceID)
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

// invalidateOrgID removes a workspace ID from the org ID cache.
func (c *Client) invalidateOrgID(workspaceID string) {
	c.orgIDCacheMu.Lock()
	delete(c.orgIDCache, workspaceID)
	c.orgIDCacheMu.Unlock()
}

// displayNameFromOrg extracts the display name from a Logto organization.
// Returns customData.displayName if set, otherwise falls back to org.Name.
func displayNameFromOrg(org logto.UserOrganization) string {
	if dn, ok := org.CustomData["displayName"].(string); ok && dn != "" {
		return dn
	}
	return org.Name
}

// displayNameFromOrgData extracts the display name from an Organization's custom data.
func displayNameFromOrgData(org *logto.Organization) string {
	if org.CustomData != nil {
		if dn, ok := org.CustomData["displayName"].(string); ok && dn != "" {
			return dn
		}
	}
	return org.Name
}

// Workspaces returns all workspaces the current user is a member of.
// Fetches user's organizations from Logto.
func (c *Client) Workspaces(ctx context.Context) ([]Workspace, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}
	if c.Logto == nil {
		return nil, fmt.Errorf("logto not configured")
	}

	orgs, err := c.Logto.UserOrganizations(ctx, claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user organizations: %w", err)
	}

	workspaces := make([]Workspace, 0, len(orgs))
	for _, org := range orgs {
		personal, _ := org.CustomData["personal"].(bool)
		suspended, _ := org.CustomData["suspended"].(bool)

		// Cache the org ID mapping while we have it
		c.cacheOrgID(org.Name, org.ID)

		workspaces = append(workspaces, Workspace{
			ID:        org.Name,
			Name:      displayNameFromOrg(org),
			Personal:  personal,
			Suspended: suspended,
		})
	}

	return workspaces, nil
}

// Workspace returns metadata and members for the active workspace.
func (c *Client) Workspace(ctx context.Context) (*Workspace, error) {
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}
	if c.Logto == nil {
		return nil, fmt.Errorf("logto not configured")
	}

	orgID, err := c.resolveOrgID(ctx, ws)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace: %w", err)
	}

	org, err := c.Logto.Organization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	personal, _ := org.CustomData["personal"].(bool)
	suspended, _ := org.CustomData["suspended"].(bool)

	result := &Workspace{
		ID:        ws,
		Name:      displayNameFromOrgData(org),
		Personal:  personal,
		Suspended: suspended,
	}

	// Fetch members
	members, err := c.WorkspaceMembers(ctx)
	if err != nil {
		slog.Warn("failed to get workspace members", "workspace", ws, "error", err)
	} else {
		result.Members = members
	}

	return result, nil
}

// WorkspaceMembers returns all members of the active workspace from Logto.
func (c *Client) WorkspaceMembers(ctx context.Context) ([]WorkspaceMember, error) {
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}
	if c.Logto == nil {
		return nil, fmt.Errorf("logto not configured")
	}

	orgID, err := c.resolveOrgID(ctx, ws)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace: %w", err)
	}

	logtoMembers, err := c.Logto.OrganizationMembers(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization members: %w", err)
	}

	members := make([]WorkspaceMember, 0, len(logtoMembers))
	for _, m := range logtoMembers {
		role := auth.WorkspaceRoleUser
		for _, r := range m.OrgRoles {
			if r.Name == "admin" {
				role = auth.WorkspaceRoleAdmin
				break
			}
		}
		members = append(members, WorkspaceMember{
			ID:    m.ID,
			Email: m.Email,
			Name:  m.Name,
			Role:  role,
		})
	}

	return members, nil
}

// CreateWorkspace creates a new workspace as a Logto organization.
// The creator is automatically added as admin.
func (c *Client) CreateWorkspace(ctx context.Context, id, name string) (*Workspace, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}
	if c.Logto == nil {
		return nil, fmt.Errorf("logto not configured")
	}

	if !workspaceIDPattern.MatchString(id) {
		return nil, fmt.Errorf("invalid workspace ID: must be 3-63 lowercase alphanumeric characters or hyphens")
	}

	// Check if workspace ID is already taken by searching by name.
	_, err := c.Logto.OrganizationByName(ctx, id)
	if err == nil {
		return nil, fmt.Errorf("workspace ID %q is already taken", id)
	}

	adminRoleID, memberRoleID, err := c.orgRoleIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve org role IDs: %w", err)
	}

	// Create Logto organization (name=id, displayName=name in customData)
	org, err := c.Logto.CreateOrganization(ctx, id, name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Cache the org ID mapping
	c.cacheOrgID(id, org.ID)

	// Add creator as member + assign admin and member roles
	if err := c.Logto.AddOrganizationMember(ctx, org.ID, claims.Subject); err != nil {
		return nil, fmt.Errorf("failed to add creator to organization: %w", err)
	}
	if err := c.Logto.AssignOrganizationRoles(ctx, org.ID, claims.Subject, []string{adminRoleID, memberRoleID}); err != nil {
		return nil, fmt.Errorf("failed to assign admin role to creator: %w", err)
	}

	slog.Info("workspace created", "id", id, "name", name, "creator", claims.Email)

	// Best-effort: set up Stripe customer + subscription with no trial.
	// Additional workspaces require a payment method — no free trial.
	c.setupBilling(ctx, id, name, claims.Email, 0)

	// Fetch the user to get their name for the member list
	user, _ := c.Logto.User(ctx, claims.Subject)
	memberName := ""
	memberEmail := claims.Email
	if user != nil {
		memberName = user.Name
		if user.PrimaryEmail != "" {
			memberEmail = user.PrimaryEmail
		}
	}

	return &Workspace{
		ID:   id,
		Name: name,
		Members: []WorkspaceMember{
			{ID: claims.Subject, Email: memberEmail, Name: memberName, Role: auth.WorkspaceRoleAdmin},
		},
	}, nil
}

// UpdateWorkspace updates the workspace display name. Admin-only.
func (c *Client) UpdateWorkspace(ctx context.Context, name string) (*Workspace, error) {
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}

	if err := c.requireWorkspaceAdmin(ctx, ws); err != nil {
		return nil, err
	}
	if c.Logto == nil {
		return nil, fmt.Errorf("logto not configured")
	}

	orgID, err := c.resolveOrgID(ctx, ws)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace: %w", err)
	}

	// Store display name in customData (org name stays as workspace ID)
	customData := map[string]interface{}{}
	if name != ws {
		customData["displayName"] = name
	}

	// Read existing customData to preserve other fields
	org, orgErr := c.Logto.Organization(ctx, orgID)
	if orgErr == nil && org.CustomData != nil {
		for k, v := range org.CustomData {
			if k != "displayName" {
				customData[k] = v
			}
		}
		if name != ws {
			customData["displayName"] = name
		}
	}

	if err := c.Logto.UpdateOrganizationCustomData(ctx, orgID, customData); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return c.Workspace(ctx)
}

// DeleteWorkspace deletes the active workspace. Admin-only. Errors if projects exist.
func (c *Client) DeleteWorkspace(ctx context.Context) (bool, error) {
	ws, err := tenant.Require(ctx)
	if err != nil {
		return false, err
	}

	if err := c.requireWorkspaceAdmin(ctx, ws); err != nil {
		return false, err
	}
	if c.Logto == nil {
		return false, fmt.Errorf("logto not configured")
	}

	orgID, err := c.resolveOrgID(ctx, ws)
	if err != nil {
		return false, fmt.Errorf("failed to resolve workspace: %w", err)
	}

	// Check if workspace is personal
	org, err := c.Logto.Organization(ctx, orgID)
	if err != nil {
		return false, fmt.Errorf("failed to get organization: %w", err)
	}
	if personal, _ := org.CustomData["personal"].(bool); personal {
		return false, fmt.Errorf("cannot delete personal workspace")
	}

	// Check no projects exist
	outCtx := auth.OutgoingContext(ctx)
	projCtx := tenant.OutgoingContext(outCtx)
	listCtx, listCancel := context.WithTimeout(projCtx, grpcTimeout)
	defer listCancel()
	resp, err := c.Packager.ListProjects(listCtx, &packager.ListProjectsRequest{})
	if err != nil {
		return false, fmt.Errorf("failed to check projects: %w", err)
	}
	if len(resp.Projects) > 0 {
		return false, fmt.Errorf("cannot delete workspace: %d projects still exist — delete them first", len(resp.Projects))
	}

	// Delete Logto organization (removes all members automatically)
	if err := c.Logto.DeleteOrganization(ctx, orgID); err != nil {
		return false, fmt.Errorf("failed to delete organization: %w", err)
	}

	// Invalidate cache
	c.invalidateOrgID(ws)

	slog.Info("workspace deleted", "id", ws)
	return true, nil
}

// InviteMember adds a user to the active workspace. Admin-only.
func (c *Client) InviteMember(ctx context.Context, email string, role auth.WorkspaceRole) (*WorkspaceMember, error) {
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}

	if err := c.requireWorkspaceAdmin(ctx, ws); err != nil {
		return nil, err
	}
	if c.Logto == nil {
		return nil, fmt.Errorf("logto not configured")
	}

	orgID, err := c.resolveOrgID(ctx, ws)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace: %w", err)
	}

	// Find user by email
	user, err := c.Logto.UserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user with email %q not found in identity provider", email)
	}

	adminRoleID, memberRoleID, err := c.orgRoleIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve org role IDs: %w", err)
	}

	// Add user to organization
	if err := c.Logto.AddOrganizationMember(ctx, orgID, user.ID); err != nil {
		return nil, fmt.Errorf("failed to add user to organization: %w", err)
	}

	// Assign role(s)
	roleIDs := []string{memberRoleID}
	if role == auth.WorkspaceRoleAdmin {
		roleIDs = append(roleIDs, adminRoleID)
	}
	if err := c.Logto.AssignOrganizationRoles(ctx, orgID, user.ID, roleIDs); err != nil {
		return nil, fmt.Errorf("failed to assign role to member: %w", err)
	}

	slog.Info("member invited", "workspace", ws, "email", email, "role", role)

	return &WorkspaceMember{
		ID:    user.ID,
		Email: user.PrimaryEmail,
		Name:  user.Name,
		Role:  role,
	}, nil
}

// RemoveMember removes a user from the active workspace. Admin-only.
func (c *Client) RemoveMember(ctx context.Context, userID string) (bool, error) {
	ws, err := tenant.Require(ctx)
	if err != nil {
		return false, err
	}

	if err := c.requireWorkspaceAdmin(ctx, ws); err != nil {
		return false, err
	}
	if c.Logto == nil {
		return false, fmt.Errorf("logto not configured")
	}

	orgID, err := c.resolveOrgID(ctx, ws)
	if err != nil {
		return false, fmt.Errorf("failed to resolve workspace: %w", err)
	}

	// Prevent removing yourself
	claims := auth.FromContext(ctx)
	if claims != nil && claims.Subject == userID {
		return false, fmt.Errorf("cannot remove yourself from workspace")
	}

	if err := c.Logto.RemoveOrganizationMember(ctx, orgID, userID); err != nil {
		return false, fmt.Errorf("failed to remove member from organization: %w", err)
	}

	slog.Info("member removed", "workspace", ws, "user_id", userID)
	return true, nil
}

// UpdateMemberRole changes a member's role in the active workspace. Admin-only.
func (c *Client) UpdateMemberRole(ctx context.Context, userID string, role auth.WorkspaceRole) (*WorkspaceMember, error) {
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}

	if err := c.requireWorkspaceAdmin(ctx, ws); err != nil {
		return nil, err
	}
	if c.Logto == nil {
		return nil, fmt.Errorf("logto not configured")
	}

	orgID, err := c.resolveOrgID(ctx, ws)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace: %w", err)
	}

	adminRoleID, _, err := c.orgRoleIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve org role IDs: %w", err)
	}

	if role == auth.WorkspaceRoleAdmin {
		// Assign admin role
		if err := c.Logto.AssignOrganizationRoles(ctx, orgID, userID, []string{adminRoleID}); err != nil {
			return nil, fmt.Errorf("failed to assign admin role: %w", err)
		}
	} else {
		// Remove admin role
		if err := c.Logto.RemoveOrganizationRole(ctx, orgID, userID, adminRoleID); err != nil {
			return nil, fmt.Errorf("failed to remove admin role: %w", err)
		}
	}

	// Fetch updated user info
	user, err := c.Logto.User(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &WorkspaceMember{
		ID:    user.ID,
		Email: user.PrimaryEmail,
		Name:  user.Name,
		Role:  role,
	}, nil
}

// EnsurePersonalWorkspace creates a personal workspace for a new user if they have none.
// The workspace ID is derived from the user's username. Idempotent: if the workspace
// already exists and belongs to this user, returns the existing ID.
// On genuine collision (different owner), picks {id}-0, {id}-1, etc.
// Returns the workspace ID and whether it was newly created (true) or restored (false).
func (c *Client) EnsurePersonalWorkspace(ctx context.Context, userID, username string) (string, bool, error) {
	if c.Logto == nil {
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
	existing, err := c.Logto.OrganizationByName(ctx, wsID)
	if err == nil {
		// Cache the mapping
		c.cacheOrgID(wsID, existing.ID)

		// Workspace exists. Check if this user is already a member.
		members, memErr := c.Logto.OrganizationMembers(ctx, existing.ID)
		if memErr == nil {
			for _, m := range members {
				if m.ID == userID {
					// User is already a member. Self-heal billing if needed.
					if existing.CustomData != nil {
						stripeCustomerID, _ := existing.CustomData["stripeCustomerId"].(string)
						stripeSubID, _ := existing.CustomData["stripeSubscriptionId"].(string)
						if stripeCustomerID == "" || stripeSubID == "" {
							user, _ := c.Logto.User(ctx, userID)
							email := ""
							if user != nil {
								email = user.PrimaryEmail
							}
							c.setupBilling(ctx, wsID, username, email, 14)
						}
					}
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
			_ = c.Logto.AddOrganizationMember(ctx, existing.ID, userID)
			_ = c.Logto.AssignOrganizationRoles(ctx, existing.ID, userID, []string{adminRoleID, memberRoleID})
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
	org, err := c.Logto.CreateOrganization(ctx, wsID, username, customData)
	if err != nil {
		return "", false, fmt.Errorf("failed to create personal workspace: %w", err)
	}

	// Cache the org ID mapping
	c.cacheOrgID(wsID, org.ID)

	// Add user as admin member
	if err := c.Logto.AddOrganizationMember(ctx, org.ID, userID); err != nil {
		return "", false, fmt.Errorf("failed to add user to personal workspace: %w", err)
	}
	if err := c.Logto.AssignOrganizationRoles(ctx, org.ID, userID, []string{adminRoleID, memberRoleID}); err != nil {
		return "", false, fmt.Errorf("failed to assign roles in personal workspace: %w", err)
	}

	slog.Info("personal workspace created", "id", wsID, "user", userID)

	// Best-effort: set up Stripe customer + subscription with 14-day trial
	user, _ := c.Logto.User(ctx, userID)
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
		_, err := c.Logto.OrganizationByName(ctx, candidate)
		if err != nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("all workspace ID slots exhausted for base %q", base)
}

// setupBilling ensures a Stripe customer and subscription exist for a workspace and
// stores their IDs in the Logto org customData. Idempotent: skips steps that are
// already complete, and self-heals on every login if a previous attempt partially failed.
//
// creditDays > 0 creates a promotional credit grant that expires after that many days.
// Personal workspaces get 14 days of credits; additional workspaces get 0 (no promo credits).
//
// Uses context.WithoutCancel to detach from the HTTP request lifecycle. Billing setup
// must complete even if the browser navigates away during the OIDC callback redirect.
func (c *Client) setupBilling(ctx context.Context, workspace, name, email string, creditDays int32) {
	if c.Cashier == nil {
		return
	}
	if c.Logto == nil {
		return
	}

	// Detach from the HTTP request context. The OIDC callback redirects the browser
	// immediately after this returns, which cancels r.Context(). Billing setup must
	// survive that cancellation.
	billingCtx := context.WithoutCancel(ctx)

	// Resolve workspace ID to Logto org ID for API calls
	orgID, err := c.resolveOrgID(billingCtx, workspace)
	if err != nil {
		slog.Warn("failed to resolve org ID for billing setup", "workspace", workspace, "error", err)
		return
	}

	// Read current org customData to check if billing is already (partially) set up.
	var customData map[string]interface{}
	org, err := c.Logto.Organization(billingCtx, orgID)
	if err != nil {
		slog.Warn("failed to read org for billing setup", "workspace", workspace, "error", err)
		customData = make(map[string]interface{})
	} else {
		customData = org.CustomData
		if customData == nil {
			customData = make(map[string]interface{})
		}
	}

	// Step 1: Ensure Stripe customer exists.
	customerID, _ := customData["stripeCustomerId"].(string)
	if customerID == "" {
		outCtx := auth.OutgoingContext(billingCtx)
		custCtx, custCancel := context.WithTimeout(outCtx, grpcTimeout)
		defer custCancel()
		custResp, custErr := c.Cashier.CreateCustomer(custCtx, &cashier.CreateCustomerRequest{
			Workspace: workspace,
			Name:      name,
			Email:     email,
		})
		if custErr != nil {
			slog.Warn("failed to create Stripe customer for workspace", "workspace", workspace, "error", custErr)
			return // Will retry on next login
		}
		customerID = custResp.CustomerId
	}

	// Step 2: Ensure Stripe subscription exists (metered items only, no plan).
	subscriptionID, _ := customData["stripeSubscriptionId"].(string)
	if subscriptionID == "" {
		outCtx := auth.OutgoingContext(billingCtx)
		subCtx, subCancel := context.WithTimeout(outCtx, grpcTimeout)
		defer subCancel()
		subResp, subErr := c.Cashier.CreateSubscription(subCtx, &cashier.CreateSubscriptionRequest{
			Workspace:  workspace,
			CustomerId: customerID,
			CreditDays: creditDays,
		})
		if subErr != nil {
			slog.Warn("failed to create Stripe subscription for workspace", "workspace", workspace, "error", subErr)
			// Store at least the customer ID so we don't re-create it next time.
			customData["stripeCustomerId"] = customerID
			_ = c.Logto.UpdateOrganizationCustomData(billingCtx, orgID, customData)
			return // Will retry subscription on next login
		}
		subscriptionID = subResp.SubscriptionId
	}

	// Step 3: Persist both IDs to the org customData.
	// Skip if both are already stored (nothing changed).
	existingCustID, _ := customData["stripeCustomerId"].(string)
	existingSubID, _ := customData["stripeSubscriptionId"].(string)
	if existingCustID == customerID && existingSubID == subscriptionID {
		slog.Debug("billing already set up", "workspace", workspace)
		return
	}

	customData["stripeCustomerId"] = customerID
	customData["stripeSubscriptionId"] = subscriptionID
	if err := c.Logto.UpdateOrganizationCustomData(billingCtx, orgID, customData); err != nil {
		slog.Warn("failed to store billing IDs in org customData", "workspace", workspace, "error", err)
		return // Will retry on next login
	}

	slog.Info("billing setup complete", "workspace", workspace, "customer_id", customerID, "subscription_id", subscriptionID)
}

// CreateWorkspaceCheckout creates a Stripe Checkout Session for a new workspace subscription.
// The workspace is not created until the checkout completes (see CompleteWorkspaceCheckout).
func (c *Client) CreateWorkspaceCheckout(ctx context.Context, id, name, plan string) (string, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return "", fmt.Errorf("unauthenticated")
	}
	if c.Logto == nil {
		return "", fmt.Errorf("logto not configured")
	}
	if c.Cashier == nil {
		return "", fmt.Errorf("billing not configured")
	}

	if !workspaceIDPattern.MatchString(id) {
		return "", fmt.Errorf("invalid workspace ID: must be 3-63 lowercase alphanumeric characters or hyphens")
	}

	// Check if workspace ID is already taken.
	_, err := c.Logto.OrganizationByName(ctx, id)
	if err == nil {
		return "", fmt.Errorf("workspace ID %q is already taken", id)
	}

	// Build Stripe Checkout URLs.
	successURL := fmt.Sprintf("%s/checkout/success?session_id={CHECKOUT_SESSION_ID}", c.DashboardURL)
	cancelURL := c.DashboardURL

	outCtx := auth.OutgoingContext(ctx)
	callCtx, cancel := context.WithTimeout(outCtx, grpcTimeout)
	defer cancel()

	planProto := stringToPlanProto(plan)
	resp, err := c.Cashier.CreateCheckoutSession(callCtx, &cashier.CreateCheckoutSessionRequest{
		Workspace:  id,
		Name:       name,
		Plan:       planProto,
		Email:      claims.Email,
		SuccessUrl: successURL,
		CancelUrl:  cancelURL,
		UserId:     claims.Subject,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	slog.Info("workspace checkout initiated", "workspace", id, "plan", plan, "user", claims.Email)
	return resp.Url, nil
}

// CompleteWorkspaceCheckout verifies a completed Stripe Checkout Session and creates the workspace.
// Called after the user is redirected back from Stripe.
func (c *Client) CompleteWorkspaceCheckout(ctx context.Context, sessionID string) (*Workspace, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}
	if c.Logto == nil {
		return nil, fmt.Errorf("logto not configured")
	}
	if c.Cashier == nil {
		return nil, fmt.Errorf("billing not configured")
	}

	// Retrieve the checkout session from Cashier/Stripe.
	outCtx := auth.OutgoingContext(ctx)
	callCtx, cancel := context.WithTimeout(outCtx, grpcTimeout)
	defer cancel()

	session, err := c.Cashier.RetrieveCheckoutSession(callCtx, &cashier.RetrieveCheckoutSessionRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve checkout session: %w", err)
	}

	if session.Status != "complete" {
		return nil, fmt.Errorf("checkout session not complete (status: %s)", session.Status)
	}

	// Verify the session belongs to the current user.
	if session.UserId != claims.Subject {
		return nil, fmt.Errorf("checkout session does not belong to current user")
	}

	wsID := session.Workspace
	wsName := session.Name

	// Idempotent: if workspace already exists with this user as member, return it.
	existing, existErr := c.Logto.OrganizationByName(ctx, wsID)
	if existErr == nil {
		c.cacheOrgID(wsID, existing.ID)
		members, memErr := c.Logto.OrganizationMembers(ctx, existing.ID)
		if memErr == nil {
			for _, m := range members {
				if m.ID == claims.Subject {
					slog.Info("workspace checkout completed (already exists)", "workspace", wsID)
					return &Workspace{
						ID:   wsID,
						Name: displayNameFromOrgData(existing),
					}, nil
				}
			}
		}
		return nil, fmt.Errorf("workspace ID %q is already taken", wsID)
	}

	adminRoleID, memberRoleID, err := c.orgRoleIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve org role IDs: %w", err)
	}

	// Create Logto organization with Stripe IDs in customData.
	customData := map[string]interface{}{
		"stripeCustomerId":     session.CustomerId,
		"stripeSubscriptionId": session.SubscriptionId,
	}
	org, err := c.Logto.CreateOrganization(ctx, wsID, wsName, customData)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	c.cacheOrgID(wsID, org.ID)

	// Add creator as member + assign admin and member roles.
	if err := c.Logto.AddOrganizationMember(ctx, org.ID, claims.Subject); err != nil {
		return nil, fmt.Errorf("failed to add creator to organization: %w", err)
	}
	if err := c.Logto.AssignOrganizationRoles(ctx, org.ID, claims.Subject, []string{adminRoleID, memberRoleID}); err != nil {
		return nil, fmt.Errorf("failed to assign admin role to creator: %w", err)
	}

	slog.Info("workspace created via checkout", "id", wsID, "name", wsName, "customer_id", session.CustomerId, "subscription_id", session.SubscriptionId)

	// Fetch user info for member list.
	user, _ := c.Logto.User(ctx, claims.Subject)
	memberName := ""
	memberEmail := claims.Email
	if user != nil {
		memberName = user.Name
		if user.PrimaryEmail != "" {
			memberEmail = user.PrimaryEmail
		}
	}

	return &Workspace{
		ID:   wsID,
		Name: wsName,
		Members: []WorkspaceMember{
			{ID: claims.Subject, Email: memberEmail, Name: memberName, Role: auth.WorkspaceRoleAdmin},
		},
	}, nil
}

// requireWorkspaceAdmin checks that the current user is an admin of the given workspace.
func (c *Client) requireWorkspaceAdmin(ctx context.Context, workspace string) error {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return fmt.Errorf("unauthenticated")
	}
	if claims.WorkspaceRoleIn(workspace) != auth.WorkspaceRoleAdmin {
		return fmt.Errorf("forbidden: workspace admin role required")
	}
	return nil
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
