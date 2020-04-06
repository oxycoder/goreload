package internal

import (
	"os/exec"
	"syscall"
	"time"
)

func kill(cmd *exec.Cmd) error {
	pid := cmd.Process.Pid
	if err := syscall.Kill(-pid, syscall.SIGINT); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	err := syscall.Kill(-pid, syscall.SIGKILL)
	_, err = cmd.Process.Wait()
	return err
}
