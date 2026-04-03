package context

// OSInfo holds operating system details.
type OSInfo struct {
	Name    string `json:"name"`
	Distro  string `json:"distro"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
	Kernel  string `json:"kernel"`
}

// PackageManagerInfo holds package manager state.
type PackageManagerInfo struct {
	Name       string `json:"name"`
	Available  bool   `json:"available"`
	NeedsUpdate bool  `json:"needs_update"`
	LastUpdate string `json:"last_update"`
}

// TargetState holds the current state of the operation target.
type TargetState struct {
	Installed  *bool   `json:"installed"`
	Version    *string `json:"version"`
	Running    *bool   `json:"running"`
	Enabled    *bool   `json:"enabled"`
	ConfigPath *string `json:"config_path"`
}

// UserInfo holds current user details.
type UserInfo struct {
	Name          string `json:"name"`
	UID           int    `json:"uid"`
	IsRoot        bool   `json:"is_root"`
	SudoAvailable bool   `json:"sudo_available"`
}

// ResourceInfo holds system resource availability.
type ResourceInfo struct {
	DiskFreeMB   int64 `json:"disk_free_mb"`
	MemoryFreeMB int64 `json:"memory_free_mb"`
}

// SystemContext is the complete system state passed from the Context Analyzer to the Planner.
type SystemContext struct {
	OS             OSInfo             `json:"os"`
	PackageManager PackageManagerInfo `json:"package_manager"`
	TargetState    TargetState        `json:"target_state"`
	DependenciesMet bool              `json:"dependencies_met"`
	User           UserInfo           `json:"user"`
	Resources      ResourceInfo       `json:"resources"`
}
