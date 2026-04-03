package osutil

import (
	"os/exec"
)

// LinuxSystemInfo implements SystemInfo for Linux.
// Methods are split across linux_sysinfo.go (linux) and linux_sysinfo_other.go (non-linux).
type LinuxSystemInfo struct{}

// LinuxPlatform implements Platform for Linux.
type LinuxPlatform struct {
	pkgMgr  PackageManager
	svcMgr  ServiceManager
	cmdRun  CommandRunner
	sysInfo SystemInfo
}

// NewLinuxPlatform detects the Linux distro and sets up the appropriate managers.
func NewLinuxPlatform() (*LinuxPlatform, error) {
	pkgMgr := detectLinuxPackageManager()
	return &LinuxPlatform{
		pkgMgr:  pkgMgr,
		svcMgr:  &SystemdManager{},
		cmdRun:  &BashRunner{},
		sysInfo: &LinuxSystemInfo{},
	}, nil
}

func (l *LinuxPlatform) Name() string                 { return "linux" }
func (l *LinuxPlatform) PackageManager() PackageManager { return l.pkgMgr }
func (l *LinuxPlatform) ServiceManager() ServiceManager { return l.svcMgr }
func (l *LinuxPlatform) CommandRunner() CommandRunner   { return l.cmdRun }
func (l *LinuxPlatform) SystemInfo() SystemInfo         { return l.sysInfo }

// detectLinuxPackageManager probes for available package managers in priority order.
func detectLinuxPackageManager() PackageManager {
	// Priority: apt > dnf > pacman > zypper
	managers := []struct {
		binary string
		create func() PackageManager
	}{
		{"apt", func() PackageManager { return &AptManager{} }},
		{"dnf", func() PackageManager { return &DnfManager{} }},
		{"pacman", func() PackageManager { return &PacmanManager{} }},
		{"zypper", func() PackageManager { return &ZypperManager{} }},
	}

	for _, m := range managers {
		if _, err := exec.LookPath(m.binary); err == nil {
			return m.create()
		}
	}

	// Fallback: return apt (will report unavailable)
	return &AptManager{}
}
