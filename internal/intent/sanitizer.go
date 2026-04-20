package intent

import (
	"runtime"
	"strings"
)

// SanitizeCommands converts Linux commands to Windows equivalents (and vice versa)
// when the model suggests commands for the wrong OS. Small models often default
// to Linux commands regardless of the system prompt.
func SanitizeCommands(commands FlexibleStrings) FlexibleStrings {
	if len(commands) == 0 {
		return commands
	}

	if runtime.GOOS == "windows" {
		return linuxToWindows(commands)
	}
	return commands
}

// linuxToWindows converts common Linux commands to their Windows/PowerShell equivalents.
func linuxToWindows(commands FlexibleStrings) FlexibleStrings {
	var result FlexibleStrings

	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		// Strip sudo prefix
		cmd = strings.TrimPrefix(cmd, "sudo ")

		lower := strings.ToLower(cmd)

		// Package management
		if strings.HasPrefix(lower, "apt-get update") || strings.HasPrefix(lower, "apt update") {
			result = append(result, "winget upgrade --include-unknown")
			continue
		}
		if strings.HasPrefix(lower, "apt-get install") || strings.HasPrefix(lower, "apt install") {
			pkg := extractAfter(cmd, "install")
			pkg = strings.TrimPrefix(pkg, "-y ")
			pkg = strings.TrimSpace(pkg)
			if pkg != "" {
				result = append(result, "winget install "+pkg)
			}
			continue
		}
		if strings.HasPrefix(lower, "apt-get remove") || strings.HasPrefix(lower, "apt remove") {
			pkg := extractAfter(cmd, "remove")
			pkg = strings.TrimPrefix(pkg, "-y ")
			pkg = strings.TrimSpace(pkg)
			if pkg != "" {
				result = append(result, "winget uninstall "+pkg)
			}
			continue
		}
		if strings.HasPrefix(lower, "yum ") || strings.HasPrefix(lower, "dnf ") || strings.HasPrefix(lower, "pacman ") {
			// Generic package manager → winget
			if strings.Contains(lower, "install") {
				pkg := extractLastWord(cmd)
				result = append(result, "winget install "+pkg)
			} else if strings.Contains(lower, "remove") || strings.Contains(lower, "erase") {
				pkg := extractLastWord(cmd)
				result = append(result, "winget uninstall "+pkg)
			}
			continue
		}

		// Service management
		if strings.HasPrefix(lower, "systemctl restart") {
			svc := extractAfter(cmd, "restart")
			result = append(result, "Restart-Service "+strings.TrimSpace(svc))
			continue
		}
		if strings.HasPrefix(lower, "systemctl start") {
			svc := extractAfter(cmd, "start")
			result = append(result, "Start-Service "+strings.TrimSpace(svc))
			continue
		}
		if strings.HasPrefix(lower, "systemctl stop") {
			svc := extractAfter(cmd, "stop")
			result = append(result, "Stop-Service "+strings.TrimSpace(svc))
			continue
		}
		if strings.HasPrefix(lower, "systemctl status") {
			svc := extractAfter(cmd, "status")
			result = append(result, "Get-Service "+strings.TrimSpace(svc))
			continue
		}
		if strings.HasPrefix(lower, "systemctl enable") {
			svc := extractAfter(cmd, "enable")
			result = append(result, "Set-Service -Name "+strings.TrimSpace(svc)+" -StartupType Automatic")
			continue
		}

		// File system
		if strings.HasPrefix(lower, "df -h") || strings.HasPrefix(lower, "df") {
			result = append(result, "Get-PSDrive -PSProvider FileSystem")
			continue
		}
		if strings.HasPrefix(lower, "free -") || lower == "free" {
			result = append(result, "Get-CimInstance Win32_OperatingSystem | Select FreePhysicalMemory,TotalVisibleMemorySize")
			continue
		}
		if strings.HasPrefix(lower, "rm -rf ") || strings.HasPrefix(lower, "rm -f ") || strings.HasPrefix(lower, "rm ") {
			target := extractLastWord(cmd)
			result = append(result, "Remove-Item -Recurse -Force "+target)
			continue
		}
		if strings.HasPrefix(lower, "ls") {
			result = append(result, "Get-ChildItem")
			continue
		}
		if strings.HasPrefix(lower, "cat ") {
			target := extractAfter(cmd, "cat")
			result = append(result, "Get-Content "+strings.TrimSpace(target))
			continue
		}

		// Network
		if strings.HasPrefix(lower, "ss -") || strings.HasPrefix(lower, "netstat") {
			result = append(result, "Get-NetTCPConnection | Where-Object State -eq Listen")
			continue
		}

		// Uptime, uname etc
		if lower == "uptime" {
			result = append(result, "(Get-CimInstance Win32_OperatingSystem).LastBootUpTime")
			continue
		}
		if strings.HasPrefix(lower, "uname") {
			result = append(result, "[System.Environment]::OSVersion")
			continue
		}

		// If no conversion matched, keep original (might be a valid PowerShell command)
		result = append(result, cmd)
	}

	return result
}

// extractAfter returns everything after the first occurrence of keyword in the string.
func extractAfter(s, keyword string) string {
	lower := strings.ToLower(s)
	idx := strings.Index(lower, strings.ToLower(keyword))
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(s[idx+len(keyword):])
}

// extractLastWord returns the last whitespace-separated token.
func extractLastWord(s string) string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	// Strip common flags
	if strings.HasPrefix(last, "-") {
		if len(parts) >= 2 {
			return parts[len(parts)-2]
		}
		return ""
	}
	return last
}
