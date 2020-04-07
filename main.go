package main

import (
	"fmt"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/oxycoder/goreload/internal"
	"github.com/urfave/cli/v2"

	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var logger = log.New(os.Stdout, "[💕Go] ", 0)

const ver = "v1.1.8"

func main() {
	version()
	app := cliApp()
	app.Run(os.Args)
}

func mainAction(c *cli.Context) error {
	if c.String("logPrefix") != "" {
		logger.SetPrefix(c.String("logPrefix"))
	}

	sourcePath := c.String("path")
	// cd to source folder
	if err := os.Chdir(sourcePath); err != nil {
		logger.Fatal(err)
	}
	buildArgs, err := shellwords.Parse(c.String("buildArgs"))
	if err != nil {
		logger.Fatal(err)
	}
	if c.Bool("debug") {
		logInfo("dlv enabled on address %s", c.String("dlvAddr"))
	}
	builder := internal.NewBuilder(".", c.String("bin"), c.Bool("debug"), buildArgs)
	runArgs, err := shellwords.Parse(c.String("runArgs"))
	if err != nil {
		logger.Fatal(err)
	}
	runner := internal.NewRunner(c.String("bin"), c.Bool("debug"), c.String("dlvAddr"), logger, runArgs...)
	runner.SetWriter(os.Stdout)

	shutdown(runner)

	// build right now
	start(builder, runner, logger, c.Bool("debug"))
	if c.String("watcher") == "fsnotify" {
		fwatcher(c, runner, builder)
	} else {
		kwatch(c, runner, builder)
	}
	return err
}

func start(builder internal.Builder, runner internal.Runner, logger *log.Logger, isDebug bool) {
	logWarn("Building...")

	err := builder.Build()
	if err != nil {
		logError("Build failed")
		fmt.Println(builder.Errors())
	} else {
		logSuccess("Build finished")
		p, err := runner.Run()
		if err != nil {
			logError(err.Error())
			return
		}

		logDebug(`Process started with (pid=%d)`, p.Process.Pid)
		if isDebug {
			dlv, err := runner.AttachDebugger()
			if err != nil {
				logger.Fatal(err)
			}
			logDebug("Dlv (pid=%d) attached to pid=%d", dlv.Process.Pid, p.Process.Pid)
		}
	}

	time.Sleep(100 * time.Millisecond)
}

func shutdown(runner internal.Runner) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		logger.Println("Got signal: ", s)
		err := runner.Kill()
		if err != nil {
			logger.Println("Error killing process: ", err)
		}
		os.Exit(1)
	}()
}

func version() {
	logInfo(`
---------------------
     GoReload
🎉 version: %s 🎉
---------------------`, ver)
}
