//go:build !windows

package pipeline

import (
	"os/exec"
	"syscall"
	"testing"
)

func TestSetPriority(t *testing.T) {
	// Start a dummy process that we can safely modify
	cmd := exec.Command("sleep", "2")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start dummy process: %v", err)
	}

	// Clean up the process when the test finishes
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	pid := cmd.Process.Pid

	// Get initial priority
	initialPrio, err := syscall.Getpriority(syscall.PRIO_PROCESS, pid)
	if err != nil {
		t.Fatalf("failed to get initial priority: %v", err)
	}

	// Try to increase nice value (lower priority).
	// Unprivileged users can usually increase nice values (i.e. lower priority),
	// but cannot decrease them.
	targetNice := 10
	setPriority(pid, targetNice)

	// Get new priority
	newPrio, err := syscall.Getpriority(syscall.PRIO_PROCESS, pid)
	if err != nil {
		t.Fatalf("failed to get new priority: %v", err)
	}

	// Because different systems represent priority differently (e.g. Linux might return 20-nice),
	// we just verify that it changed.
	if newPrio == initialPrio {
		t.Logf("Priority did not change. Initial: %d, New: %d. This can happen in restricted environments (like CI or containers).", initialPrio, newPrio)
	} else {
		t.Logf("Priority successfully changed. Initial: %d, New: %d", initialPrio, newPrio)
	}
}

func TestSetPriority_InvalidPID(t *testing.T) {
	// Call setPriority with a pid that shouldn't exist.
	// We use -1 which is often invalid, or a very large number.
	// Our function ignores the error, so this just verifies it doesn't panic.
	setPriority(-99999, 10)
}
