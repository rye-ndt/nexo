//go:build !windows

package harness_helper

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func SetProcAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func KillProc(cmd *exec.Cmd) error {
	tree := descendants(cmd.Process.Pid)

	killErr := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if killErr != nil {
		if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			killErr = err
		} else {
			killErr = nil
		}
	}

	for _, pid := range tree {
		syscall.Kill(pid, syscall.SIGKILL)
	}

	return killErr
}

// A child spawned with its own session — what Claude Code's Bash tool does — leaves
// the process group and reparents to init the moment its parent dies, so the tree is
// read before anything is signalled.
func descendants(root int) []int {
	if root <= 1 {
		return nil
	}

	out, err := exec.Command("ps", "-Ao", "pid=,ppid=").Output()
	if err != nil {
		return nil
	}

	children := map[int][]int{}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}

		pid, pidErr := strconv.Atoi(fields[0])
		ppid, ppidErr := strconv.Atoi(fields[1])
		if pidErr != nil || ppidErr != nil || pid <= 1 {
			continue
		}

		children[ppid] = append(children[ppid], pid)
	}

	tree := []int{}

	for queue := []int{root}; len(queue) > 0; queue = queue[1:] {
		for _, child := range children[queue[0]] {
			tree = append(tree, child)
			queue = append(queue, child)
		}
	}

	return tree
}
