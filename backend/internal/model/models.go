package model

import "time"

type User struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	FullName     string    `json:"fullName" db:"full_name"`
	Password     string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	AuthProvider string    `json:"authProvider,omitempty" db:"auth_provider"`
	ExternalID   string    `json:"externalId,omitempty" db:"external_id"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

type Service struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Team        string    `json:"team" db:"team"`
	Type        string    `json:"type" db:"type"`
	Lifecycle   string    `json:"lifecycle" db:"lifecycle"`
	Description string    `json:"description" db:"description"`
	Language    string    `json:"language" db:"language"`
	Repository  string    `json:"repository" db:"repository"`
	Owner       string    `json:"owner" db:"owner"`
	Tags        []string  `json:"tags" db:"tags"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type Release struct {
	ID                string    `json:"id" db:"id"`
	ServiceID         string    `json:"serviceId" db:"service_id"`
	Version           string    `json:"version" db:"version"`
	Status            string    `json:"status" db:"status"`
	Description       string    `json:"description" db:"description"`
	Branch            string    `json:"branch" db:"branch"`
	GitCommit         string    `json:"gitCommit,omitempty" db:"git_commit"`
	CreatedBy         string    `json:"createdBy" db:"created_by"`
	Approvals         int       `json:"approvals" db:"approvals"`
	RequiredApprovals int       `json:"requiredApprovals" db:"required_approvals"`
	IsPrimary         bool      `json:"isPrimary" db:"is_primary"`
	TrafficPercentage int       `json:"trafficPercentage" db:"traffic_percentage"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time `json:"updatedAt" db:"updated_at"`
}

type ActiveVersion struct {
	ID                string    `json:"id" db:"id"`
	ServiceID         string    `json:"serviceId" db:"service_id"`
	ReleaseID         string    `json:"releaseId" db:"release_id"`
	Version           string    `json:"version" db:"version"`
	Branch            string    `json:"branch" db:"branch"`
	GitCommit         string    `json:"gitCommit" db:"git_commit"`
	Environment       string    `json:"environment" db:"environment"`
	TrafficPercentage int       `json:"trafficPercentage" db:"traffic_percentage"`
	Status            string    `json:"status" db:"status"`
	DeployedAt        time.Time `json:"deployedAt" db:"deployed_at"`
	DeployedBy        string    `json:"deployedBy" db:"deployed_by"`
}

type AuditEntry struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"userId" db:"user_id"`
	UserName   string    `json:"userName" db:"user_name"`
	Action     string    `json:"action" db:"action"`
	Resource   string    `json:"resource" db:"resource"`
	ResourceID string    `json:"resourceId" db:"resource_id"`
	Details    string    `json:"details" db:"details"`
	CreatedAt  time.Time `json:"timestamp" db:"created_at"`
}

type Integration struct {
	ID        string                 `json:"id" db:"id"`
	Type      string                 `json:"type" db:"type"`
	Name      string                 `json:"name" db:"name"`
	URL       string                 `json:"url" db:"url"`
	Enabled   bool                   `json:"enabled" db:"enabled"`
	Config    map[string]interface{} `json:"config" db:"config"`
	CreatedAt time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time              `json:"updatedAt" db:"updated_at"`
}

type CanaryDeployment struct {
	ID             string         `json:"id" db:"id"`
	ReleaseID      string         `json:"releaseId" db:"release_id"`
	ServiceID      string         `json:"serviceId" db:"service_id"`
	Status         string         `json:"status" db:"status"`
	Percentage     int            `json:"percentage" db:"percentage"`
	StableVersion  string         `json:"stableVersion" db:"stable_version"`
	CanaryVersion  string         `json:"canaryVersion" db:"canary_version"`
	StableRevision string         `json:"stableRevision" db:"stable_revision"`
	CanaryRevision string         `json:"canaryRevision" db:"canary_revision"`
	ArgoApp        string         `json:"argoApp" db:"argo_app"`
	GitTag         string         `json:"gitTag" db:"git_tag"`
	StartedAt      time.Time      `json:"startedAt" db:"started_at"`
	UpdatedAt      time.Time      `json:"updatedAt" db:"updated_at"`
	Metrics        *CanaryMetrics `json:"metrics,omitempty"`
}

type CanaryMetrics struct {
	ErrorRate     float64   `json:"errorRate"`
	LatencyP99    float64   `json:"latencyP99"`
	LatencyP95    float64   `json:"latencyP95"`
	RequestCount  int64     `json:"requestCount"`
	SuccessRate   float64   `json:"successRate"`
	LastCheckedAt time.Time `json:"lastCheckedAt"`
}

type VaultSecret struct {
	Path     string                 `json:"path"`
	Data     map[string]interface{} `json:"data"`
	Metadata *VaultSecretMetadata   `json:"metadata,omitempty"`
}

type VaultSecretMetadata struct {
	CreatedTime string `json:"created_time"`
	UpdatedTime string `json:"updated_time"`
	Version     int    `json:"version"`
}

type ProjectPermission struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"userId" db:"user_id"`
	ServiceID string    `json:"serviceId" db:"service_id"`
	Role      string    `json:"role" db:"role"`
	GrantedBy string    `json:"grantedBy" db:"granted_by"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"fullName" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
