package argocd

// Application represents an ArgoCD Application resource.
type Application struct {
	Metadata Metadata        `json:"metadata"`
	Spec     ApplicationSpec `json:"spec"`
	Status   *AppStatus      `json:"status,omitempty"`
}

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type ApplicationSpec struct {
	Source      Source      `json:"source"`
	Destination Destination `json:"destination"`
	Project     string      `json:"project"`
	SyncPolicy  *SyncPolicy `json:"syncPolicy,omitempty"`
}

type Source struct {
	RepoURL        string `json:"repoURL"`
	Path           string `json:"path"`
	TargetRevision string `json:"targetRevision"`
	Helm           *Helm  `json:"helm,omitempty"`
}

type Helm struct {
	ValueFiles []string `json:"valueFiles,omitempty"`
}

type Destination struct {
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

type SyncPolicy struct {
	Automated *Automated `json:"automated,omitempty"`
}

type Automated struct {
	Prune    bool `json:"prune"`
	SelfHeal bool `json:"selfHeal"`
}

// AppStatus holds ArgoCD Application status fields.
type AppStatus struct {
	Health HealthStatus `json:"health"`
	Sync   SyncStatus   `json:"sync"`
}

type HealthStatus struct {
	Status  string `json:"status"` // Healthy, Progressing, Degraded, Missing, Unknown
	Message string `json:"message,omitempty"`
}

type SyncStatus struct {
	Status string `json:"status"` // Synced, OutOfSync, Unknown
}

// SyncRequest is the body for POST /api/v1/applications/{name}/sync.
type SyncRequest struct {
	Prune bool `json:"prune"`
}

// Repository represents an ArgoCD repository connection.
type Repository struct {
	Repo     string `json:"repo"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Type     string `json:"type,omitempty"`
}
