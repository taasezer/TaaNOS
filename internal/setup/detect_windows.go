//go:build windows

package setup

import (
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
)

// MEMORYSTATUSEX structure
type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func detectKernel() string {
	return "" // Feature requested by user only for Linux initially
}

func detectRAM() (int64, int64) {
	var mem memoryStatusEx
	mem.Length = uint32(unsafe.Sizeof(mem))

	ret, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&mem)))
	if ret == 0 {
		return 0, 0
	}

	totalMB := int64(mem.TotalPhys / (1024 * 1024))
	availMB := int64(mem.AvailPhys / (1024 * 1024))
	return totalMB, availMB
}

func detectVirtualization() (bool, string) {
	// Try wmic (works on all Windows versions)
	out, err := exec.Command("cmd", "/C", "wmic computersystem get model /value").Output()
	if err == nil {
		model := strings.TrimSpace(strings.ToLower(string(out)))
		vmIndicators := map[string]string{
			"virtualbox":      "virtualbox",
			"vmware":          "vmware",
			"virtual machine": "hyper-v",
			"kvm":             "kvm",
			"qemu":            "qemu",
			"xen":             "xen",
		}
		for indicator, name := range vmIndicators {
			if strings.Contains(model, indicator) {
				return true, name
			}
		}
	}
	return false, ""
}
