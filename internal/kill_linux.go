package internal

import (
	"os/exec"
	"syscall"
)

func (r *runner) killDbg() error {
	pid := r.dbg.Process.Pid
	err := r.dbg.Process.Kill()
	if err == nil {
		r.log.Println("Killing dbg pid=", pid)
		return err
	}
	r.log.Println("Soft kill err ", err.Error())
	r.log.Println("Try kill dbg with SIGKILL pid=", pid)
	err = syscall.Kill(-pid, syscall.SIGKILL)
	r.log.Println(err)
	return err
}

func (r *runner) killApp() error {
	if r.debug {
		// ensure child process is killed
		exec.Command("bash", "-c", "pkill "+r.bin).Output()
	}
	pid := r.cmd.Process.Pid
	if err := r.cmd.Process.Kill(); err == nil {
		r.log.Println("Killing app pid=", pid)
		return err
	}
	r.log.Println("Try kill app with SIGKILL pid=", pid)
	err := syscall.Kill(-pid, syscall.SIGKILL)
	return err
}
