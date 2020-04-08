package internal

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Runner interface
type Runner interface {
	Run() (*exec.Cmd, error)
	Info() (os.FileInfo, error)
	SetWriter(io.Writer)
	Kill() error
}

type runner struct {
	bin       string
	args      []string
	writer    io.Writer
	cmd       *exec.Cmd
	dbg       *exec.Cmd
	debug     bool
	dlvAddr   string
	dlvWriter io.WriteCloser
	starttime time.Time
	log       *log.Logger
}

// NewRunner creates new runner
func NewRunner(bin string, isDebug bool, addr string, log *log.Logger, args ...string) Runner {
	return &runner{
		bin:       bin,
		args:      args,
		writer:    ioutil.Discard,
		debug:     isDebug,
		dlvAddr:   addr,
		starttime: time.Now(),
		log:       log,
	}
}

func (r *runner) Run() (*exec.Cmd, error) {
	if r.needsRefresh() {
		r.Kill()
	}

	if r.cmd == nil || r.Exited() {
		err := r.runBin()
		time.Sleep(250 * time.Millisecond)
		return r.cmd, err
	}
	return r.cmd, nil
}

func (r *runner) Kill() error {
	// if r.debug && r.dbg != nil {
	// 	r.killDbg()
	// 	r.dbg = nil
	// }
	if r.cmd != nil {
		r.killApp()
		r.cmd = nil
	}
	return nil
}

func (r *runner) Info() (os.FileInfo, error) {
	return os.Stat(r.bin)
}

func (r *runner) SetWriter(writer io.Writer) {
	r.writer = writer
}

func (r *runner) Exited() bool {
	return r.cmd != nil && r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited()
}

func (r *runner) runBin() error {
	r.cmd = exec.Command(r.bin, r.args...)
	if r.debug {
		args := []string{
			"exec",
			"--continue",
			"--accept-multiclient",
			"--listen=" + r.dlvAddr,
			"--headless=true",
			"--api-version=2",
			r.bin,
			"--",
		}
		args = append(args, r.args...)
		r.cmd = exec.Command("dlv", args...)
		r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	stdout, err := r.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := r.cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = r.cmd.Start()
	if err != nil {
		return err
	}

	r.starttime = time.Now()

	go io.Copy(r.writer, stdout)
	go io.Copy(r.writer, stderr)
	go r.cmd.Wait()

	return nil
}

func (r *runner) needsRefresh() bool {
	info, err := r.Info()
	if err != nil {
		return false
	}
	return info.ModTime().After(r.starttime)
}
