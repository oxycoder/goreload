package internal

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Builder interface
type Builder interface {
	Build() error
	Errors() string
}

type builder struct {
	dir       string
	binary    string
	errors    string
	buildArgs []string
	debug     bool
}

// NewBuilder creates new builder
func NewBuilder(sourcePath string, bin string, isDebug bool, buildArgs []string) Builder {
	if len(bin) == 0 {
		bin = "bin"
	}

	// does not work on Windows without the ".exe" extension
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(bin, ".exe") { // check if it already has the .exe extension
			bin += ".exe"
		}
	}

	return &builder{dir: sourcePath, binary: bin, debug: isDebug, buildArgs: buildArgs}
}

func (b *builder) Errors() string {
	return b.errors
}

func (b *builder) Build() error {
	args := append([]string{"go", "build", "-o", b.binary})
	if b.debug {
		args = append(args, `-gcflags="all=-N -l"`)
	}
	args = append(args, b.buildArgs...)
	args = append(args, b.dir)

	command := exec.Command(args[0], args[1:]...)

	output, err := command.CombinedOutput()
	if command.ProcessState.Success() {
		b.errors = ""
	} else {
		b.errors = string(output) + err.Error()
	}

	if len(b.errors) > 0 {
		return fmt.Errorf(b.errors)
	}

	return err
}
