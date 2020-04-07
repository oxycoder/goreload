package internal

import (
	"os/exec"
	"strconv"
)

func (r *runner) killDbg() error {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(r.dbg.Process.Pid))
	return kill.Run()
}

func (r *runner) killApp() error {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(r.cmd.Process.Pid))
	return kill.Run()
}
