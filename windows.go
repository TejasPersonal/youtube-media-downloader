//go:build windows
// +build windows

package main

import (
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
)

func removeTerminal(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
}

func GetVideoFolder() string {
	path, _ := windows.KnownFolderPath(windows.FOLDERID_Videos, windows.KF_FLAG_DEFAULT)
	return filepath.Join(path, "Youtube")
}
