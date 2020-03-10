package main

import (
	"fmt"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/oxycoder/goreload/color"
	"github.com/oxycoder/goreload/internal"
	"github.com/radovskyb/watcher"
	"github.com/urfave/cli/v2"

	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

var (
	startTime = time.Now()
	logger    = log.New(os.Stdout, "[💕Go] ", 0)

	sha1ver   string
	buildTime string
)

func main() {
	app := cli.NewApp()
	app.Name = "goreload"
	app.Usage = fmt.Sprintf("A live reload utility for Go web applications, sha1: %s, build time: %s", sha1ver, buildTime)
	app.Action = mainAction
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "bin",
			Aliases: []string{"b"},
			Value:   "./bin/gorl",
			Usage:   "path to generated binary file",
		},
		&cli.StringFlag{
			Name:    "ext",
			Aliases: []string{"e"},
			Value:   `go|html`,
			Usage:   "File extention to watch changes",
		},
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Value:   ".",
			Usage:   "Path to watch files from",
		},
		&cli.StringSliceFlag{
			Name:    "excludeDir",
			Aliases: []string{"x"},
			Value:   cli.NewStringSlice("bin", ".git"),
			Usage:   "Relative directories to exclude",
		},
		&cli.StringFlag{
			Name:  "buildArgs",
			Usage: "Additional go build arguments",
		},
		&cli.StringFlag{
			Name:  "logPrefix",
			Usage: "Log prefix",
			Value: "",
		},
		&cli.Int64Flag{
			Name:  "delay",
			Usage: "Delay build after detect changes",
			Value: 400,
		},
		&cli.BoolFlag{
			Name:  "showWatchedFiles, swf",
			Usage: "Verbose output",
		},
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"dlv"},
			Usage:   "Start with debugger attached, require dlv installed, default address is :2345",
			Value:   false,
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:   "run",
			Usage:  "Run the goreload",
			Action: mainAction,
		},
		{
			Name:   "version",
			Usage:  "Version info",
			Action: verAction,
		},
	}

	app.Run(os.Args)
}

func verAction(c *cli.Context) error {
	logInfo(" -------------------GoReload------------------ ")
	logInfo("| build time: %-32v|", buildTime)
	logInfo("| sha1: %-38v|", sha1ver)
	logInfo(" ---------------------------------------------- ")
	return nil
}

func mainAction(c *cli.Context) error {
	verAction(c) // Display version on running
	if c.String("logPrefix") != "" {
		logger.SetPrefix(c.String("logPrefix"))
	}

	buildArgs, err := shellwords.Parse(c.String("buildArgs"))
	if err != nil {
		logger.Fatal(err)
	}

	sourcePath := c.String("path")
	// cd to source folder
	if err := os.Chdir(sourcePath); err != nil {
		logger.Fatal(err)
	}
	builder := internal.NewBuilder(".", c.String("bin"), c.Bool("debug"), buildArgs)
	runner := internal.NewRunner(c.String("bin"), c.Bool("debug"), c.Args().Slice()...)
	runner.SetWriter(os.Stdout)

	shutdown(runner)

	// build right now
	build(builder, runner, logger, c.Bool("debug"))

	w := watcher.New()
	defer w.Close()
	w.IgnoreHiddenFiles(true)
	for _, x := range c.StringSlice("excludeDir") {
		w.Ignore(x)
	}

	filterPattern := fmt.Sprintf(`.*\.(%s)`, c.String("ext"))
	r := regexp.MustCompile(filterPattern)
	w.AddFilterHook(watcher.RegexFilterHook(r, false))
	if err := w.AddRecursive("."); err != nil {
		logger.Fatalln(err)
	}
	if c.Bool("showWatchedFiles") {
		for path, f := range w.WatchedFiles() {
			fmt.Printf("%s: %s\n", path, f.Name())
		}
	}

	go func() {
		haveModified := false
		for {
			select {
			case event, ok := <-w.Event:
				if !ok {
					return
				}
				logInfo("modified file: %s", event.Name())
				haveModified = true

			case err, ok := <-w.Error:
				if !ok {
					return
				}
				logger.Println("error:", err)

			case <-w.Closed:
				return

			case <-time.After(time.Millisecond * 200):
				if haveModified {
					runner.Kill()
					build(builder, runner, logger, c.Bool("debug"))
					haveModified = false
				}
			}
		}
	}()
	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 200); err != nil {
		logger.Fatalln(err)
	}
	return err
}

func build(builder internal.Builder, runner internal.Runner, logger *log.Logger, isDebug bool) {
	logInfo("Building...")

	err := builder.Build()
	if err != nil {
		logError("Build failed")
		fmt.Println(builder.Errors())
	} else {
		logSuccess("Build finished")
		p, err := runner.Run()
		if err != nil {
			logger.Fatal(err)
		}

		logDebug(`Process started with (pid=%d)`, p.Process.Pid)
		if isDebug {
			_, err := runner.AttachDebugger()
			if err != nil {
				logger.Fatal(err)
			}
			logDebug("Debugger attached to pid=%d", p.Process.Pid)
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

func logSuccess(text string, args ...interface{}) {
	args = append([]interface{}{color.Green}, args...)
	args = append(args, color.Reset)
	logger.Printf("%s"+text+"%s", args...)
}

func logInfo(text string, args ...interface{}) {
	args = append([]interface{}{color.Cyan}, args...)
	args = append(args, color.Reset)
	logger.Printf("%s"+text+"%s", args...)
}

func logError(text string, args ...interface{}) {
	args = append([]interface{}{color.Red}, args...)
	args = append(args, color.Reset)
	logger.Printf("%s"+text+"%s", args...)
}

func logWarn(text string, args ...interface{}) {
	args = append([]interface{}{color.Yellow}, args...)
	args = append(args, color.Reset)
	logger.Printf("%s"+text+"%s", args...)
}

func logDebug(text string, args ...interface{}) {
	args = append([]interface{}{color.Magenta}, args...)
	args = append(args, color.Reset)
	logger.Printf("%s"+text+"%s", args...)
}
