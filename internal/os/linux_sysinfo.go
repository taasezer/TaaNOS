//go:build linux

package osutil

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// OS detects the Linux distribution, version, architecture, and kernel.
func (l *LinuxSystemInfo) OS() (OSInfo, error) {
	info := OSInfo{
		Name: "linux",
		Arch: runtime.GOARCH,
	}

	// Kernel version from uname
	out, err := exec.Command("uname", "-r").Output()
	if err == nil {
		info.Kernel = strings.TrimSpace(string(out))
	}

	// Distribution info from /etc/os-release
	distro, version, err := parseOSRelease()
	if err == nil {
		info.Distro = distro
		info.Version = version
	}

	return info, nil
}

// User detects the current user's identity and sudo availability.
func (l *LinuxSystemInfo) User() (UserInfo, error) {
	u, err := user.Current()
	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to get current user: %w", err)
	}

	uid, _ := strconv.Atoi(u.Uid)

	info := UserInfo{
		Name:   u.Username,
		UID:    uid,
		IsRoot: uid == 0,
	}

	// Check sudo availability
	_, err = exec.LookPath("sudo")
	if err == nil {
		cmd := exec.Command("sudo", "-n", "true")
		err = cmd.Run()
		info.SudoAvailable = err == nil
	}

	return info, nil
}

// Disk reports free and total disk space for the root filesystem.
func (l *LinuxSystemInfo) Disk() (DiskInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return DiskInfo{}, fmt.Errorf("failed to stat filesystem: %w", err)
	}

	totalBytes := stat.Blocks * uint64(stat.Bsize)
	freeBytes := stat.Bavail * uint64(stat.Bsize)

	return DiskInfo{
		TotalMB: int64(totalBytes / 1024 / 1024),
		FreeMB:  int64(freeBytes / 1024 / 1024),
	}, nil
}

// Memory reports free and total system memory from /proc/meminfo.
func (l *LinuxSystemInfo) Memory() (MemoryInfo, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("failed to read /proc/meminfo: %w", err)
	}
	defer f.Close()

	var totalKB, freeKB, availableKB int64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			totalKB = value
		case "MemFree:":
			freeKB = value
		case "MemAvailable:":
			availableKB = value
		}
	}

	free := availableKB
	if free == 0 {
		free = freeKB
	}

	return MemoryInfo{
		TotalMB: totalKB / 1024,
		FreeMB:  free / 1024,
	}, nil
}

// parseOSRelease reads /etc/os-release and extracts ID and VERSION_ID.
func parseOSRelease() (distro string, version string, err error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			distro = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}

	return distro, version, nil
}
