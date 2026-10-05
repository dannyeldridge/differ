//go:build unix

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// detach re-launches this binary in its own session with --foreground, output
// going to a log file, and returns without waiting so the shell is freed. Later
// launches find the running instance via the single-instance lock.
func detach(root string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	logFile, err := openLog()
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd := exec.Command(exe, "--foreground")
	cmd.Dir = root
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start differ: %w", err)
	}
	return cmd.Process.Release()
}

func openLog() (*os.File, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(dir, "differ")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(filepath.Join(dir, "differ.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}
