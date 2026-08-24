//go:build windows

package harness_helper

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
)

func SetProcAttrs(cmd *exec.Cmd) {}

func KillProc(cmd *exec.Cmd) error {
	pid := strconv.Itoa(cmd.Process.Pid)

	if err := exec.Command("taskkill", "/T", "/F", "/PID", pid).Run(); err == nil {
		return nil
	}

	if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}

	return nil
}
