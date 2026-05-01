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
	OrganizationID          string   `json:"organization_id"`
	UserID                  string   `json:"user_id"`
	RoleTemplateID          *string  `json:"role_template_id,omitempty"`
	RoleTemplateName        *string  `json:"role_template_name,omitempty"`
	RoleTemplateDisplayName *string  `json:"role_template_display_name,omitempty"`
	RoleTemplateScopes      []string `json:"role_template_scopes,omitempty"`
	UserName                string   `json:"user_name"`
	UserEmail               string   `json:"user_email"`
	CreatedAt               string   `json:"created_at"`
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
	ID             string   `json:"id"`
	UserID         *string  `json:"user_id,omitempty"`
	OrganizationID string   `json:"organization_id"`
	Name           string   `json:"name"`
	Description    *string  `json:"description,omitempty"`
	KeyPrefix      string   `json:"key_prefix"`
	Scopes         []string `json:"scopes"`
	ExpiresAt      *string  `json:"expires_at,omitempty"`
	LastUsedAt     *string  `json:"last_used_at,omitempty"`
	CreatedAt      string   `json:"created_at"`
	UserName       *string  `json:"user_name,omitempty"`
}

// CreateAPIKeyRequest is the payload for creating an API key.
type CreateAPIKeyRequest struct {
	OrganizationID string   `json:"organization_id"`
	Name           string   `json:"name"`
	Description    *string  `json:"description,omitempty"`
	Scopes         []string `json:"scopes"`
	ExpiresAt      *string  `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse is the flat response body returned by POST /api/v1/apikeys.
// The raw key value is only available at creation time.
type CreateAPIKeyResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	RawKey      string   `json:"key"`
	KeyPrefix   string   `json:"key_prefix"`
	Scopes      []string `json:"scopes"`
	ExpiresAt   *string  `json:"expires_at,omitempty"`
	CreatedAt   string   `json:"created_at"`
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
//
// Mirrors the backend scm.SCMProvider type. ClientID is exposed on the
// response (not the secret); WebhookSecret is never returned.
type SCMProvider struct {
	ID             string  `json:"id"`
	OrganizationID *string `json:"organization_id,omitempty"`
	Name           string  `json:"name"`
	ProviderType   string  `json:"provider_type"`
	BaseURL        *string `json:"base_url,omitempty"`
	TenantID       *string `json:"tenant_id,omitempty"`
	ClientID       string  `json:"client_id,omitempty"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// CreateSCMProviderRequest is the payload for creating an SCM provider.
type CreateSCMProviderRequest struct {
	OrganizationID *string `json:"organization_id,omitempty"`
	Name           string  `json:"name"`
	ProviderType   string  `json:"provider_type"`
	BaseURL        *string `json:"base_url,omitempty"`
	TenantID       *string `json:"tenant_id,omitempty"`
	ClientID       string  `json:"client_id,omitempty"`
	ClientSecret   string  `json:"client_secret,omitempty"`
	WebhookSecret  string  `json:"webhook_secret,omitempty"`
}

// UpdateSCMProviderRequest is the payload for updating an SCM provider.
//
// All fields are pointers; omitted fields leave the existing value
// unchanged on the backend.
type UpdateSCMProviderRequest struct {
	Name          *string `json:"name,omitempty"`
	BaseURL       *string `json:"base_url,omitempty"`
	TenantID      *string `json:"tenant_id,omitempty"`
	ClientID      *string `json:"client_id,omitempty"`
	ClientSecret  *string `json:"client_secret,omitempty"`
	WebhookSecret *string `json:"webhook_secret,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// ModuleSCMLink represents a link between a module and an SCM repository.
//
// JSON field naming on read intentionally differs from the create/update
// request bodies: the response carries module_path / auto_publish_enabled
// while the request bodies use repository_path / auto_publish_enabled.
// See backend/internal/scm/types.go (response) and
// backend/internal/api/modules/scm_linking.go (request).
type ModuleSCMLink struct {
	ID              string `json:"id"`
	ModuleID        string `json:"module_id"`
	SCMProviderID   string `json:"scm_provider_id"`
	RepositoryOwner string `json:"repository_owner"`
	RepositoryName  string `json:"repository_name"`
	DefaultBranch   string `json:"default_branch"`
	ModulePath      string `json:"module_path"`
	TagPattern      string `json:"tag_pattern"`
	AutoPublish     bool   `json:"auto_publish_enabled"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// CreateModuleSCMLinkRequest is the payload for creating a module SCM link.
//
// Note the request field is repository_path, not module_path; the backend
// renames it on persist. See LinkSCMRequest in
// backend/internal/api/modules/scm_linking.go.
type CreateModuleSCMLinkRequest struct {
	SCMProviderID   string `json:"provider_id"`
	RepositoryOwner string `json:"repository_owner"`
	RepositoryName  string `json:"repository_name"`
	DefaultBranch   string `json:"default_branch"`
	RepositoryPath  string `json:"repository_path,omitempty"`
	TagPattern      string `json:"tag_pattern,omitempty"`
	AutoPublish     bool   `json:"auto_publish_enabled"`
}

// UpdateModuleSCMLinkRequest is the payload for updating a module SCM link.
type UpdateModuleSCMLinkRequest struct {
	SCMProviderID   string `json:"provider_id"`
	RepositoryOwner string `json:"repository_owner"`
	RepositoryName  string `json:"repository_name"`
	DefaultBranch   string `json:"default_branch"`
	RepositoryPath  string `json:"repository_path,omitempty"`
	TagPattern      string `json:"tag_pattern,omitempty"`
	AutoPublish     bool   `json:"auto_publish_enabled"`
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
//
// Optional scalar fields use pointers so the wire payload omits them entirely
// when the caller has nothing to set, matching the backend's
// CreateMirrorConfigRequest semantics in
// backend/internal/db/models/mirror.go.
type CreateMirrorRequest struct {
	Name                string   `json:"name"`
	Description         *string  `json:"description,omitempty"`
	UpstreamRegistryURL string   `json:"upstream_registry_url"`
	OrganizationID      *string  `json:"organization_id,omitempty"`
	NamespaceFilter     []string `json:"namespace_filter,omitempty"`
	ProviderFilter      []string `json:"provider_filter,omitempty"`
	VersionFilter       *string  `json:"version_filter,omitempty"`
	PlatformFilter      []string `json:"platform_filter,omitempty"`
	Enabled             *bool    `json:"enabled,omitempty"`
	SyncIntervalHours   *int     `json:"sync_interval_hours,omitempty"`
}

// UpdateMirrorRequest is the payload for updating a mirror.
//
// All fields are pointers; only fields the caller explicitly populates are
// sent on the wire. This matches the backend's UpdateMirrorConfigRequest,
// where omitted fields leave the existing value unchanged. Sending plain
// bool/int zero values would silently disable the mirror or reset its sync
// interval to zero.
type UpdateMirrorRequest struct {
	Name                *string  `json:"name,omitempty"`
	Description         *string  `json:"description,omitempty"`
	UpstreamRegistryURL *string  `json:"upstream_registry_url,omitempty"`
	OrganizationID      *string  `json:"organization_id,omitempty"`
	NamespaceFilter     []string `json:"namespace_filter,omitempty"`
	ProviderFilter      []string `json:"provider_filter,omitempty"`
	VersionFilter       *string  `json:"version_filter,omitempty"`
	PlatformFilter      []string `json:"platform_filter,omitempty"`
	Enabled             *bool    `json:"enabled,omitempty"`
	SyncIntervalHours   *int     `json:"sync_interval_hours,omitempty"`
}

// TerraformMirror represents a Terraform binary mirror configuration.
type TerraformMirror struct {
	ID                string   `json:"id"`
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
	LastSyncAt        *string  `json:"last_sync_at,omitempty"`
	LastSyncStatus    *string  `json:"last_sync_status,omitempty"`
	LastSyncError     *string  `json:"last_sync_error,omitempty"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// CreateTerraformMirrorRequest is the payload for creating a Terraform mirror.
//
// Optional scalars use pointers so the wire payload omits them when the caller
// has nothing to set; the backend then applies defaults. See
// backend/internal/db/models/terraform_mirror.go.
type CreateTerraformMirrorRequest struct {
	Name              string   `json:"name"`
	Description       *string  `json:"description,omitempty"`
	Tool              string   `json:"tool"`
	Enabled           *bool    `json:"enabled,omitempty"`
	UpstreamURL       string   `json:"upstream_url"`
	PlatformFilter    []string `json:"platform_filter,omitempty"`
	VersionFilter     *string  `json:"version_filter,omitempty"`
	GPGVerify         *bool    `json:"gpg_verify,omitempty"`
	StableOnly        *bool    `json:"stable_only,omitempty"`
	SyncIntervalHours *int     `json:"sync_interval_hours,omitempty"`
}

// UpdateTerraformMirrorRequest is the payload for updating a Terraform mirror.
//
// All fields are pointers; omitted fields leave the existing value unchanged
// to match backend UpdateTerraformMirrorConfigRequest semantics. Sending plain
// bool/int zero values would silently disable the mirror or skip GPG
// verification.
type UpdateTerraformMirrorRequest struct {
	Name              *string  `json:"name,omitempty"`
	Description       *string  `json:"description,omitempty"`
	Tool              *string  `json:"tool,omitempty"`
	Enabled           *bool    `json:"enabled,omitempty"`
	UpstreamURL       *string  `json:"upstream_url,omitempty"`
	PlatformFilter    []string `json:"platform_filter,omitempty"`
	VersionFilter     *string  `json:"version_filter,omitempty"`
	GPGVerify         *bool    `json:"gpg_verify,omitempty"`
	StableOnly        *bool    `json:"stable_only,omitempty"`
	SyncIntervalHours *int     `json:"sync_interval_hours,omitempty"`
}

// StorageConfig represents a storage backend configuration.
type StorageConfig struct {
	ID          string `json:"id"`
	BackendType string `json:"backend_type"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	// Individual backend fields returned by API (credentials redacted)
	LocalBasePath      *string `json:"local_base_path,omitempty"`
	LocalServeDirectly *bool   `json:"local_serve_directly,omitempty"`
	AzureAccountName   *string `json:"azure_account_name,omitempty"`
	AzureContainerName *string `json:"azure_container_name,omitempty"`
	S3Region           *string `json:"s3_region,omitempty"`
	S3Bucket           *string `json:"s3_bucket,omitempty"`
	S3Endpoint         *string `json:"s3_endpoint,omitempty"`
	GCSBucket          *string `json:"gcs_bucket,omitempty"`
	GCSProjectID       *string `json:"gcs_project_id,omitempty"`
}

// CreateStorageConfigRequest is the payload for creating a storage config.
type CreateStorageConfigRequest struct {
	BackendType        string `json:"backend_type"`
	LocalBasePath      string `json:"local_base_path,omitempty"`
	LocalServeDirectly *bool  `json:"local_serve_directly,omitempty"`
	AzureAccountName   string `json:"azure_account_name,omitempty"`
	AzureAccountKey    string `json:"azure_account_key,omitempty"`
	AzureContainerName string `json:"azure_container_name,omitempty"`
	AzureCDNURL        string `json:"azure_cdn_url,omitempty"`
	S3Endpoint         string `json:"s3_endpoint,omitempty"`
	S3Region           string `json:"s3_region,omitempty"`
	S3Bucket           string `json:"s3_bucket,omitempty"`
	S3AuthMethod       string `json:"s3_auth_method,omitempty"`
	S3AccessKeyID      string `json:"s3_access_key_id,omitempty"`
	S3SecretAccessKey  string `json:"s3_secret_access_key,omitempty"`
	GCSBucket          string `json:"gcs_bucket,omitempty"`
	GCSProjectID       string `json:"gcs_project_id,omitempty"`
	GCSAuthMethod      string `json:"gcs_auth_method,omitempty"`
	GCSCredentialsFile string `json:"gcs_credentials_file,omitempty"`
	GCSCredentialsJSON string `json:"gcs_credentials_json,omitempty"`
	GCSEndpoint        string `json:"gcs_endpoint,omitempty"`
}

// UpdateStorageConfigRequest is the payload for updating a storage config.
type UpdateStorageConfigRequest struct {
	BackendType        string `json:"backend_type"`
	LocalBasePath      string `json:"local_base_path,omitempty"`
	LocalServeDirectly *bool  `json:"local_serve_directly,omitempty"`
	AzureAccountName   string `json:"azure_account_name,omitempty"`
	AzureAccountKey    string `json:"azure_account_key,omitempty"`
	AzureContainerName string `json:"azure_container_name,omitempty"`
	AzureCDNURL        string `json:"azure_cdn_url,omitempty"`
	S3Endpoint         string `json:"s3_endpoint,omitempty"`
	S3Region           string `json:"s3_region,omitempty"`
	S3Bucket           string `json:"s3_bucket,omitempty"`
	S3AuthMethod       string `json:"s3_auth_method,omitempty"`
	S3AccessKeyID      string `json:"s3_access_key_id,omitempty"`
	S3SecretAccessKey  string `json:"s3_secret_access_key,omitempty"`
	GCSBucket          string `json:"gcs_bucket,omitempty"`
	GCSProjectID       string `json:"gcs_project_id,omitempty"`
	GCSAuthMethod      string `json:"gcs_auth_method,omitempty"`
	GCSCredentialsFile string `json:"gcs_credentials_file,omitempty"`
	GCSCredentialsJSON string `json:"gcs_credentials_json,omitempty"`
	GCSEndpoint        string `json:"gcs_endpoint,omitempty"`
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
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	PolicyType       string  `json:"policy_type"`
	UpstreamRegistry *string `json:"upstream_registry,omitempty"`
	NamespacePattern *string `json:"namespace_pattern,omitempty"`
	ProviderPattern  *string `json:"provider_pattern,omitempty"`
	Priority         int     `json:"priority"`
	IsActive         bool    `json:"is_active"`
	RequiresApproval bool    `json:"requires_approval"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// CreatePolicyRequest is the payload for creating a policy.
type CreatePolicyRequest struct {
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	PolicyType       string  `json:"policy_type"`
	UpstreamRegistry *string `json:"upstream_registry,omitempty"`
	NamespacePattern *string `json:"namespace_pattern,omitempty"`
	ProviderPattern  *string `json:"provider_pattern,omitempty"`
	Priority         int     `json:"priority"`
	IsActive         bool    `json:"is_active"`
	RequiresApproval bool    `json:"requires_approval"`
}

// UpdatePolicyRequest is the payload for updating a policy.
type UpdatePolicyRequest struct {
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	PolicyType       string  `json:"policy_type"`
	UpstreamRegistry *string `json:"upstream_registry,omitempty"`
	NamespacePattern *string `json:"namespace_pattern,omitempty"`
	ProviderPattern  *string `json:"provider_pattern,omitempty"`
	Priority         int     `json:"priority"`
	IsActive         bool    `json:"is_active"`
	RequiresApproval bool    `json:"requires_approval"`
}

// ApprovalRequest represents a mirror approval request.
//
// Mirrors backend's MirrorApprovalRequest in
// backend/internal/db/models/mirror_approval.go. The OrganizationID,
// RequestedBy, ReviewedAt, ExpiresAt, RequestedByName, ReviewedByName,
// and MirrorName fields are populated by the backend on read; none of
// them are accepted on the create payload.
type ApprovalRequest struct {
	ID                string  `json:"id"`
	MirrorConfigID    string  `json:"mirror_config_id"`
	OrganizationID    *string `json:"organization_id,omitempty"`
	RequestedBy       *string `json:"requested_by,omitempty"`
	ProviderNamespace string  `json:"provider_namespace"`
	ProviderName      *string `json:"provider_name,omitempty"`
	Reason            string  `json:"reason,omitempty"`
	Status            string  `json:"status"`
	ReviewedBy        *string `json:"reviewed_by,omitempty"`
	ReviewedAt        *string `json:"reviewed_at,omitempty"`
	ReviewNotes       *string `json:"review_notes,omitempty"`
	AutoApproved      bool    `json:"auto_approved"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	ExpiresAt         *string `json:"expires_at,omitempty"`

	// Joined fields (populated server-side via LEFT JOIN; never written).
	RequestedByName string `json:"requested_by_name,omitempty"`
	ReviewedByName  string `json:"reviewed_by_name,omitempty"`
	MirrorName      string `json:"mirror_name,omitempty"`
}

// CreateApprovalRequestRequest is the payload for creating an approval request.
type CreateApprovalRequestRequest struct {
	MirrorConfigID    string  `json:"mirror_config_id"`
	ProviderNamespace string  `json:"provider_namespace"`
	ProviderName      *string `json:"provider_name,omitempty"`
	Reason            string  `json:"reason,omitempty"`
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

// Stats mirrors backend admin.DashboardStats. The shape was reworked
// alongside the v1 backend release: counters are now grouped under
// per-resource sub-objects with separate totals for versions, downloads,
// and (for providers) manual vs. mirrored counts. The previous flat
// counters (total_modules, total_users, ...) are no longer returned.
type Stats struct {
	Modules         ModuleStats         `json:"modules"`
	Providers       ProviderStats       `json:"providers"`
	ProviderMirrors ProviderMirrorStats `json:"provider_mirrors"`
	BinaryMirrors   BinaryMirrorStats   `json:"binary_mirrors"`
	Users           int                 `json:"users"`
	Organizations   int                 `json:"organizations"`
	SCMProviders    int                 `json:"scm_providers"`
	Downloads       int64               `json:"downloads"`
	RecentSyncs     []RecentSyncEntry   `json:"recent_syncs"`
}

// ModuleStats is the modules sub-object of dashboard stats.
type ModuleStats struct {
	Total     int                `json:"total"`
	Versions  int                `json:"versions"`
	Downloads int64              `json:"downloads"`
	BySystem  []ModuleSystemStat `json:"by_system,omitempty"`
}

// ModuleSystemStat is one entry in the modules.by_system breakdown.
type ModuleSystemStat struct {
	System string `json:"system"`
	Count  int    `json:"count"`
}

// ProviderStats is the providers sub-object of dashboard stats.
type ProviderStats struct {
	Total            int   `json:"total"`
	TotalVersions    int   `json:"total_versions"`
	Manual           int   `json:"manual"`
	ManualVersions   int   `json:"manual_versions"`
	Mirrored         int   `json:"mirrored"`
	MirroredVersions int   `json:"mirrored_versions"`
	Downloads        int64 `json:"downloads"`
}

// ProviderMirrorStats is the provider_mirrors sub-object.
type ProviderMirrorStats struct {
	Total   int `json:"total"`
	Healthy int `json:"healthy"`
	Failed  int `json:"failed"`
}

// BinaryMirrorStats is the binary_mirrors (Terraform/OpenTofu) sub-object.
type BinaryMirrorStats struct {
	Total     int                `json:"total"`
	Healthy   int                `json:"healthy"`
	Failed    int                `json:"failed"`
	Syncing   int                `json:"syncing"`
	Downloads int64              `json:"downloads"`
	Platforms int                `json:"platforms"`
	ByTool    []BinaryToolStat   `json:"by_tool,omitempty"`
}

// BinaryToolStat is one entry in binary_mirrors.by_tool.
type BinaryToolStat struct {
	Tool  string `json:"tool"`
	Count int    `json:"count"`
}

// RecentSyncEntry is one row of the recent_syncs ledger across all mirror
// types (provider mirrors + binary mirrors).
type RecentSyncEntry struct {
	MirrorName      string `json:"mirror_name"`
	MirrorType      string `json:"mirror_type"` // "binary" | "provider"
	Status          string `json:"status"`
	TriggeredBy     string `json:"triggered_by"`
	StartedAt       string `json:"started_at"`
	CompletedAt     string `json:"completed_at,omitempty"`
	VersionsSynced  int    `json:"versions_synced"`
	PlatformsSynced int    `json:"platforms_synced"`
}
