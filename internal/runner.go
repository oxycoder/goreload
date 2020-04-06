package internal

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

// Runner interface
type Runner interface {
	Run() (*exec.Cmd, error)
	Info() (os.FileInfo, error)
	SetWriter(io.Writer)
	Kill() error
	AttachDebugger() (*exec.Cmd, error)
}

type runner struct {
	bin       string
	args      []string
	writer    io.Writer
	cmd       *exec.Cmd
	dbg       *exec.Cmd
	debug     bool
	dlvAddr   string
	starttime time.Time
}

// NewRunner creates new runner
func NewRunner(bin string, isDebug bool, addr string, args ...string) Runner {
	return &runner{
		bin:       bin,
		args:      args,
		writer:    ioutil.Discard,
		debug:     isDebug,
		dlvAddr:   addr,
		starttime: time.Now(),
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

func (r *runner) Info() (os.FileInfo, error) {
	return os.Stat(r.bin)
}

func (r *runner) SetWriter(writer io.Writer) {
	r.writer = writer
}

func (r *runner) Kill() error {
	if r.cmd == nil {
		return nil
	}
	if r.cmd.Process == nil {
		return nil
	}

	done := make(chan error)
	go func() {
		r.cmd.Wait()
		if r.debug {
			r.dbg.Wait()
		}
		close(done)
	}()

	if r.debug {
		err := kill(r.dbg)
		if err != nil {
			return err
		}
	}
	// trying a "soft" kill first
	if runtime.GOOS == "windows" {
		if err := r.cmd.Process.Kill(); err != nil {
			return err
		}
	} else {
		err := r.cmd.Process.Signal(os.Interrupt)
		if err != nil {
			return err
		}
	}

	// wait for our process to die before we return or hard kill after 3 sec
	select {
	case <-time.After(3 * time.Second):
		if err := kill(r.cmd); err != nil {
			return err
		}
	case <-done:
	}
	r.cmd = nil

	return nil
}

func (r *runner) Exited() bool {
	return r.cmd != nil && r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited()
}

func (r *runner) runBin() error {
	r.cmd = exec.Command(r.bin, r.args...)
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

func (r *runner) AttachDebugger() (*exec.Cmd, error) {
	r.dbg = exec.Command(
		"dlv",
		"attach",
		// "--continue", // Do not pause process on attach
		"--accept-multiclient",
		"--listen="+r.dlvAddr,
		"--headless=true",
		"--api-version=2",
		strconv.Itoa(r.cmd.Process.Pid),
	)

	stdout, err := r.dbg.StdoutPipe()
	if err != nil {
		return r.dbg, err
	}
	stderr, err := r.dbg.StderrPipe()
	if err != nil {
		return r.dbg, err
	}

	err = r.dbg.Start()
	if err != nil {
		return r.dbg, err
	}

	r.starttime = time.Now()

	go io.Copy(r.writer, stdout)
	go io.Copy(r.writer, stderr)
	go r.dbg.Wait()
	return r.dbg, err
}
