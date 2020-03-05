package main

import (
	"fmt"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/oxycoder/goreload/internal"
	"github.com/radovskyb/watcher"
	"github.com/urfave/cli/v2"

	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"
)

var (
	startTime  = time.Now()
	logger     = log.New(os.Stdout, "[goreload] ", 0)
	colorGreen = string([]byte{27, 91, 57, 55, 59, 51, 50, 59, 49, 109})
	colorRed   = string([]byte{27, 91, 57, 55, 59, 51, 49, 59, 49, 109})
	colorReset = string([]byte{27, 91, 48, 109})
)

func main() {
	app := cli.NewApp()
	app.Name = "goreload"
	app.Usage = "A live reload utility for Go web applications."
	app.Action = mainAction
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "bin",
			Aliases: []string{"b"},
			Value:   "./bin/goreload",
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
			Aliases: []string{"t"},
			Value:   ".",
			Usage:   "Path to watch files from",
		},
		&cli.StringFlag{
			Name:    "build",
			Aliases: []string{"d"},
			Value:   "",
			Usage:   "Path to build files from (defaults to same value as --path)",
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
			Value: "goreload",
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
	}
	app.Commands = []*cli.Command{
		{
			Name:   "run",
			Usage:  "Run the goreload",
			Action: mainAction,
		},
	}

	app.Run(os.Args)
}

func mainAction(c *cli.Context) error {
	logPrefix := c.String("logPrefix")

	logger.SetPrefix(fmt.Sprintf("[%s] ", logPrefix))

	wd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}

	buildArgs, err := shellwords.Parse(c.String("buildArgs"))
	if err != nil {
		logger.Fatal(err)
	}

	buildPath := c.String("build")
	if buildPath == "" {
		buildPath = c.String("path")
	}
	builder := internal.NewBuilder(buildPath, c.String("bin"), wd, buildArgs)
	runner := internal.NewRunner(filepath.Join(wd, builder.Binary()), c.Args().Slice()...)
	runner.SetWriter(os.Stdout)

	shutdown(runner)

	// build right now
	build(builder, runner, logger)

	w := watcher.New()
	defer w.Close()
	w.IgnoreHiddenFiles(true)
	for _, x := range c.StringSlice("excludeDir") {
		w.Ignore(x)
	}

	filterPattern := fmt.Sprintf(`.*\.(%s)`, c.String("ext"))
	r := regexp.MustCompile(filterPattern)
	w.AddFilterHook(watcher.RegexFilterHook(r, false))
	if err := w.AddRecursive(c.String("path")); err != nil {
		log.Fatalln(err)
	}
	if c.Bool("showWatchedFiles") {
		for path, f := range w.WatchedFiles() {
			fmt.Printf("%s: %s\n", path, f.Name())
		}
		fmt.Println("--------------")
	}

	go func() {
		haveModified := false
		for {
			select {
			case event, ok := <-w.Event:
				if !ok {
					return
				}
				log.Println("modified file:", event.Name())
				haveModified = true

			case err, ok := <-w.Error:
				if !ok {
					return
				}
				log.Println("error:", err)

			case <-w.Closed:
				return

			case <-time.After(time.Millisecond * 200):
				if haveModified {
					runner.Kill()
					build(builder, runner, logger)
					haveModified = false
				}
			}
		}
	}()
	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 200); err != nil {
		log.Fatalln(err)
	}
	return err
}

func build(builder internal.Builder, runner internal.Runner, logger *log.Logger) {
	logger.Println("Building...")

	err := builder.Build()
	if err != nil {
		logger.Printf("%sBuild failed%s\n", colorRed, colorReset)
		fmt.Println(builder.Errors())
	} else {
		logger.Printf("%sBuild finished%s\n", colorGreen, colorReset)
		runner.Run()
	}

	time.Sleep(100 * time.Millisecond)
}

func shutdown(runner internal.Runner) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		log.Println("Got signal: ", s)
		err := runner.Kill()
		if err != nil {
			log.Print("Error killing: ", err)
		}
		os.Exit(1)
	}()
}
