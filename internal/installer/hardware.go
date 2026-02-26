// hardware.go provides system hardware detection for preflight checks.
// It reads CPU, memory, and disk information from the Linux proc and
// sys filesystems to compare against Sonar-recommended minimums.
package installer

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// HardwareReqs specifies the minimum recommended hardware for an installer.
type HardwareReqs struct {
	// MinCPUCores is the minimum number of CPU cores recommended.
	MinCPUCores int
	// MinRAMMB is the minimum RAM in megabytes recommended.
	MinRAMMB int
	// MinDiskGB is the minimum free disk space in gigabytes recommended.
	MinDiskGB int
}

// HardwareInfo holds detected system hardware specifications.
type HardwareInfo struct {
	// CPUCores is the number of logical CPU cores available.
	CPUCores int
	// RAMMB is the total system RAM in megabytes.
	RAMMB int
	// DiskFreeGB is the free disk space in gigabytes on the root filesystem.
	DiskFreeGB int
}

// DetectHardware reads system hardware information from the Linux host.
func DetectHardware() (*HardwareInfo, error) {
	return DetectHardwareWith(os.ReadFile, statfsFunc)
}

// StatfsFunc is a function that performs a statfs call on a path.
// Allows dependency injection for testing.
type StatfsFunc func(path string, buf *unix.Statfs_t) error

// statfsFunc is the default production statfs implementation.
func statfsFunc(path string, buf *unix.Statfs_t) error {
	return unix.Statfs(path, buf)
}

// DetectHardwareWith reads system hardware using injected dependencies.
// This allows tests to provide fake file contents and statfs results.
func DetectHardwareWith(readFile ReadFileFunc, statfs StatfsFunc) (*HardwareInfo, error) {
	info := &HardwareInfo{}

	// CPU cores from runtime (works cross-platform).
	info.CPUCores = runtime.NumCPU()

	// Total RAM from /proc/meminfo.
	memData, err := readFile("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("reading /proc/meminfo: %w", err)
	}
	info.RAMMB = parseMemTotalMB(string(memData))

	// Free disk space on root filesystem.
	var stat unix.Statfs_t
	if err := statfs("/", &stat); err != nil {
		return nil, fmt.Errorf("statfs /: %w", err)
	}
	// Available blocks * block size, converted to gigabytes.
	info.DiskFreeGB = int(stat.Bavail * uint64(stat.Bsize) / (1024 * 1024 * 1024))

	return info, nil
}

// parseMemTotalMB extracts the MemTotal value from /proc/meminfo
// content and returns it in megabytes.
func parseMemTotalMB(data string) int {
	for _, line := range strings.Split(data, "\n") {
		if !strings.HasPrefix(line, "MemTotal:") {
			continue
		}
		// Format: "MemTotal:       16384000 kB"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0
		}
		kb, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return 0
		}
		return int(kb / 1024)
	}
	return 0
}

// CheckHardware compares detected hardware against requirements and
// returns warning messages for any specs below the recommendation.
func CheckHardware(info *HardwareInfo, reqs *HardwareReqs) []string {
	var warnings []string

	if reqs.MinCPUCores > 0 && info.CPUCores < reqs.MinCPUCores {
		warnings = append(warnings, fmt.Sprintf(
			"CPU: %d cores detected, %d recommended", info.CPUCores, reqs.MinCPUCores))
	}

	if reqs.MinRAMMB > 0 && info.RAMMB < reqs.MinRAMMB {
		warnings = append(warnings, fmt.Sprintf(
			"RAM: %d MB detected, %d MB recommended", info.RAMMB, reqs.MinRAMMB))
	}

	if reqs.MinDiskGB > 0 && info.DiskFreeGB < reqs.MinDiskGB {
		warnings = append(warnings, fmt.Sprintf(
			"Disk: %d GB free, %d GB recommended", info.DiskFreeGB, reqs.MinDiskGB))
	}

	return warnings
}
