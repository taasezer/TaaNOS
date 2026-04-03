//go:build linux

package osutil

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// PortInfo represents an open port on the system.
type PortInfo struct {
	Protocol string // "tcp" or "udp"
	Port     int
	PID      int
	Process  string
	State    string // "LISTEN", "ESTABLISHED", etc.
	Address  string // "0.0.0.0", "127.0.0.1", "::", etc.
}

// ProcessInfo represents a running process.
type ProcessInfo struct {
	PID     int
	Name    string
	User    string
	CPU     float64
	Memory  float64
	Command string
}

// NetworkScanner provides read-only network inspection.
// NO firewall manipulation. NO network configuration changes.
type NetworkScanner struct{}

// NewNetworkScanner creates a NetworkScanner.
func NewNetworkScanner() *NetworkScanner {
	return &NetworkScanner{}
}

// ListOpenPorts returns all listening ports using `ss` or `netstat`.
func (n *NetworkScanner) ListOpenPorts() ([]PortInfo, error) {
	out, err := exec.Command("ss", "-tlnp").Output()
	if err != nil {
		out, err = exec.Command("netstat", "-tlnp").Output()
		if err != nil {
			return nil, fmt.Errorf("neither ss nor netstat available: %w", err)
		}
		return parseNetstat(string(out))
	}
	return parseSS(string(out))
}

// IsPortOpen checks if a specific port is listening.
func (n *NetworkScanner) IsPortOpen(port int) (bool, error) {
	ports, err := n.ListOpenPorts()
	if err != nil {
		return false, err
	}
	for _, p := range ports {
		if p.Port == port {
			return true, nil
		}
	}
	return false, nil
}

// ListProcesses returns running processes using `ps`.
func (n *NetworkScanner) ListProcesses() ([]ProcessInfo, error) {
	out, err := exec.Command("ps", "aux", "--no-headers").Output()
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}

	var processes []ProcessInfo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pid, _ := strconv.Atoi(fields[1])
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)

		processes = append(processes, ProcessInfo{
			PID:     pid,
			Name:    fields[10],
			User:    fields[0],
			CPU:     cpu,
			Memory:  mem,
			Command: strings.Join(fields[10:], " "),
		})
	}

	return processes, nil
}

// FindProcess searches for a process by name.
func (n *NetworkScanner) FindProcess(name string) ([]ProcessInfo, error) {
	all, err := n.ListProcesses()
	if err != nil {
		return nil, err
	}

	var matched []ProcessInfo
	for _, p := range all {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(name)) ||
			strings.Contains(strings.ToLower(p.Command), strings.ToLower(name)) {
			matched = append(matched, p)
		}
	}
	return matched, nil
}

// GetPortProcess returns the process using a specific port.
func (n *NetworkScanner) GetPortProcess(port int) (*PortInfo, error) {
	ports, err := n.ListOpenPorts()
	if err != nil {
		return nil, err
	}
	for _, p := range ports {
		if p.Port == port {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("no process found on port %d", port)
}

// parseSS parses output from `ss -tlnp`.
func parseSS(output string) ([]PortInfo, error) {
	var ports []PortInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	if scanner.Scan() {
		// discard header line
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		localAddr := fields[3]
		addr, portNum := parseAddrPort(localAddr)

		info := PortInfo{
			Protocol: "tcp",
			Port:     portNum,
			State:    fields[0],
			Address:  addr,
		}

		if len(fields) >= 6 {
			info.Process, info.PID = parseSSProcess(fields[5])
		}

		if portNum > 0 {
			ports = append(ports, info)
		}
	}

	return ports, nil
}

// parseNetstat parses output from `netstat -tlnp`.
func parseNetstat(output string) ([]PortInfo, error) {
	var ports []PortInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "tcp") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		localAddr := fields[3]
		addr, portNum := parseAddrPort(localAddr)

		info := PortInfo{
			Protocol: "tcp",
			Port:     portNum,
			State:    fields[5],
			Address:  addr,
		}

		if len(fields) >= 7 {
			pidProc := fields[6]
			parts := strings.SplitN(pidProc, "/", 2)
			if len(parts) == 2 {
				info.PID, _ = strconv.Atoi(parts[0])
				info.Process = parts[1]
			}
		}

		if portNum > 0 {
			ports = append(ports, info)
		}
	}

	return ports, nil
}

// parseAddrPort splits "0.0.0.0:8080" or "[::]:8080" into address and port.
func parseAddrPort(s string) (string, int) {
	if idx := strings.LastIndex(s, ":"); idx >= 0 {
		addr := s[:idx]
		portStr := s[idx+1:]
		port, _ := strconv.Atoi(portStr)
		return addr, port
	}
	return s, 0
}

// parseSSProcess extracts PID and process name from ss output.
func parseSSProcess(s string) (string, int) {
	if !strings.Contains(s, "pid=") {
		return "", 0
	}

	name := ""
	if start := strings.Index(s, "((\""); start >= 0 {
		rest := s[start+3:]
		if end := strings.Index(rest, "\""); end >= 0 {
			name = rest[:end]
		}
	}

	pid := 0
	if pidIdx := strings.Index(s, "pid="); pidIdx >= 0 {
		rest := s[pidIdx+4:]
		if end := strings.IndexAny(rest, ",)"); end >= 0 {
			pid, _ = strconv.Atoi(rest[:end])
		}
	}

	return name, pid
}
