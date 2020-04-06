package internal

import (
	"os/exec"
	"strconv"
)

func kill(cmd *exec.Cmd) error {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	return kill.Run()
}
