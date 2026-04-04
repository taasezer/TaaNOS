package planner

import (
	"fmt"
	"time"
)

// ActionKey uniquely identifies an action in the registry.
type ActionKey struct {
	Category   string
	Action     string
	PkgManager string // empty for non-package operations
}

// ActionDef defines a registered action template.
type ActionDef struct {
	Key            ActionKey
	Description    string
	CommandTemplate string // Go template with {target}, {options}
	RequiresRoot   bool
	CanFail        bool
	Timeout        time.Duration
	RollbackAction string // empty = no rollback
	PreSteps       []string // action keys to run before this (e.g., "package_update_index")
}

// Registry is the hardcoded, deterministic action registry.
// This is the SINGLE SOURCE OF TRUTH for all commands TaaNOS can execute.
// NO AI involvement. NO dynamic generation.
var Registry = map[ActionKey]ActionDef{}

func init() {
	register := func(d ActionDef) {
		Registry[d.Key] = d
	}

	// =========================================================================
	// PACKAGE MANAGEMENT — APT (Debian/Ubuntu)
	// =========================================================================
	register(ActionDef{
		Key:             ActionKey{"package_management", "install", "apt"},
		Description:     "Install package via apt",
		CommandTemplate: "apt install -y {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         300 * time.Second,
		RollbackAction:  "remove",
		PreSteps:        []string{"update_index"},
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "remove", "apt"},
		Description:     "Remove package via apt",
		CommandTemplate: "apt remove -y {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         120 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "update", "apt"},
		Description:     "Update package index",
		CommandTemplate: "apt update",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         120 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "upgrade", "apt"},
		Description:     "Upgrade all packages",
		CommandTemplate: "apt upgrade -y",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         600 * time.Second,
		PreSteps:        []string{"update_index"},
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "list", "apt"},
		Description:     "List installed packages",
		CommandTemplate: "apt list --installed",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         30 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "show", "apt"},
		Description:     "Show package details",
		CommandTemplate: "apt show {target}",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         15 * time.Second,
	})

	// =========================================================================
	// PACKAGE MANAGEMENT — DNF (Fedora/RHEL/CentOS)
	// =========================================================================
	register(ActionDef{
		Key:             ActionKey{"package_management", "install", "dnf"},
		Description:     "Install package via dnf",
		CommandTemplate: "dnf install -y {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         300 * time.Second,
		RollbackAction:  "remove",
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "remove", "dnf"},
		Description:     "Remove package via dnf",
		CommandTemplate: "dnf remove -y {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         120 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "update", "dnf"},
		Description:     "Check for updates",
		CommandTemplate: "dnf check-update",
		RequiresRoot:    true,
		CanFail:         true, // returns exit code 100 if updates available
		Timeout:         120 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "upgrade", "dnf"},
		Description:     "Upgrade all packages",
		CommandTemplate: "dnf upgrade -y",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         600 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "list", "dnf"},
		Description:     "List installed packages",
		CommandTemplate: "dnf list installed",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         30 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "show", "dnf"},
		Description:     "Show package details",
		CommandTemplate: "dnf info {target}",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         15 * time.Second,
	})

	// =========================================================================
	// PACKAGE MANAGEMENT — PACMAN (Arch/Manjaro)
	// =========================================================================
	register(ActionDef{
		Key:             ActionKey{"package_management", "install", "pacman"},
		Description:     "Install package via pacman",
		CommandTemplate: "pacman -S --noconfirm {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         300 * time.Second,
		RollbackAction:  "remove",
		PreSteps:        []string{"update_index"},
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "remove", "pacman"},
		Description:     "Remove package via pacman",
		CommandTemplate: "pacman -R --noconfirm {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         120 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "update", "pacman"},
		Description:     "Sync package database",
		CommandTemplate: "pacman -Sy",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         120 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "upgrade", "pacman"},
		Description:     "Upgrade all packages",
		CommandTemplate: "pacman -Syu --noconfirm",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         600 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "list", "pacman"},
		Description:     "List installed packages",
		CommandTemplate: "pacman -Q",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         30 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "show", "pacman"},
		Description:     "Show package details",
		CommandTemplate: "pacman -Si {target}",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         15 * time.Second,
	})

	// =========================================================================
	// PACKAGE MANAGEMENT — ZYPPER (openSUSE)
	// =========================================================================
	register(ActionDef{
		Key:             ActionKey{"package_management", "install", "zypper"},
		Description:     "Install package via zypper",
		CommandTemplate: "zypper install -y {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         300 * time.Second,
		RollbackAction:  "remove",
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "remove", "zypper"},
		Description:     "Remove package via zypper",
		CommandTemplate: "zypper remove -y {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         120 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "update", "zypper"},
		Description:     "Refresh repository data",
		CommandTemplate: "zypper refresh",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         120 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "upgrade", "zypper"},
		Description:     "Upgrade all packages",
		CommandTemplate: "zypper update -y",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         600 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "list", "zypper"},
		Description:     "List installed packages",
		CommandTemplate: "zypper search -i",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         30 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"package_management", "show", "zypper"},
		Description:     "Show package details",
		CommandTemplate: "zypper info {target}",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         15 * time.Second,
	})

	// =========================================================================
	// SERVICE MANAGEMENT (systemd — all distros)
	// =========================================================================
	register(ActionDef{
		Key:             ActionKey{"service_management", "start", ""},
		Description:     "Start service",
		CommandTemplate: "systemctl start {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         30 * time.Second,
		RollbackAction:  "stop",
	})
	register(ActionDef{
		Key:             ActionKey{"service_management", "stop", ""},
		Description:     "Stop service",
		CommandTemplate: "systemctl stop {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         30 * time.Second,
		RollbackAction:  "start",
	})
	register(ActionDef{
		Key:             ActionKey{"service_management", "restart", ""},
		Description:     "Restart service",
		CommandTemplate: "systemctl restart {target}",
		RequiresRoot:    true,
		CanFail:         false,
		Timeout:         60 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"service_management", "enable", ""},
		Description:     "Enable service to start on boot",
		CommandTemplate: "systemctl enable {target}",
		RequiresRoot:    true,
		CanFail:         true,
		Timeout:         15 * time.Second,
		RollbackAction:  "disable",
	})
	register(ActionDef{
		Key:             ActionKey{"service_management", "disable", ""},
		Description:     "Disable service from starting on boot",
		CommandTemplate: "systemctl disable {target}",
		RequiresRoot:    true,
		CanFail:         true,
		Timeout:         15 * time.Second,
		RollbackAction:  "enable",
	})
	register(ActionDef{
		Key:             ActionKey{"service_management", "show", ""},
		Description:     "Show service status",
		CommandTemplate: "systemctl status {target}",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         10 * time.Second,
	})

	// =========================================================================
	// FILE OPERATIONS
	// =========================================================================
	register(ActionDef{
		Key:             ActionKey{"file_operation", "create", ""},
		Description:     "Create file",
		CommandTemplate: "touch {target}",
		RequiresRoot:    false,
		CanFail:         false,
		Timeout:         5 * time.Second,
		RollbackAction:  "delete",
	})
	register(ActionDef{
		Key:             ActionKey{"file_operation", "delete", ""},
		Description:     "Delete file",
		CommandTemplate: "rm {target}",
		RequiresRoot:    false,
		CanFail:         false,
		Timeout:         5 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"file_operation", "list", ""},
		Description:     "List directory contents",
		CommandTemplate: "ls -la {target}",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         10 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"file_operation", "show", ""},
		Description:     "Show file contents",
		CommandTemplate: "cat {target}",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         10 * time.Second,
	})

	// =========================================================================
	// SYSTEM INFO (read-only)
	// =========================================================================
	register(ActionDef{
		Key:             ActionKey{"system_info", "show", ""},
		Description:     "Show system information",
		CommandTemplate: "uname -a",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         5 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"system_info", "list", ""},
		Description:     "Show disk usage",
		CommandTemplate: "df -h",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         5 * time.Second,
	})

	// =========================================================================
	// NETWORK (read-only — NO firewall, NO config changes)
	// =========================================================================
	register(ActionDef{
		Key:             ActionKey{"network", "show", ""},
		Description:     "Show open ports",
		CommandTemplate: "ss -tlnp",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         10 * time.Second,
	})
	register(ActionDef{
		Key:             ActionKey{"network", "list", ""},
		Description:     "Show active connections",
		CommandTemplate: "ss -tunap",
		RequiresRoot:    false,
		CanFail:         true,
		Timeout:         10 * time.Second,
	})
}

// Lookup finds an action definition in the registry.
// For package operations, it requires the package manager name.
// For non-package operations, pkgManager can be empty.
func Lookup(category, action, pkgManager string) (*ActionDef, error) {
	// Try with package manager first (for package_management)
	if pkgManager != "" {
		key := ActionKey{category, action, pkgManager}
		if def, ok := Registry[key]; ok {
			return &def, nil
		}
	}

	// Try without package manager (for service_management, file_operation, etc.)
	key := ActionKey{category, action, ""}
	if def, ok := Registry[key]; ok {
		return &def, nil
	}

	return nil, fmt.Errorf("unsupported action: category=%s action=%s pkg_manager=%s",
		category, action, pkgManager)
}

// LookupUpdateIndex finds the "update" action for a given package manager.
func LookupUpdateIndex(pkgManager string) (*ActionDef, error) {
	return Lookup("package_management", "update", pkgManager)
}
