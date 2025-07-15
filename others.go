//go:build !windows
// +build !windows

package main

import (
	"os/exec"
)

func removeTerminal(command *exec.Cmd) {
}
