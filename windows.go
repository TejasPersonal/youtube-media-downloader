//go:build windows
// +build windows

package main

import (
	"os/exec"
	"syscall"
)

func removeTerminal(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
}
