//go:build !windows

package harness_helper

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestKillProcPrunesChildThatLeftTheProcessGroup(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 is not available")
	}

	cmd := exec.Command("bash", "-c",
		`python3 -c "import os,time; os.setsid(); print(os.getpid(), flush=True); time.sleep(120)"; sleep 120`)
	SetProcAttrs(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	grandchild := 0
	buf := make([]byte, 64)

	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); {
		n, readErr := stdout.Read(buf)
		if n > 0 {
			grandchild, err = strconv.Atoi(strings.TrimSpace(string(buf[:n])))
			if err != nil {
				t.Fatalf("read grandchild pid: %v", err)
			}
			break
		}
		if readErr != nil {
			t.Fatalf("read grandchild pid: %v", readErr)
		}
	}

	if grandchild == 0 {
		t.Fatal("grandchild never reported its pid")
	}

	defer syscall.Kill(grandchild, syscall.SIGKILL)

	if err := KillProc(cmd); err != nil {
		t.Fatalf("kill: %v", err)
	}

	cmd.Wait()

	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		if err := syscall.Kill(grandchild, 0); err != nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("grandchild %d survived KillProc", grandchild)
}
