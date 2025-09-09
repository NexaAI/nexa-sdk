//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

func main() {
	cmd := exec.Command("cmd", "/c", "start", "powershell", "-NoProfile", "-NoExit", "-Command", "nexa")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Start()
}
