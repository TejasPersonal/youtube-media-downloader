//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func removeTerminal(command *exec.Cmd) {
}

func GetVideoFolder() string {
	home_path, _ := os.UserHomeDir()
	return filepath.Join(home_path, "Videos", "Youtube")
}
