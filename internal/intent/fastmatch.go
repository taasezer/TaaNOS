package intent

import (
	"runtime"
	"strings"
)

// localMatch holds a pre-built intent result for common commands.
type localMatch struct {
	Category string
	Action   string
	Target   string
	Commands []string
}

// TryLocalMatch attempts to resolve natural language input deterministically
// without requiring an active Ollama connection.
// Supports both English and Turkish syntax.
func TryLocalMatch(input string) *IntentResult {
	trimmed := strings.TrimSpace(input)
	lower := strings.ToLower(trimmed)
	words := strings.Fields(lower)
	if len(words) == 0 {
		return nil
	}

	var m *localMatch

	// 1. Common idiomatic phrases (Checked first for highest accuracy)
	switch lower {
	case "sistemi güncelle", "paketleri güncelle", "update system", "upgrade system", "sistem güncelle", "paket güncelle":
		m = &localMatch{Category: "package_management", Action: "update", Target: "all", Commands: updateCmd("all")}
	case "açık portlar", "açık portları göster", "list ports", "show ports", "portları göster":
		m = &localMatch{Category: "system_info", Action: "ports", Target: "network", Commands: portsCmd()}
	case "disk alanı", "disk durumu", "disk space", "disk kullanım":
		m = &localMatch{Category: "system_info", Action: "disk", Target: "storage", Commands: diskCmd()}
	case "ram durumu", "bellek durumu", "memory usage", "ram kullanım":
		m = &localMatch{Category: "system_info", Action: "memory", Target: "ram", Commands: ramCmd()}
	}

	// 2. Single-word commands
	if m == nil && len(words) == 1 {
		switch words[0] {
		case "update", "upgrade", "güncelle":
			m = &localMatch{
				Category: "package_management",
				Action:   "update",
				Target:   "all",
				Commands: updateCmd("all"),
			}
		case "ports", "portlar":
			m = &localMatch{
				Category: "system_info",
				Action:   "ports",
				Target:   "network",
				Commands: portsCmd(),
			}
		case "disk":
			m = &localMatch{
				Category: "system_info",
				Action:   "disk",
				Target:   "storage",
				Commands: diskCmd(),
			}
		case "ram", "memory", "hafıza":
			m = &localMatch{
				Category: "system_info",
				Action:   "memory",
				Target:   "ram",
				Commands: ramCmd(),
			}
		default:
			return nil
		}
	}

	// 3. Multi-word commands (Prefix and Suffix verb patterns)
	if m == nil && len(words) >= 2 {
		first := words[0]
		last := words[len(words)-1]

		// Prefix verb: "install nginx", "kur nginx", "kaldır docker"
		switch first {
		case "install":
			target := strings.Join(words[1:], " ")
			m = &localMatch{Category: "package_management", Action: "install", Target: target, Commands: installCmd(target)}
		case "kur", "yükle":
			target := strings.Join(words[1:], " ")
			m = &localMatch{Category: "package_management", Action: "install", Target: target, Commands: installCmd(target)}
		case "remove", "uninstall":
			target := strings.Join(words[1:], " ")
			m = &localMatch{Category: "package_management", Action: "remove", Target: target, Commands: removeCmd(target)}
		case "kaldır", "sil":
			target := strings.Join(words[1:], " ")
			m = &localMatch{Category: "package_management", Action: "remove", Target: target, Commands: removeCmd(target)}
		case "update", "upgrade":
			target := strings.Join(words[1:], " ")
			m = &localMatch{Category: "package_management", Action: "update", Target: target, Commands: updateCmd(target)}
		case "güncelle":
			target := strings.Join(words[1:], " ")
			m = &localMatch{Category: "package_management", Action: "update", Target: target, Commands: updateCmd(target)}
		case "restart":
			target := strings.Join(words[1:], " ")
			m = &localMatch{Category: "service_management", Action: "restart", Target: target, Commands: serviceCmd("restart", target)}
		case "start":
			target := strings.Join(words[1:], " ")
			m = &localMatch{Category: "service_management", Action: "start", Target: target, Commands: serviceCmd("start", target)}
		case "stop":
			target := strings.Join(words[1:], " ")
			m = &localMatch{Category: "service_management", Action: "stop", Target: target, Commands: serviceCmd("stop", target)}
		}

		// Suffix verb (Turkish SOV order): "nginx kur", "docker kaldır", "apache yeniden başlat"
		if m == nil {
			switch last {
			case "kur", "yükle":
				target := strings.Join(words[:len(words)-1], " ")
				m = &localMatch{Category: "package_management", Action: "install", Target: target, Commands: installCmd(target)}
			case "kaldır", "sil":
				target := strings.Join(words[:len(words)-1], " ")
				m = &localMatch{Category: "package_management", Action: "remove", Target: target, Commands: removeCmd(target)}
			case "başlat":
				target := strings.Join(words[:len(words)-1], " ")
				if strings.HasSuffix(target, "yeniden") {
					target = strings.TrimSpace(strings.TrimSuffix(target, "yeniden"))
					m = &localMatch{Category: "service_management", Action: "restart", Target: target, Commands: serviceCmd("restart", target)}
				} else {
					m = &localMatch{Category: "service_management", Action: "start", Target: target, Commands: serviceCmd("start", target)}
				}
			case "durdur":
				target := strings.Join(words[:len(words)-1], " ")
				m = &localMatch{Category: "service_management", Action: "stop", Target: target, Commands: serviceCmd("stop", target)}
			case "güncelle":
				target := strings.Join(words[:len(words)-1], " ")
				m = &localMatch{Category: "package_management", Action: "update", Target: target, Commands: updateCmd(target)}
			}
		}
	}

	if m == nil {
		return nil
	}

	cmds := make(FlexibleStrings, len(m.Commands))
	copy(cmds, m.Commands)

	return &IntentResult{
		Intent:            m.Action + " " + m.Target,
		Category:          Category(m.Category),
		Action:            Action(m.Action),
		Parameters:        Parameters{Target: m.Target, Scope: "system"},
		Confidence:        1.0,
		SuggestedCommands: cmds,
		ExtractionTimeMs:  0,
	}
}

func installCmd(target string) []string {
	if runtime.GOOS == "windows" {
		return []string{"winget install " + target}
	}
	return []string{"sudo apt-get install -y " + target}
}

func removeCmd(target string) []string {
	if runtime.GOOS == "windows" {
		return []string{"winget uninstall " + target}
	}
	return []string{"sudo apt-get remove -y " + target}
}

func updateCmd(target string) []string {
	if runtime.GOOS == "windows" {
		if target == "all" || target == "system" || target == "everything" || target == "sistem" {
			return []string{"winget upgrade --all"}
		}
		return []string{"winget upgrade " + target}
	}
	if target == "all" || target == "system" || target == "everything" || target == "sistem" {
		return []string{"sudo apt-get update && sudo apt-get upgrade -y"}
	}
	return []string{"sudo apt-get install --only-upgrade -y " + target}
}

func serviceCmd(action, target string) []string {
	if runtime.GOOS == "windows" {
		switch action {
		case "restart":
			return []string{"Restart-Service " + target}
		case "start":
			return []string{"Start-Service " + target}
		case "stop":
			return []string{"Stop-Service " + target}
		}
	}
	return []string{"sudo systemctl " + action + " " + target}
}

func portsCmd() []string {
	if runtime.GOOS == "windows" {
		return []string{"netstat -ano"}
	}
	return []string{"sudo ss -tulanp"}
}

func diskCmd() []string {
	if runtime.GOOS == "windows" {
		return []string{"Get-PSDrive -PSProvider FileSystem"}
	}
	return []string{"df -h /"}
}

func ramCmd() []string {
	if runtime.GOOS == "windows" {
		return []string{"Get-CimInstance Win32_OperatingSystem | Select-Object TotalVisibleMemorySize,FreePhysicalMemory"}
	}
	return []string{"free -h"}
}
