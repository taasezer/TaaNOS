//go:build linux

package setup

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func detectKernel() string {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err == nil {
		return strings.TrimSpace(string(data))
	}
	return ""
}

func detectRAM() (int64, int64) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0
	}

	var totalKB, availableKB int64
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			totalKB = val
		case "MemAvailable:":
			availableKB = val
		}
	}

	return totalKB / 1024, availableKB / 1024
}

func detectVirtualization() (bool, string) {
	// Method 1: systemd-detect-virt
	out, err := exec.Command("systemd-detect-virt").Output()
	if err == nil {
		virt := strings.TrimSpace(string(out))
		if virt != "none" && virt != "" {
			return true, virt
		}
	}

	// Method 2: Check DMI product name
	data, err := os.ReadFile("/sys/class/dmi/id/product_name")
	if err == nil {
		product := strings.TrimSpace(strings.ToLower(string(data)))
		vmIndicators := []string{"virtualbox", "vmware", "kvm", "qemu", "hyper-v", "xen", "parallels"}
		for _, indicator := range vmIndicators {
			if strings.Contains(product, indicator) {
				return true, indicator
			}
		}
	}

	// Method 3: Check if /proc/cpuinfo contains hypervisor flag
	cpuinfo, err := os.ReadFile("/proc/cpuinfo")
	if err == nil && strings.Contains(string(cpuinfo), "hypervisor") {
		return true, "hypervisor"
	}

	return false, ""
}
