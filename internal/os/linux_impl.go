package osutil

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// =============================================================================
// APT (Debian/Ubuntu)
// =============================================================================

// AptManager implements PackageManager for apt-based systems.
type AptManager struct{}

func (a *AptManager) Name() string        { return "apt" }
func (a *AptManager) IsAvailable() bool    { _, err := exec.LookPath("apt"); return err == nil }
func (a *AptManager) NeedsUpdate() (bool, error) {
	// Check if apt cache is older than 24 hours — simplified for skeleton
	return true, nil
}
func (a *AptManager) IsInstalled(pkg string) (bool, error) {
	cmd := exec.Command("dpkg", "-s", pkg)
	err := cmd.Run()
	return err == nil, nil
}
func (a *AptManager) InstallCmd(pkg string) string  { return fmt.Sprintf("apt install -y %s", pkg) }
func (a *AptManager) RemoveCmd(pkg string) string    { return fmt.Sprintf("apt remove -y %s", pkg) }
func (a *AptManager) UpdateCmd() string              { return "apt update" }
func (a *AptManager) UpgradeCmd() string             { return "apt upgrade -y" }
func (a *AptManager) ListCmd() string                { return "apt list --installed" }
func (a *AptManager) ShowCmd(pkg string) string      { return fmt.Sprintf("apt show %s", pkg) }

// =============================================================================
// DNF (Fedora/RHEL/CentOS)
// =============================================================================

// DnfManager implements PackageManager for dnf-based systems.
type DnfManager struct{}

func (d *DnfManager) Name() string        { return "dnf" }
func (d *DnfManager) IsAvailable() bool    { _, err := exec.LookPath("dnf"); return err == nil }
func (d *DnfManager) NeedsUpdate() (bool, error) { return true, nil }
func (d *DnfManager) IsInstalled(pkg string) (bool, error) {
	cmd := exec.Command("rpm", "-q", pkg)
	err := cmd.Run()
	return err == nil, nil
}
func (d *DnfManager) InstallCmd(pkg string) string  { return fmt.Sprintf("dnf install -y %s", pkg) }
func (d *DnfManager) RemoveCmd(pkg string) string    { return fmt.Sprintf("dnf remove -y %s", pkg) }
func (d *DnfManager) UpdateCmd() string              { return "dnf check-update" }
func (d *DnfManager) UpgradeCmd() string             { return "dnf upgrade -y" }
func (d *DnfManager) ListCmd() string                { return "dnf list installed" }
func (d *DnfManager) ShowCmd(pkg string) string      { return fmt.Sprintf("dnf info %s", pkg) }

// =============================================================================
// PACMAN (Arch/Manjaro)
// =============================================================================

// PacmanManager implements PackageManager for pacman-based systems.
type PacmanManager struct{}

func (p *PacmanManager) Name() string        { return "pacman" }
func (p *PacmanManager) IsAvailable() bool    { _, err := exec.LookPath("pacman"); return err == nil }
func (p *PacmanManager) NeedsUpdate() (bool, error) { return true, nil }
func (p *PacmanManager) IsInstalled(pkg string) (bool, error) {
	cmd := exec.Command("pacman", "-Qi", pkg)
	err := cmd.Run()
	return err == nil, nil
}
func (p *PacmanManager) InstallCmd(pkg string) string  { return fmt.Sprintf("pacman -S --noconfirm %s", pkg) }
func (p *PacmanManager) RemoveCmd(pkg string) string    { return fmt.Sprintf("pacman -R --noconfirm %s", pkg) }
func (p *PacmanManager) UpdateCmd() string              { return "pacman -Sy" }
func (p *PacmanManager) UpgradeCmd() string             { return "pacman -Syu --noconfirm" }
func (p *PacmanManager) ListCmd() string                { return "pacman -Q" }
func (p *PacmanManager) ShowCmd(pkg string) string      { return fmt.Sprintf("pacman -Si %s", pkg) }

// =============================================================================
// ZYPPER (openSUSE)
// =============================================================================

// ZypperManager implements PackageManager for zypper-based systems.
type ZypperManager struct{}

func (z *ZypperManager) Name() string        { return "zypper" }
func (z *ZypperManager) IsAvailable() bool    { _, err := exec.LookPath("zypper"); return err == nil }
func (z *ZypperManager) NeedsUpdate() (bool, error) { return true, nil }
func (z *ZypperManager) IsInstalled(pkg string) (bool, error) {
	cmd := exec.Command("rpm", "-q", pkg)
	err := cmd.Run()
	return err == nil, nil
}
func (z *ZypperManager) InstallCmd(pkg string) string  { return fmt.Sprintf("zypper install -y %s", pkg) }
func (z *ZypperManager) RemoveCmd(pkg string) string    { return fmt.Sprintf("zypper remove -y %s", pkg) }
func (z *ZypperManager) UpdateCmd() string              { return "zypper refresh" }
func (z *ZypperManager) UpgradeCmd() string             { return "zypper update -y" }
func (z *ZypperManager) ListCmd() string                { return "zypper search -i" }
func (z *ZypperManager) ShowCmd(pkg string) string      { return fmt.Sprintf("zypper info %s", pkg) }

// =============================================================================
// SYSTEMD Service Manager
// =============================================================================

// SystemdManager implements ServiceManager for systemd-based Linux.
type SystemdManager struct{}

func (s *SystemdManager) Name() string { return "systemd" }
func (s *SystemdManager) IsRunning(svc string) (bool, error) {
	cmd := exec.Command("systemctl", "is-active", svc)
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out)) == "active", nil
}
func (s *SystemdManager) IsEnabled(svc string) (bool, error) {
	cmd := exec.Command("systemctl", "is-enabled", svc)
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out)) == "enabled", nil
}
func (s *SystemdManager) StartCmd(svc string) string   { return fmt.Sprintf("systemctl start %s", svc) }
func (s *SystemdManager) StopCmd(svc string) string    { return fmt.Sprintf("systemctl stop %s", svc) }
func (s *SystemdManager) RestartCmd(svc string) string { return fmt.Sprintf("systemctl restart %s", svc) }
func (s *SystemdManager) EnableCmd(svc string) string  { return fmt.Sprintf("systemctl enable %s", svc) }
func (s *SystemdManager) DisableCmd(svc string) string { return fmt.Sprintf("systemctl disable %s", svc) }
func (s *SystemdManager) StatusCmd(svc string) string  { return fmt.Sprintf("systemctl status %s", svc) }

// =============================================================================
// BASH Command Runner
// =============================================================================

// BashRunner implements CommandRunner for Linux bash.
type BashRunner struct{}

func (b *BashRunner) Shell() string { return "/bin/bash" }

func (b *BashRunner) Run(cmd string, args []string, timeout time.Duration) (*CmdResult, error) {
	return runCommand(false, cmd, args, timeout)
}

func (b *BashRunner) RunWithSudo(cmd string, args []string, timeout time.Duration) (*CmdResult, error) {
	return runCommand(true, cmd, args, timeout)
}

func runCommand(sudo bool, cmd string, args []string, timeout time.Duration) (*CmdResult, error) {
	fullArgs := args
	binary := cmd
	if sudo {
		fullArgs = append([]string{cmd}, args...)
		binary = "sudo"
	}

	c := exec.Command(binary, fullArgs...)

	start := time.Now()
	out, err := c.CombinedOutput()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}

	return &CmdResult{
		Command:  fmt.Sprintf("%s %s", binary, strings.Join(fullArgs, " ")),
		ExitCode: exitCode,
		Stdout:   string(out),
		Stderr:   "",
		Duration: duration,
	}, nil
}
