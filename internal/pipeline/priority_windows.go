//go:build windows

package pipeline

import (
	"syscall"
)

const (
	// Scheduling priorities: https://learn.microsoft.com/en-us/windows/win32/procthread/scheduling-priorities
	idlePriorityClass        = 0x00000040
	belowNormalPriorityClass = 0x00004000
	normalPriorityClass      = 0x00000020
)

// setPriority sets the process priority class on Windows.
func setPriority(pid, niceLevel int) {
	// PROCESS_SET_INFORMATION = 0x0200
	handle, err := syscall.OpenProcess(0x0200, false, uint32(pid))
	if err != nil {
		return
	}
	defer syscall.CloseHandle(handle)

	// Map niceLevel (1-19) to Windows priority classes
	var priorityClass uintptr = normalPriorityClass
	if niceLevel >= 15 {
		priorityClass = idlePriorityClass
	} else if niceLevel >= 5 {
		priorityClass = belowNormalPriorityClass
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setPriorityClass := kernel32.NewProc("SetPriorityClass")
	_, _, _ = setPriorityClass.Call(uintptr(handle), priorityClass)
}
