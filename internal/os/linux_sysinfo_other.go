//go:build !linux

package osutil

// Fallback LinuxSystemInfo for non-Linux platforms (allows cross-compilation).
// Returns minimal/empty data — real data comes from the linux build.

func (l *LinuxSystemInfo) OS() (OSInfo, error) {
	return OSInfo{Name: "linux", Distro: "unknown (cross-compiled)"}, nil
}

func (l *LinuxSystemInfo) User() (UserInfo, error) {
	return UserInfo{Name: "unknown"}, nil
}

func (l *LinuxSystemInfo) Disk() (DiskInfo, error) {
	return DiskInfo{}, nil
}

func (l *LinuxSystemInfo) Memory() (MemoryInfo, error) {
	return MemoryInfo{}, nil
}
