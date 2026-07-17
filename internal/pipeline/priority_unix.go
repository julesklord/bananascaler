//go:build !windows

package pipeline

import (
	"syscall"
)

// setPriority sets the process priority (nice level) on Unix-like operating systems.
func setPriority(pid, niceLevel int) {
	_ = syscall.Setpriority(syscall.PRIO_PROCESS, pid, niceLevel)
}
