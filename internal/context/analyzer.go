package context

import (
	"fmt"

	osutil "github.com/taasezer/TaaNOS/internal/os"
)

// Analyzer queries the current system state to provide context for the planner.
type Analyzer struct {
	platform osutil.Platform
}

// NewAnalyzer creates a Context Analyzer using the detected platform.
func NewAnalyzer(platform osutil.Platform) *Analyzer {
	return &Analyzer{
		platform: platform,
	}
}

// Analyze gathers system context relevant to the given intent target.
// It queries OS info, package manager state, target state, user info, and resources.
func (a *Analyzer) Analyze(category, action, target string) (*SystemContext, error) {
	ctx := &SystemContext{}

	// 1. OS detection
	osInfo, err := a.platform.SystemInfo().OS()
	if err != nil {
		return nil, fmt.Errorf("OS detection failed: %w", err)
	}
	ctx.OS = OSInfo{
		Name:    osInfo.Name,
		Distro:  osInfo.Distro,
		Version: osInfo.Version,
		Arch:    osInfo.Arch,
		Kernel:  osInfo.Kernel,
	}

	// 2. Package manager state
	pkgMgr := a.platform.PackageManager()
	needsUpdate, _ := pkgMgr.NeedsUpdate()
	ctx.PackageManager = PackageManagerInfo{
		Name:       pkgMgr.Name(),
		Available:  pkgMgr.IsAvailable(),
		NeedsUpdate: needsUpdate,
	}

	// 3. Target state (depends on category)
	targetState, err := a.probeTargetState(category, target)
	if err != nil {
		// Non-fatal — we can still plan with partial context
		targetState = &TargetState{}
	}
	ctx.TargetState = *targetState

	// 4. User info
	userInfo, err := a.platform.SystemInfo().User()
	if err != nil {
		return nil, fmt.Errorf("user detection failed: %w", err)
	}
	ctx.User = UserInfo{
		Name:          userInfo.Name,
		UID:           userInfo.UID,
		IsRoot:        userInfo.IsRoot,
		SudoAvailable: userInfo.SudoAvailable,
	}

	// 5. Resource checks
	diskInfo, err := a.platform.SystemInfo().Disk()
	if err == nil {
		ctx.Resources.DiskFreeMB = diskInfo.FreeMB
	}

	memInfo, err := a.platform.SystemInfo().Memory()
	if err == nil {
		ctx.Resources.MemoryFreeMB = memInfo.FreeMB
	}

	// 6. Dependencies met (simplified — full check in validator)
	ctx.DependenciesMet = true

	return ctx, nil
}

// probeTargetState checks the current state of the target based on category.
func (a *Analyzer) probeTargetState(category, target string) (*TargetState, error) {
	if target == "" {
		return &TargetState{}, nil
	}

	state := &TargetState{}

	switch category {
	case "package_management":
		installed, err := a.platform.PackageManager().IsInstalled(target)
		if err == nil {
			state.Installed = boolPtr(installed)
		}

	case "service_management":
		running, err := a.platform.ServiceManager().IsRunning(target)
		if err == nil {
			state.Running = boolPtr(running)
		}
		enabled, err := a.platform.ServiceManager().IsEnabled(target)
		if err == nil {
			state.Enabled = boolPtr(enabled)
		}

	case "network":
		// Network target probing handled by network scanner
		// No target state to probe here

	case "file_operation":
		// File existence could be checked here
		// Deferred for now

	case "system_info":
		// No target state needed
	}

	return state, nil
}

func boolPtr(b bool) *bool {
	return &b
}
