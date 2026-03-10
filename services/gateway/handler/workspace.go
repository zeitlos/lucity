package handler

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/cashier"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/pkg/tenant"
)

// Workspace represents a workspace with metadata and members.
type Workspace struct {
	ID       string
	Name     string
	Personal bool
	Members  []WorkspaceMember
}

// WorkspaceMember represents a user's membership in a workspace.
type WorkspaceMember struct {
	ID    string
	Email string
	Name  string
	Role  auth.WorkspaceRole
}

var workspaceIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

// Workspaces returns all workspaces the current user is a member of.
// Reads workspace IDs from JWT claims, fetches metadata from deployer for each.
func (c *Client) Workspaces(ctx context.Context) ([]Workspace, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}

	outCtx := auth.OutgoingContext(ctx)

	var mu sync.Mutex
	var wg sync.WaitGroup
	workspaces := make([]Workspace, 0, len(claims.Workspaces))

	for _, m := range claims.Workspaces {
		wg.Add(1)
		go func(membership auth.WorkspaceMembership) {
			defer wg.Done()
			callCtx, cancel := context.WithTimeout(outCtx, grpcTimeout)
			defer cancel()

			resp, err := c.Deployer.WorkspaceMetadata(callCtx, &deployer.WorkspaceMetadataRequest{
				Workspace: membership.Workspace,
			})
			if err != nil {
				slog.Warn("failed to get workspace metadata", "workspace", membership.Workspace, "error", err)
				// Still include the workspace with minimal info
				mu.Lock()
				workspaces = append(workspaces, Workspace{
					ID:   membership.Workspace,
					Name: membership.Workspace,
				})
				mu.Unlock()
				return
			}

			mu.Lock()
			workspaces = append(workspaces, Workspace{
				ID:       membership.Workspace,
				Name:     resp.Name,
				Personal: resp.Personal,
			})
			mu.Unlock()
		}(m)
	}
	wg.Wait()

	return workspaces, nil
}

// Workspace returns metadata and members for the active workspace.
func (c *Client) Workspace(ctx context.Context) (*Workspace, error) {
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}
	outCtx := auth.OutgoingContext(ctx)

	// Fetch metadata
	callCtx, cancel := context.WithTimeout(outCtx, grpcTimeout)
	defer cancel()
	resp, err := c.Deployer.WorkspaceMetadata(callCtx, &deployer.WorkspaceMetadataRequest{
		Workspace: ws,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace metadata: %w", err)
	}

	result := &Workspace{
		ID:       ws,
		Name:     resp.Name,
		Personal: resp.Personal,
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

// WorkspaceMembers returns all members of the active workspace by querying Rauthy groups.
func (c *Client) WorkspaceMembers(ctx context.Context) ([]WorkspaceMember, error) {
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}
	if c.Rauthy == nil {
		return nil, fmt.Errorf("rauthy not configured")
	}

	// Find the workspace groups in Rauthy
	memberGroupName := "ws:" + ws
	adminGroupName := "ws:" + ws + ":admin"

	memberGroup, err := c.Rauthy.GroupByName(ctx, memberGroupName)
	if err != nil {
		return nil, fmt.Errorf("failed to find member group: %w", err)
	}
	adminGroup, err := c.Rauthy.GroupByName(ctx, adminGroupName)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin group: %w", err)
	}

	if memberGroup == nil {
		return nil, nil
	}

	// Build a set of admin user IDs
	adminSet := make(map[string]bool)
	if adminGroup != nil {
		admins, err := c.Rauthy.UsersByGroupID(ctx, adminGroup.ID)
		if err != nil {
			slog.Warn("failed to list admin group members", "workspace", ws, "error", err)
		}
		for _, u := range admins {
			adminSet[u.ID] = true
		}
	}

	// List all members
	users, err := c.Rauthy.UsersByGroupID(ctx, memberGroup.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspace members: %w", err)
	}

	members := make([]WorkspaceMember, 0, len(users))
	for _, u := range users {
		role := auth.WorkspaceRoleUser
		if adminSet[u.ID] {
			role = auth.WorkspaceRoleAdmin
		}
		members = append(members, WorkspaceMember{
			ID:    u.ID,
			Email: u.Email,
			Name:  u.Name(),
			Role:  role,
		})
	}

	// Also include admin-only users (in admin group but not member group)
	if adminGroup != nil {
		memberSet := make(map[string]bool, len(users))
		for _, u := range users {
			memberSet[u.ID] = true
		}
		admins, _ := c.Rauthy.UsersByGroupID(ctx, adminGroup.ID)
		for _, u := range admins {
			if !memberSet[u.ID] {
				members = append(members, WorkspaceMember{
					ID:    u.ID,
					Email: u.Email,
					Name:  u.Name(),
					Role:  auth.WorkspaceRoleAdmin,
				})
			}
		}
	}

	return members, nil
}

// CreateWorkspace creates a new workspace: Rauthy groups + K8s ConfigMap.
// The creator is automatically added as admin.
func (c *Client) CreateWorkspace(ctx context.Context, id, name string) (*Workspace, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}
	if c.Rauthy == nil {
		return nil, fmt.Errorf("rauthy not configured")
	}

	if !workspaceIDPattern.MatchString(id) {
		return nil, fmt.Errorf("invalid workspace ID: must be 3-63 lowercase alphanumeric characters or hyphens")
	}

	// Check if workspace ID is already taken.
	outCtx := auth.OutgoingContext(ctx)
	checkCtx, checkCancel := context.WithTimeout(outCtx, grpcTimeout)
	defer checkCancel()
	_, err := c.Deployer.WorkspaceMetadata(checkCtx, &deployer.WorkspaceMetadataRequest{Workspace: id})
	if err == nil {
		return nil, fmt.Errorf("workspace ID %q is already taken", id)
	}

	// 1. Create Rauthy groups
	memberGroup, err := c.Rauthy.CreateGroup(ctx, "ws:"+id)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace member group: %w", err)
	}

	adminGroup, err := c.Rauthy.CreateGroup(ctx, "ws:"+id+":admin")
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace admin group: %w", err)
	}

	// 2. Add creator to both groups
	user, err := c.Rauthy.User(ctx, claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current user from rauthy: %w", err)
	}

	newGroups := append(user.Groups, memberGroup.Name, adminGroup.Name)
	if err := c.Rauthy.UpdateUserGroups(ctx, user.ID, newGroups); err != nil {
		return nil, fmt.Errorf("failed to add creator to workspace groups: %w", err)
	}

	// 3. Create K8s ConfigMap via deployer
	callCtx, cancel := context.WithTimeout(outCtx, grpcTimeout)
	defer cancel()
	_, err = c.Deployer.CreateWorkspaceMetadata(callCtx, &deployer.CreateWorkspaceMetadataRequest{
		Workspace: id,
		Name:      name,
		Owner:     claims.Subject,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace metadata: %w", err)
	}

	slog.Info("workspace created", "id", id, "name", name, "creator", claims.Email)

	// Best-effort: set up Stripe customer + subscription
	c.setupBilling(ctx, id, name, claims.Email)

	return &Workspace{
		ID:   id,
		Name: name,
		Members: []WorkspaceMember{
			{ID: user.ID, Email: user.Email, Name: user.Name(), Role: auth.WorkspaceRoleAdmin},
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

	outCtx := auth.OutgoingContext(ctx)
	callCtx, cancel := context.WithTimeout(outCtx, grpcTimeout)
	defer cancel()
	_, err = c.Deployer.UpdateWorkspaceMetadata(callCtx, &deployer.UpdateWorkspaceMetadataRequest{
		Workspace: ws,
		Name:      name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
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

	// Check if workspace is personal
	outCtx := auth.OutgoingContext(ctx)
	metaCtx, metaCancel := context.WithTimeout(outCtx, grpcTimeout)
	defer metaCancel()
	meta, err := c.Deployer.WorkspaceMetadata(metaCtx, &deployer.WorkspaceMetadataRequest{Workspace: ws})
	if err != nil {
		return false, fmt.Errorf("failed to get workspace metadata: %w", err)
	}
	if meta.Personal {
		return false, fmt.Errorf("cannot delete personal workspace")
	}

	// Check no projects exist
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

	// Delete Rauthy groups (this also removes users from the groups)
	if c.Rauthy != nil {
		groups, err := c.Rauthy.FindGroupsByPrefix(ctx, "ws:"+ws)
		if err != nil {
			slog.Warn("failed to list workspace groups for deletion", "workspace", ws, "error", err)
		} else {
			for _, g := range groups {
				if err := c.Rauthy.DeleteGroup(ctx, g.ID); err != nil {
					slog.Warn("failed to delete workspace group", "group", g.Name, "error", err)
				}
			}
		}
	}

	// Delete K8s ConfigMap
	delCtx, delCancel := context.WithTimeout(outCtx, grpcTimeout)
	defer delCancel()
	_, err = c.Deployer.DeleteWorkspaceMetadata(delCtx, &deployer.DeleteWorkspaceMetadataRequest{Workspace: ws})
	if err != nil {
		return false, fmt.Errorf("failed to delete workspace metadata: %w", err)
	}

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
	if c.Rauthy == nil {
		return nil, fmt.Errorf("rauthy not configured")
	}

	// Find user by email
	user, err := c.Rauthy.UserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user with email %q not found in identity provider", email)
	}

	// Find workspace groups
	memberGroup, err := c.Rauthy.GroupByName(ctx, "ws:"+ws)
	if err != nil {
		return nil, fmt.Errorf("failed to find member group: %w", err)
	}
	if memberGroup == nil {
		return nil, fmt.Errorf("workspace member group not found")
	}

	// Add to member group
	newGroups := appendUnique(user.Groups, memberGroup.Name)

	// Add to admin group if admin role
	if role == auth.WorkspaceRoleAdmin {
		adminGroup, err := c.Rauthy.GroupByName(ctx, "ws:"+ws+":admin")
		if err != nil {
			return nil, fmt.Errorf("failed to find admin group: %w", err)
		}
		if adminGroup != nil {
			newGroups = appendUnique(newGroups, adminGroup.Name)
		}
	}

	if err := c.Rauthy.UpdateUserGroups(ctx, user.ID, newGroups); err != nil {
		return nil, fmt.Errorf("failed to add user to workspace groups: %w", err)
	}

	slog.Info("member invited", "workspace", ws, "email", email, "role", role)

	return &WorkspaceMember{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name(),
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
	if c.Rauthy == nil {
		return false, fmt.Errorf("rauthy not configured")
	}

	// Prevent removing yourself
	claims := auth.FromContext(ctx)
	if claims != nil && claims.Subject == userID {
		return false, fmt.Errorf("cannot remove yourself from workspace")
	}

	// Find workspace groups
	wsGroups, err := c.Rauthy.FindGroupsByPrefix(ctx, "ws:"+ws)
	if err != nil {
		return false, fmt.Errorf("failed to find workspace groups: %w", err)
	}

	wsGroupNames := make(map[string]bool, len(wsGroups))
	for _, g := range wsGroups {
		wsGroupNames[g.Name] = true
	}

	// Get user and remove workspace groups
	user, err := c.Rauthy.User(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	var remaining []string
	for _, name := range user.Groups {
		if !wsGroupNames[name] {
			remaining = append(remaining, name)
		}
	}

	if err := c.Rauthy.UpdateUserGroups(ctx, userID, remaining); err != nil {
		return false, fmt.Errorf("failed to remove user from workspace groups: %w", err)
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
	if c.Rauthy == nil {
		return nil, fmt.Errorf("rauthy not configured")
	}

	adminGroup, err := c.Rauthy.GroupByName(ctx, "ws:"+ws+":admin")
	if err != nil {
		return nil, fmt.Errorf("failed to find admin group: %w", err)
	}
	if adminGroup == nil {
		return nil, fmt.Errorf("workspace admin group not found")
	}

	user, err := c.Rauthy.User(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var newGroups []string
	if role == auth.WorkspaceRoleAdmin {
		newGroups = appendUnique(user.Groups, adminGroup.Name)
	} else {
		newGroups = removeFromSlice(user.Groups, adminGroup.Name)
	}

	if err := c.Rauthy.UpdateUserGroups(ctx, user.ID, newGroups); err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	return &WorkspaceMember{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name(),
		Role:  role,
	}, nil
}

// EnsurePersonalWorkspace creates a personal workspace for a new user if they have none.
// The workspace ID is derived from the user's GitHub login. Idempotent: if the workspace
// already exists and belongs to this user, re-adds them to the Rauthy groups and returns
// the existing ID. On genuine collision (different owner), picks {id}-0, {id}-1, etc.
func (c *Client) EnsurePersonalWorkspace(ctx context.Context, rauthyUserID, githubLogin string) (string, error) {
	if c.Rauthy == nil {
		return "", fmt.Errorf("rauthy not configured")
	}

	wsID := sanitizeWorkspaceID(githubLogin)
	if wsID == "" {
		return "", fmt.Errorf("cannot derive workspace ID from login %q", githubLogin)
	}

	outCtx := auth.OutgoingContext(ctx)

	// Check if preferred workspace ID already exists.
	checkCtx, checkCancel := context.WithTimeout(outCtx, grpcTimeout)
	defer checkCancel()
	existing, err := c.Deployer.WorkspaceMetadata(checkCtx, &deployer.WorkspaceMetadataRequest{Workspace: wsID})
	if err == nil {
		// Workspace exists. Check if it belongs to this user.
		// Pre-migration workspaces have no owner — treat personal ones as ours.
		if existing.Owner == rauthyUserID || (existing.Owner == "" && existing.Personal) {
			if err := c.ensureWorkspaceGroups(ctx, wsID, rauthyUserID); err != nil {
				return "", fmt.Errorf("failed to restore workspace groups: %w", err)
			}
			slog.Info("personal workspace restored", "id", wsID, "user", rauthyUserID)
			return wsID, nil
		}

		// Genuine collision — someone else owns this ID.
		wsID, err = c.findAvailableWorkspaceID(outCtx, wsID)
		if err != nil {
			return "", fmt.Errorf("failed to find available workspace ID: %w", err)
		}
	}

	// Create workspace groups and add user.
	if err := c.ensureWorkspaceGroups(ctx, wsID, rauthyUserID); err != nil {
		return "", fmt.Errorf("failed to create workspace groups: %w", err)
	}

	// Create ConfigMap.
	createCtx, createCancel := context.WithTimeout(outCtx, grpcTimeout)
	defer createCancel()
	_, err = c.Deployer.CreateWorkspaceMetadata(createCtx, &deployer.CreateWorkspaceMetadataRequest{
		Workspace: wsID,
		Name:      githubLogin,
		Personal:  true,
		Owner:     rauthyUserID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create personal workspace metadata: %w", err)
	}

	slog.Info("personal workspace created", "id", wsID, "user", rauthyUserID)

	// Best-effort: set up Stripe customer + subscription
	user, _ := c.Rauthy.User(ctx, rauthyUserID)
	email := ""
	if user != nil {
		email = user.Email
	}
	c.setupBilling(ctx, wsID, githubLogin, email)

	return wsID, nil
}

// ensureWorkspaceGroups creates Rauthy groups for a workspace and adds the user to them.
// Idempotent: CreateGroup returns existing groups, appendUnique avoids duplicate membership.
func (c *Client) ensureWorkspaceGroups(ctx context.Context, wsID, rauthyUserID string) error {
	memberGroup, err := c.Rauthy.CreateGroup(ctx, "ws:"+wsID)
	if err != nil {
		return fmt.Errorf("failed to create member group: %w", err)
	}

	adminGroup, err := c.Rauthy.CreateGroup(ctx, "ws:"+wsID+":admin")
	if err != nil {
		return fmt.Errorf("failed to create admin group: %w", err)
	}

	user, err := c.Rauthy.User(ctx, rauthyUserID)
	if err != nil {
		return fmt.Errorf("failed to fetch user from rauthy: %w", err)
	}

	newGroups := appendUnique(user.Groups, memberGroup.Name)
	newGroups = appendUnique(newGroups, adminGroup.Name)
	if err := c.Rauthy.UpdateUserGroups(ctx, user.ID, newGroups); err != nil {
		return fmt.Errorf("failed to add user to workspace groups: %w", err)
	}
	return nil
}

// findAvailableWorkspaceID tries {base}-0, {base}-1, ... up to {base}-9
// to find an available workspace ID. Returns error if all slots are taken.
func (c *Client) findAvailableWorkspaceID(ctx context.Context, base string) (string, error) {
	for i := 0; i < 10; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		checkCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
		_, err := c.Deployer.WorkspaceMetadata(checkCtx, &deployer.WorkspaceMetadataRequest{Workspace: candidate})
		cancel()
		if err != nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("all workspace ID slots exhausted for base %q", base)
}

// setupBilling creates a Stripe customer and subscription for a new workspace.
// Best-effort — logs warnings on failure but never returns errors.
func (c *Client) setupBilling(ctx context.Context, workspace, name, email string) {
	if c.Cashier == nil {
		return
	}

	outCtx := auth.OutgoingContext(ctx)

	custCtx, custCancel := context.WithTimeout(outCtx, grpcTimeout)
	defer custCancel()
	custResp, err := c.Cashier.CreateCustomer(custCtx, &cashier.CreateCustomerRequest{
		Workspace: workspace,
		Name:      name,
		Email:     email,
	})
	if err != nil {
		slog.Warn("failed to create Stripe customer for workspace", "workspace", workspace, "error", err)
		return
	}

	subCtx, subCancel := context.WithTimeout(outCtx, grpcTimeout)
	defer subCancel()
	_, err = c.Cashier.CreateSubscription(subCtx, &cashier.CreateSubscriptionRequest{
		Workspace:  workspace,
		CustomerId: custResp.CustomerId,
		Plan:       cashier.Plan_PLAN_HOBBY,
	})
	if err != nil {
		slog.Warn("failed to create Stripe subscription for workspace", "workspace", workspace, "error", err)
		return
	}

	slog.Info("billing setup complete", "workspace", workspace)
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

// sanitizeWorkspaceID converts a GitHub login to a valid workspace ID.
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

// appendUnique appends value to slice if not already present.
func appendUnique(slice []string, value string) []string {
	for _, v := range slice {
		if v == value {
			return slice
		}
	}
	return append(slice, value)
}

// removeFromSlice removes value from slice.
func removeFromSlice(slice []string, value string) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}
