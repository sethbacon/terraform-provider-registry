package client

// User represents a registry user.
type User struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	OIDCSub   *string `json:"oidc_sub,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CreateUserRequest is the payload for creating a user.
type CreateUserRequest struct {
	Email   string  `json:"email"`
	Name    string  `json:"name"`
	OIDCSub *string `json:"oidc_sub,omitempty"`
}

// UpdateUserRequest is the payload for updating a user.
type UpdateUserRequest struct {
	Email   string  `json:"email"`
	Name    string  `json:"name"`
	OIDCSub *string `json:"oidc_sub,omitempty"`
}

// Organization represents a registry organization.
type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CreateOrganizationRequest is the payload for creating an organization.
type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// UpdateOrganizationRequest is the payload for updating an organization.
type UpdateOrganizationRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// OrganizationMember represents a user's membership in an organization.
type OrganizationMember struct {
	OrganizationID           string   `json:"organization_id"`
	UserID                   string   `json:"user_id"`
	RoleTemplateID           *string  `json:"role_template_id,omitempty"`
	RoleTemplateName         *string  `json:"role_template_name,omitempty"`
	RoleTemplateDisplayName  *string  `json:"role_template_display_name,omitempty"`
	RoleTemplateScopes       []string `json:"role_template_scopes,omitempty"`
	UserName                 string   `json:"user_name"`
	UserEmail                string   `json:"user_email"`
	CreatedAt                string   `json:"created_at"`
}

// AddMemberRequest is the payload for adding a member to an organization.
type AddMemberRequest struct {
	UserID         string  `json:"user_id"`
	RoleTemplateID *string `json:"role_template_id,omitempty"`
}

// UpdateMemberRequest is the payload for updating a member's role.
type UpdateMemberRequest struct {
	RoleTemplateID *string `json:"role_template_id"`
}

// APIKey represents a registry API key (token never returned after creation).
type APIKey struct {
	ID                       string   `json:"id"`
	UserID                   *string  `json:"user_id,omitempty"`
	OrganizationID           string   `json:"organization_id"`
	Name                     string   `json:"name"`
	Description              *string  `json:"description,omitempty"`
	KeyPrefix                string   `json:"key_prefix"`
	Scopes                   []string `json:"scopes"`
	ExpiresAt                *string  `json:"expires_at,omitempty"`
	LastUsedAt               *string  `json:"last_used_at,omitempty"`
	CreatedAt                string   `json:"created_at"`
	UserName                 *string  `json:"user_name,omitempty"`
}

// CreateAPIKeyRequest is the payload for creating an API key.
type CreateAPIKeyRequest struct {
	OrganizationID string   `json:"organization_id"`
	Name           string   `json:"name"`
	Description    *string  `json:"description,omitempty"`
	Scopes         []string `json:"scopes"`
	ExpiresAt      *string  `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse includes the raw key value (only available on creation).
type CreateAPIKeyResponse struct {
	Key APIKey `json:"key"`
	// RawKey is only present on create; store securely in state.
	RawKey string `json:"raw_key,omitempty"`
}

// UpdateAPIKeyRequest is the payload for updating an API key.
type UpdateAPIKeyRequest struct {
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Scopes      []string `json:"scopes"`
	ExpiresAt   *string  `json:"expires_at,omitempty"`
}

// Module represents a registry module record.
type Module struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	Namespace      string  `json:"namespace"`
	Name           string  `json:"name"`
	System         string  `json:"system"`
	Description    *string `json:"description,omitempty"`
	Source         *string `json:"source,omitempty"`
	CreatedBy      *string `json:"created_by,omitempty"`
	CreatedByName  *string `json:"created_by_name,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// CreateModuleRequest is the payload for creating a module record.
type CreateModuleRequest struct {
	OrganizationID string  `json:"organization_id"`
	Namespace      string  `json:"namespace"`
	Name           string  `json:"name"`
	System         string  `json:"system"`
	Description    *string `json:"description,omitempty"`
	Source         *string `json:"source,omitempty"`
}

// UpdateModuleRequest is the payload for updating a module record.
type UpdateModuleRequest struct {
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`
}

// ProviderRecord represents a registry provider record.
type ProviderRecord struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	Namespace      string  `json:"namespace"`
	Type           string  `json:"type"`
	Description    *string `json:"description,omitempty"`
	Source         *string `json:"source,omitempty"`
	CreatedBy      *string `json:"created_by,omitempty"`
	CreatedByName  *string `json:"created_by_name,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// CreateProviderRecordRequest is the payload for creating a provider record.
type CreateProviderRecordRequest struct {
	OrganizationID string  `json:"organization_id"`
	Namespace      string  `json:"namespace"`
	Type           string  `json:"type"`
	Description    *string `json:"description,omitempty"`
	Source         *string `json:"source,omitempty"`
}

// UpdateProviderRecordRequest is the payload for updating a provider record.
type UpdateProviderRecordRequest struct {
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`
}

// SCMProvider represents an SCM (source control) integration.
type SCMProvider struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"` // github, gitlab, azure, bitbucket
	BaseURL     *string `json:"base_url,omitempty"`
	OAuthStatus *string `json:"oauth_status,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateSCMProviderRequest is the payload for creating an SCM provider.
type CreateSCMProviderRequest struct {
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	BaseURL *string `json:"base_url,omitempty"`
}

// UpdateSCMProviderRequest is the payload for updating an SCM provider.
type UpdateSCMProviderRequest struct {
	Name    string  `json:"name"`
	BaseURL *string `json:"base_url,omitempty"`
}

// ModuleSCMLink represents a link between a module and an SCM repository.
type ModuleSCMLink struct {
	ModuleID      string  `json:"module_id"`
	SCMProviderID string  `json:"scm_provider_id"`
	Owner         string  `json:"owner"`
	Repo          string  `json:"repo"`
	Branch        string  `json:"branch"`
	TagPattern    *string `json:"tag_pattern,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// CreateModuleSCMLinkRequest is the payload for creating a module SCM link.
type CreateModuleSCMLinkRequest struct {
	SCMProviderID string  `json:"scm_provider_id"`
	Owner         string  `json:"owner"`
	Repo          string  `json:"repo"`
	Branch        string  `json:"branch"`
	TagPattern    *string `json:"tag_pattern,omitempty"`
}

// UpdateModuleSCMLinkRequest is the payload for updating a module SCM link.
type UpdateModuleSCMLinkRequest struct {
	Owner      string  `json:"owner"`
	Repo       string  `json:"repo"`
	Branch     string  `json:"branch"`
	TagPattern *string `json:"tag_pattern,omitempty"`
}

// Mirror represents a provider mirror configuration.
type Mirror struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Description         *string  `json:"description,omitempty"`
	UpstreamRegistryURL string   `json:"upstream_registry_url"`
	OrganizationID      *string  `json:"organization_id,omitempty"`
	NamespaceFilter     []string `json:"namespace_filter,omitempty"`
	ProviderFilter      []string `json:"provider_filter,omitempty"`
	VersionFilter       *string  `json:"version_filter,omitempty"`
	PlatformFilter      []string `json:"platform_filter,omitempty"`
	Enabled             bool     `json:"enabled"`
	SyncIntervalHours   int      `json:"sync_interval_hours"`
	LastSyncAt          *string  `json:"last_sync_at,omitempty"`
	LastSyncStatus      *string  `json:"last_sync_status,omitempty"`
	LastSyncError       *string  `json:"last_sync_error,omitempty"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
	CreatedBy           *string  `json:"created_by,omitempty"`
}

// CreateMirrorRequest is the payload for creating a mirror.
type CreateMirrorRequest struct {
	Name                string   `json:"name"`
	Description         *string  `json:"description,omitempty"`
	UpstreamRegistryURL string   `json:"upstream_registry_url"`
	OrganizationID      *string  `json:"organization_id,omitempty"`
	NamespaceFilter     []string `json:"namespace_filter,omitempty"`
	ProviderFilter      []string `json:"provider_filter,omitempty"`
	VersionFilter       *string  `json:"version_filter,omitempty"`
	PlatformFilter      []string `json:"platform_filter,omitempty"`
	Enabled             bool     `json:"enabled"`
	SyncIntervalHours   int      `json:"sync_interval_hours"`
}

// UpdateMirrorRequest is the payload for updating a mirror.
type UpdateMirrorRequest struct {
	Name                string   `json:"name"`
	Description         *string  `json:"description,omitempty"`
	UpstreamRegistryURL string   `json:"upstream_registry_url"`
	OrganizationID      *string  `json:"organization_id,omitempty"`
	NamespaceFilter     []string `json:"namespace_filter,omitempty"`
	ProviderFilter      []string `json:"provider_filter,omitempty"`
	VersionFilter       *string  `json:"version_filter,omitempty"`
	PlatformFilter      []string `json:"platform_filter,omitempty"`
	Enabled             bool     `json:"enabled"`
	SyncIntervalHours   int      `json:"sync_interval_hours"`
}

// TerraformMirror represents a Terraform binary mirror configuration.
type TerraformMirror struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       *string  `json:"description,omitempty"`
	Tool              string   `json:"tool"` // terraform, opentofu, custom
	Enabled           bool     `json:"enabled"`
	UpstreamURL       string   `json:"upstream_url"`
	PlatformFilter    []string `json:"platform_filter,omitempty"`
	VersionFilter     *string  `json:"version_filter,omitempty"`
	GPGVerify         bool     `json:"gpg_verify"`
	StableOnly        bool     `json:"stable_only"`
	SyncIntervalHours int      `json:"sync_interval_hours"`
	LastSyncAt        *string  `json:"last_sync_at,omitempty"`
	LastSyncStatus    *string  `json:"last_sync_status,omitempty"`
	LastSyncError     *string  `json:"last_sync_error,omitempty"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// CreateTerraformMirrorRequest is the payload for creating a Terraform mirror.
type CreateTerraformMirrorRequest struct {
	Name              string   `json:"name"`
	Description       *string  `json:"description,omitempty"`
	Tool              string   `json:"tool"`
	Enabled           bool     `json:"enabled"`
	UpstreamURL       string   `json:"upstream_url"`
	PlatformFilter    []string `json:"platform_filter,omitempty"`
	VersionFilter     *string  `json:"version_filter,omitempty"`
	GPGVerify         bool     `json:"gpg_verify"`
	StableOnly        bool     `json:"stable_only"`
	SyncIntervalHours int      `json:"sync_interval_hours"`
}

// UpdateTerraformMirrorRequest is the payload for updating a Terraform mirror.
type UpdateTerraformMirrorRequest struct {
	Name              string   `json:"name"`
	Description       *string  `json:"description,omitempty"`
	Tool              string   `json:"tool"`
	Enabled           bool     `json:"enabled"`
	UpstreamURL       string   `json:"upstream_url"`
	PlatformFilter    []string `json:"platform_filter,omitempty"`
	VersionFilter     *string  `json:"version_filter,omitempty"`
	GPGVerify         bool     `json:"gpg_verify"`
	StableOnly        bool     `json:"stable_only"`
	SyncIntervalHours int      `json:"sync_interval_hours"`
}

// StorageConfig represents a storage backend configuration.
type StorageConfig struct {
	ID        string            `json:"id"`
	Backend   string            `json:"backend"` // local, s3, azure, gcs
	Config    map[string]string `json:"config"`  // sensitive
	Active    bool              `json:"active"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

// CreateStorageConfigRequest is the payload for creating a storage config.
type CreateStorageConfigRequest struct {
	Backend string            `json:"backend"`
	Config  map[string]string `json:"config"`
	Activate bool             `json:"activate,omitempty"`
}

// UpdateStorageConfigRequest is the payload for updating a storage config.
type UpdateStorageConfigRequest struct {
	Backend string            `json:"backend"`
	Config  map[string]string `json:"config"`
}

// RoleTemplate represents an RBAC role template.
type RoleTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description *string  `json:"description,omitempty"`
	Scopes      []string `json:"scopes"`
	IsSystem    bool     `json:"is_system"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// CreateRoleTemplateRequest is the payload for creating a role template.
type CreateRoleTemplateRequest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description *string  `json:"description,omitempty"`
	Scopes      []string `json:"scopes"`
}

// UpdateRoleTemplateRequest is the payload for updating a role template.
type UpdateRoleTemplateRequest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description *string  `json:"description,omitempty"`
	Scopes      []string `json:"scopes"`
}

// Policy represents a mirror approval policy.
type Policy struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Rules       string  `json:"rules"` // JSON-encoded rules
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreatePolicyRequest is the payload for creating a policy.
type CreatePolicyRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Rules       string  `json:"rules"`
}

// UpdatePolicyRequest is the payload for updating a policy.
type UpdatePolicyRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Rules       string  `json:"rules"`
}

// ApprovalRequest represents a mirror approval request.
type ApprovalRequest struct {
	ID            string  `json:"id"`
	MirrorID      string  `json:"mirror_id"`
	Justification string  `json:"justification"`
	ReviewStatus  string  `json:"review_status"` // pending, approved, rejected
	ReviewerID    *string `json:"reviewer_id,omitempty"`
	ReviewNote    *string `json:"review_note,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// CreateApprovalRequestRequest is the payload for creating an approval request.
type CreateApprovalRequestRequest struct {
	MirrorID      string `json:"mirror_id"`
	Justification string `json:"justification"`
}

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID             string                 `json:"id"`
	UserID         *string                `json:"user_id,omitempty"`
	UserEmail      *string                `json:"user_email,omitempty"`
	UserName       *string                `json:"user_name,omitempty"`
	OrganizationID *string                `json:"organization_id,omitempty"`
	Action         string                 `json:"action"`
	ResourceType   *string                `json:"resource_type,omitempty"`
	ResourceID     *string                `json:"resource_id,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	IPAddress      *string                `json:"ip_address,omitempty"`
	CreatedAt      string                 `json:"created_at"`
}

// Stats represents dashboard statistics.
type Stats struct {
	TotalModules   int `json:"total_modules"`
	TotalProviders int `json:"total_providers"`
	TotalUsers     int `json:"total_users"`
	TotalOrgs      int `json:"total_organizations"`
	TotalMirrors   int `json:"total_mirrors"`
	TotalAPIKeys   int `json:"total_api_keys"`
}
