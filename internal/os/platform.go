package osutil

import "time"

// CmdResult holds the result of executing a shell command.
type CmdResult struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Platform is the top-level OS abstraction interface.
type Platform interface {
	Name() string
	PackageManager() PackageManager
	ServiceManager() ServiceManager
	CommandRunner() CommandRunner
	SystemInfo() SystemInfo
}

// PackageManager abstracts package management across distros.
type PackageManager interface {
	Name() string
	IsAvailable() bool
	NeedsUpdate() (bool, error)
	IsInstalled(pkg string) (bool, error)
	InstallCmd(pkg string) string
	RemoveCmd(pkg string) string
	UpdateCmd() string
	UpgradeCmd() string
	ListCmd() string
	ShowCmd(pkg string) string
}

// ServiceManager abstracts service management.
type ServiceManager interface {
	Name() string
	IsRunning(svc string) (bool, error)
	IsEnabled(svc string) (bool, error)
	StartCmd(svc string) string
	StopCmd(svc string) string
	RestartCmd(svc string) string
	EnableCmd(svc string) string
	DisableCmd(svc string) string
	StatusCmd(svc string) string
}

// CommandRunner abstracts shell command execution.
type CommandRunner interface {
	Run(cmd string, args []string, timeout time.Duration) (*CmdResult, error)
	RunWithSudo(cmd string, args []string, timeout time.Duration) (*CmdResult, error)
	Shell() string
}

// OSInfo holds OS identification.
type OSInfo struct {
	Name    string
	Distro  string
	Version string
	Arch    string
	Kernel  string
}

// UserInfo holds user details.
type UserInfo struct {
	Name          string
	UID           int
	IsRoot        bool
	SudoAvailable bool
}

// DiskInfo holds disk usage stats.
type DiskInfo struct {
	FreeMB  int64
	TotalMB int64
}

// MemoryInfo holds memory stats.
type MemoryInfo struct {
	FreeMB  int64
	TotalMB int64
}

// SystemInfo abstracts system information queries.
type SystemInfo interface {
	OS() (OSInfo, error)
	User() (UserInfo, error)
	Disk() (DiskInfo, error)
	Memory() (MemoryInfo, error)
}
