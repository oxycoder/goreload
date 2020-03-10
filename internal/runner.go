package internal

import (
	"io"
	"io/ioutil"
	"log"
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
	command   *exec.Cmd
	dbg       *exec.Cmd
	debug     bool
	starttime time.Time
}

// NewRunner creates new runner
func NewRunner(bin string, isDebug bool, args ...string) Runner {
	return &runner{
		bin:       bin,
		args:      args,
		writer:    ioutil.Discard,
		debug:     isDebug,
		starttime: time.Now(),
	}
}

func (r *runner) Run() (*exec.Cmd, error) {
	if r.needsRefresh() {
		r.Kill()
	}

	if r.command == nil || r.Exited() {
		err := r.runBin()
		if err != nil {
			log.Print("Error running: ", err)
		}
		time.Sleep(250 * time.Millisecond)
		return r.command, err
	}
	return r.command, nil
}

func (r *runner) Info() (os.FileInfo, error) {
	return os.Stat(r.bin)
}

func (r *runner) SetWriter(writer io.Writer) {
	r.writer = writer
}

func (r *runner) Kill() error {
	if r.command == nil {
		return nil
	}
	if r.command.Process == nil {
		return nil
	}

	done := make(chan error)
	go func() {
		r.command.Wait()
		if r.debug {
			r.dbg.Wait()
		}
		close(done)
	}()

	// trying a "soft" kill first
	if runtime.GOOS == "windows" {
		if err := r.command.Process.Kill(); err != nil {
			return err
		}
	} else {
		err := r.command.Process.Signal(os.Interrupt)
		if err != nil {
			return err
		}
	}
	if r.debug {
		err := r.dbg.Process.Kill()
		if err != nil {
			return err
		}
	}

	// wait for our process to die before we return or hard kill after 3 sec
	select {
	case <-time.After(3 * time.Second):
		if err := r.command.Process.Kill(); err != nil {
			return err
		}
	case <-done:
	}
	r.command = nil

	return nil
}

func (r *runner) Exited() bool {
	return r.command != nil && r.command.ProcessState != nil && r.command.ProcessState.Exited()
}

func (r *runner) runBin() error {
	r.command = exec.Command(r.bin, r.args...)
	stdout, err := r.command.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := r.command.StderrPipe()
	if err != nil {
		return err
	}

	err = r.command.Start()
	if err != nil {
		return err
	}

	r.starttime = time.Now()

	go io.Copy(r.writer, stdout)
	go io.Copy(r.writer, stderr)
	go r.command.Wait()

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
		"--accept-multiclient",
		"--listen=:2435",
		"--headless=true",
		"--api-version=2",
		"--log",
		strconv.Itoa(r.command.Process.Pid),
	)
	log.Printf("DLV attaching to process (pid=%d)", r.command.Process.Pid)

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
