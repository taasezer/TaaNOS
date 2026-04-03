package osutil

import "time"

// MacOSStub is a placeholder for the macOS platform.
// Full implementation deferred to future release.
type MacOSStub struct{}

func (m *MacOSStub) Name() string                     { return "darwin" }
func (m *MacOSStub) PackageManager() PackageManager   { return &BrewStub{} }
func (m *MacOSStub) ServiceManager() ServiceManager   { return &LaunchctlStub{} }
func (m *MacOSStub) CommandRunner() CommandRunner     { return &ZshStub{} }
func (m *MacOSStub) SystemInfo() SystemInfo           { return &MacSystemInfoStub{} }

// BrewStub is a placeholder for Homebrew.
type BrewStub struct{}
func (b *BrewStub) Name() string                        { return "brew" }
func (b *BrewStub) IsAvailable() bool                   { return false }
func (b *BrewStub) NeedsUpdate() (bool, error)          { return false, nil }
func (b *BrewStub) IsInstalled(pkg string) (bool, error) { return false, nil }
func (b *BrewStub) InstallCmd(pkg string) string        { return "" }
func (b *BrewStub) RemoveCmd(pkg string) string         { return "" }
func (b *BrewStub) UpdateCmd() string                   { return "" }
func (b *BrewStub) UpgradeCmd() string                  { return "" }
func (b *BrewStub) ListCmd() string                     { return "" }
func (b *BrewStub) ShowCmd(pkg string) string           { return "" }

// LaunchctlStub is a placeholder for macOS service management.
type LaunchctlStub struct{}
func (l *LaunchctlStub) Name() string                        { return "launchctl" }
func (l *LaunchctlStub) IsRunning(svc string) (bool, error)  { return false, nil }
func (l *LaunchctlStub) IsEnabled(svc string) (bool, error)  { return false, nil }
func (l *LaunchctlStub) StartCmd(svc string) string          { return "" }
func (l *LaunchctlStub) StopCmd(svc string) string           { return "" }
func (l *LaunchctlStub) RestartCmd(svc string) string        { return "" }
func (l *LaunchctlStub) EnableCmd(svc string) string         { return "" }
func (l *LaunchctlStub) DisableCmd(svc string) string        { return "" }
func (l *LaunchctlStub) StatusCmd(svc string) string         { return "" }

// ZshStub is a placeholder for zsh runner.
type ZshStub struct{}
func (z *ZshStub) Shell() string { return "/bin/zsh" }
func (z *ZshStub) Run(cmd string, args []string, timeout time.Duration) (*CmdResult, error) {
	return nil, nil
}
func (z *ZshStub) RunWithSudo(cmd string, args []string, timeout time.Duration) (*CmdResult, error) {
	return nil, nil
}

// MacSystemInfoStub is a placeholder.
type MacSystemInfoStub struct{}
func (m *MacSystemInfoStub) OS() (OSInfo, error)       { return OSInfo{Name: "darwin"}, nil }
func (m *MacSystemInfoStub) User() (UserInfo, error)   { return UserInfo{}, nil }
func (m *MacSystemInfoStub) Disk() (DiskInfo, error)   { return DiskInfo{}, nil }
func (m *MacSystemInfoStub) Memory() (MemoryInfo, error) { return MemoryInfo{}, nil }
