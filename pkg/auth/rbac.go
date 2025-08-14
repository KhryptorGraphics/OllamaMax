package auth

import (
	"errors"
	"fmt"
	"sync"
)

// RBAC implements Role-Based Access Control
type RBAC struct {
	roles       map[string]*Role
	users       map[string]*User
	permissions map[string]*Permission
	mutex       sync.RWMutex
}

// Role represents a role in the RBAC system
type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	Inherits    []string `json:"inherits,omitempty"`
}

// User represents a user in the RBAC system
type User struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email,omitempty"`
	Roles       []string          `json:"roles"`
	Permissions []string          `json:"permissions,omitempty"` // Direct permissions
	Metadata    map[string]string `json:"metadata,omitempty"`
	Active      bool              `json:"active"`
}

// Permission represents a permission in the RBAC system
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// NewRBAC creates a new RBAC instance with default roles and permissions
func NewRBAC() *RBAC {
	rbac := &RBAC{
		roles:       make(map[string]*Role),
		users:       make(map[string]*User),
		permissions: make(map[string]*Permission),
	}

	// Initialize default permissions
	rbac.initializeDefaultPermissions()
	
	// Initialize default roles
	rbac.initializeDefaultRoles()

	return rbac
}

// initializeDefaultPermissions sets up the default permission set
func (r *RBAC) initializeDefaultPermissions() {
	defaultPermissions := []*Permission{
		{Name: PermissionModelManage, Description: "Manage models", Resource: "model", Action: "manage"},
		{Name: PermissionModelRead, Description: "Read model information", Resource: "model", Action: "read"},
		{Name: PermissionClusterManage, Description: "Manage cluster", Resource: "cluster", Action: "manage"},
		{Name: PermissionClusterRead, Description: "Read cluster information", Resource: "cluster", Action: "read"},
		{Name: PermissionNodeManage, Description: "Manage nodes", Resource: "node", Action: "manage"},
		{Name: PermissionNodeRead, Description: "Read node information", Resource: "node", Action: "read"},
		{Name: PermissionInferenceRun, Description: "Run inference", Resource: "inference", Action: "run"},
		{Name: PermissionMetricsRead, Description: "Read metrics", Resource: "metrics", Action: "read"},
		{Name: PermissionSystemManage, Description: "Manage system", Resource: "system", Action: "manage"},
	}

	for _, perm := range defaultPermissions {
		r.permissions[perm.Name] = perm
	}
}

// initializeDefaultRoles sets up the default role hierarchy
func (r *RBAC) initializeDefaultRoles() {
	defaultRoles := []*Role{
		{
			Name:        RoleAdmin,
			Description: "Full system administrator",
			Permissions: GetRolePermissions(RoleAdmin),
		},
		{
			Name:        RoleOperator,
			Description: "System operator with limited management access",
			Permissions: GetRolePermissions(RoleOperator),
		},
		{
			Name:        RoleUser,
			Description: "Regular user with inference access",
			Permissions: GetRolePermissions(RoleUser),
		},
		{
			Name:        RoleReadonly,
			Description: "Read-only access to system information",
			Permissions: GetRolePermissions(RoleReadonly),
		},
	}

	for _, role := range defaultRoles {
		r.roles[role.Name] = role
	}
}

// CreateRole creates a new role
func (r *RBAC) CreateRole(role *Role) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.roles[role.Name]; exists {
		return fmt.Errorf("role %s already exists", role.Name)
	}

	// Validate permissions exist
	for _, perm := range role.Permissions {
		if _, exists := r.permissions[perm]; !exists {
			return fmt.Errorf("permission %s does not exist", perm)
		}
	}

	r.roles[role.Name] = role
	return nil
}

// GetRole retrieves a role by name
func (r *RBAC) GetRole(name string) (*Role, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	role, exists := r.roles[name]
	if !exists {
		return nil, fmt.Errorf("role %s not found", name)
	}

	return role, nil
}

// CreateUser creates a new user
func (r *RBAC) CreateUser(user *User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return fmt.Errorf("user %s already exists", user.ID)
	}

	// Validate roles exist
	for _, roleName := range user.Roles {
		if _, exists := r.roles[roleName]; !exists {
			return fmt.Errorf("role %s does not exist", roleName)
		}
	}

	// Validate direct permissions exist
	for _, perm := range user.Permissions {
		if _, exists := r.permissions[perm]; !exists {
			return fmt.Errorf("permission %s does not exist", perm)
		}
	}

	r.users[user.ID] = user
	return nil
}

// GetUser retrieves a user by ID
func (r *RBAC) GetUser(userID string) (*User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (r *RBAC) GetUserByUsername(username string) (*User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user with username %s not found", username)
}

// HasPermission checks if a user has a specific permission
func (r *RBAC) HasPermission(userID, permission string) (bool, error) {
	user, err := r.GetUser(userID)
	if err != nil {
		return false, err
	}

	if !user.Active {
		return false, errors.New("user is not active")
	}

	// Check direct permissions
	for _, perm := range user.Permissions {
		if perm == permission {
			return true, nil
		}
	}

	// Check role-based permissions
	for _, roleName := range user.Roles {
		role, exists := r.roles[roleName]
		if !exists {
			continue
		}

		for _, perm := range role.Permissions {
			if perm == permission {
				return true, nil
			}
		}

		// Check inherited permissions
		if r.hasInheritedPermission(role, permission) {
			return true, nil
		}
	}

	return false, nil
}

// hasInheritedPermission checks if a role has a permission through inheritance
func (r *RBAC) hasInheritedPermission(role *Role, permission string) bool {
	for _, inheritedRoleName := range role.Inherits {
		inheritedRole, exists := r.roles[inheritedRoleName]
		if !exists {
			continue
		}

		for _, perm := range inheritedRole.Permissions {
			if perm == permission {
				return true
			}
		}

		// Recursive check for nested inheritance
		if r.hasInheritedPermission(inheritedRole, permission) {
			return true
		}
	}

	return false
}

// GetUserPermissions returns all permissions for a user (direct + role-based)
func (r *RBAC) GetUserPermissions(userID string) ([]string, error) {
	user, err := r.GetUser(userID)
	if err != nil {
		return nil, err
	}

	permissionSet := make(map[string]bool)

	// Add direct permissions
	for _, perm := range user.Permissions {
		permissionSet[perm] = true
	}

	// Add role-based permissions
	for _, roleName := range user.Roles {
		role, exists := r.roles[roleName]
		if !exists {
			continue
		}

		for _, perm := range role.Permissions {
			permissionSet[perm] = true
		}

		// Add inherited permissions
		r.addInheritedPermissions(role, permissionSet)
	}

	// Convert to slice
	permissions := make([]string, 0, len(permissionSet))
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// addInheritedPermissions recursively adds inherited permissions to the set
func (r *RBAC) addInheritedPermissions(role *Role, permissionSet map[string]bool) {
	for _, inheritedRoleName := range role.Inherits {
		inheritedRole, exists := r.roles[inheritedRoleName]
		if !exists {
			continue
		}

		for _, perm := range inheritedRole.Permissions {
			permissionSet[perm] = true
		}

		// Recursive call for nested inheritance
		r.addInheritedPermissions(inheritedRole, permissionSet)
	}
}

// AssignRole assigns a role to a user
func (r *RBAC) AssignRole(userID, roleName string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	if _, exists := r.roles[roleName]; !exists {
		return fmt.Errorf("role %s not found", roleName)
	}

	// Check if user already has the role
	for _, existingRole := range user.Roles {
		if existingRole == roleName {
			return nil // Already has the role
		}
	}

	user.Roles = append(user.Roles, roleName)
	return nil
}

// RevokeRole removes a role from a user
func (r *RBAC) RevokeRole(userID, roleName string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	for i, role := range user.Roles {
		if role == roleName {
			user.Roles = append(user.Roles[:i], user.Roles[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("user %s does not have role %s", userID, roleName)
}

// ListRoles returns all available roles
func (r *RBAC) ListRoles() []*Role {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	roles := make([]*Role, 0, len(r.roles))
	for _, role := range r.roles {
		roles = append(roles, role)
	}

	return roles
}

// ListUsers returns all users
func (r *RBAC) ListUsers() []*User {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	users := make([]*User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users
}

// ListPermissions returns all available permissions
func (r *RBAC) ListPermissions() []*Permission {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	permissions := make([]*Permission, 0, len(r.permissions))
	for _, perm := range r.permissions {
		permissions = append(permissions, perm)
	}

	return permissions
}
