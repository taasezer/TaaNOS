package osutil

import "time"

// WindowsStub is a placeholder for the Windows platform.
// Full implementation deferred to future release.
type WindowsStub struct{}

func (w *WindowsStub) Name() string                     { return "windows" }
func (w *WindowsStub) PackageManager() PackageManager   { return &ChocoStub{} }
func (w *WindowsStub) ServiceManager() ServiceManager   { return &SCStub{} }
func (w *WindowsStub) CommandRunner() CommandRunner     { return &PowerShellStub{} }
func (w *WindowsStub) SystemInfo() SystemInfo           { return &WindowsSystemInfoStub{} }

// ChocoStub is a placeholder for Chocolatey package manager.
type ChocoStub struct{}
func (c *ChocoStub) Name() string                        { return "choco" }
func (c *ChocoStub) IsAvailable() bool                   { return false }
func (c *ChocoStub) NeedsUpdate() (bool, error)          { return false, nil }
func (c *ChocoStub) IsInstalled(pkg string) (bool, error) { return false, nil }
func (c *ChocoStub) InstallCmd(pkg string) string        { return "" }
func (c *ChocoStub) RemoveCmd(pkg string) string         { return "" }
func (c *ChocoStub) UpdateCmd() string                   { return "" }
func (c *ChocoStub) UpgradeCmd() string                  { return "" }
func (c *ChocoStub) ListCmd() string                     { return "" }
func (c *ChocoStub) ShowCmd(pkg string) string           { return "" }

// SCStub is a placeholder for Windows Service Control.
type SCStub struct{}
func (s *SCStub) Name() string                        { return "sc" }
func (s *SCStub) IsRunning(svc string) (bool, error)  { return false, nil }
func (s *SCStub) IsEnabled(svc string) (bool, error)  { return false, nil }
func (s *SCStub) StartCmd(svc string) string          { return "" }
func (s *SCStub) StopCmd(svc string) string           { return "" }
func (s *SCStub) RestartCmd(svc string) string        { return "" }
func (s *SCStub) EnableCmd(svc string) string         { return "" }
func (s *SCStub) DisableCmd(svc string) string        { return "" }
func (s *SCStub) StatusCmd(svc string) string         { return "" }

// PowerShellStub is a placeholder for PowerShell runner.
type PowerShellStub struct{}
func (p *PowerShellStub) Shell() string { return "powershell.exe" }
func (p *PowerShellStub) Run(cmd string, args []string, timeout time.Duration) (*CmdResult, error) {
	return nil, nil
}
func (p *PowerShellStub) RunWithSudo(cmd string, args []string, timeout time.Duration) (*CmdResult, error) {
	return nil, nil
}

// WindowsSystemInfoStub is a placeholder.
type WindowsSystemInfoStub struct{}
func (w *WindowsSystemInfoStub) OS() (OSInfo, error)       { return OSInfo{Name: "windows"}, nil }
func (w *WindowsSystemInfoStub) User() (UserInfo, error)   { return UserInfo{}, nil }
func (w *WindowsSystemInfoStub) Disk() (DiskInfo, error)   { return DiskInfo{}, nil }
func (w *WindowsSystemInfoStub) Memory() (MemoryInfo, error) { return MemoryInfo{}, nil }
