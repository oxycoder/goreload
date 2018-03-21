package main

import (
	"errors"
	"fmt"

	"github.com/acoshift/goreload/internal"
	shellwords "github.com/mattn/go-shellwords"
	"gopkg.in/urfave/cli.v1"

	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/0xAX/notificator"
)

var (
	startTime     = time.Now()
	logger        = log.New(os.Stdout, "[goreload] ", 0)
	buildError    error
	colorGreen    = string([]byte{27, 91, 57, 55, 59, 51, 50, 59, 49, 109})
	colorRed      = string([]byte{27, 91, 57, 55, 59, 51, 49, 59, 49, 109})
	colorReset    = string([]byte{27, 91, 48, 109})
	notifier      = notificator.New(notificator.Options{AppName: "Go Reload Build"})
	notifications = false
)

func main() {
	app := cli.NewApp()
	app.Name = "goreload"
	app.Usage = "A live reload utility for Go web applications."
	app.Action = mainAction
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "bin,b",
			Value: ".goreload",
			Usage: "name of generated binary file",
		},
		cli.StringFlag{
			Name:  "path,t",
			Value: ".",
			Usage: "Path to watch files from",
		},
		cli.StringFlag{
			Name:  "build,d",
			Value: "",
			Usage: "Path to build files from (defaults to same value as --path)",
		},
		cli.StringSliceFlag{
			Name:  "excludeDir,x",
			Value: &cli.StringSlice{},
			Usage: "Relative directories to exclude",
		},
		cli.BoolFlag{
			Name:  "all",
			Usage: "reloads whenever any file changes, as opposed to reloading only on .go file change",
		},
		cli.BoolFlag{
			Name:  "godep,g",
			Usage: "use godep when building",
		},
		cli.StringFlag{
			Name:  "buildArgs",
			Usage: "Additional go build arguments",
		},
		cli.StringFlag{
			Name:  "logPrefix",
			Usage: "Log prefix",
			Value: "goreload",
		},
		cli.BoolFlag{
			Name:  "notifications",
			Usage: "Enables desktop notifications",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:      "run",
			ShortName: "r",
			Usage:     "Run the goreload",
			Action:    mainAction,
		},
	}

	app.Run(os.Args)
}

func mainAction(c *cli.Context) {
	all := c.GlobalBool("all")
	logPrefix := c.GlobalString("logPrefix")
	notifications = c.GlobalBool("notifications")

	logger.SetPrefix(fmt.Sprintf("[%s] ", logPrefix))

	wd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}

	buildArgs, err := shellwords.Parse(c.GlobalString("buildArgs"))
	if err != nil {
		logger.Fatal(err)
	}

	buildPath := c.GlobalString("build")
	if buildPath == "" {
		buildPath = c.GlobalString("path")
	}
	builder := internal.NewBuilder(buildPath, c.GlobalString("bin"), c.GlobalBool("godep"), wd, buildArgs)
	runner := internal.NewRunner(filepath.Join(wd, builder.Binary()), c.Args()...)
	runner.SetWriter(os.Stdout)

	shutdown(runner)

	// build right now
	build(builder, runner, logger)

	// scan for changes
	scanChanges(c.GlobalString("path"), c.GlobalStringSlice("excludeDir"), all, func(path string) {
		runner.Kill()
		build(builder, runner, logger)
	})
}

func build(builder internal.Builder, runner internal.Runner, logger *log.Logger) {
	logger.Println("Building...")

	if notifications {
		notifier.Push("Build Started!", "Building "+builder.Binary()+"...", "", notificator.UR_NORMAL)
	}
	err := builder.Build()
	if err != nil {
		buildError = err
		logger.Printf("%sBuild failed%s\n", colorRed, colorReset)
		fmt.Println(builder.Errors())
		buildErrors := strings.Split(builder.Errors(), "\n")
		if notifications {
			if err := notifier.Push("Build FAILED!", buildErrors[1], "", notificator.UR_CRITICAL); err != nil {
				logger.Println("Notification send failed")
			}
		}
	} else {
		buildError = nil
		logger.Printf("%sBuild finished%s\n", colorGreen, colorReset)
		runner.Run()

		if notifications {
			if err := notifier.Push("Build Succeded", "Build Finished!", "", notificator.UR_CRITICAL); err != nil {
				logger.Println("Notification send failed")
			}
		}
	}

	time.Sleep(100 * time.Millisecond)
}

type scanCallback func(path string)

func scanChanges(watchPath string, excludeDirs []string, allFiles bool, cb scanCallback) {
	for {
		filepath.Walk(watchPath, func(path string, info os.FileInfo, err error) error {
			if path == ".git" && info.IsDir() {
				return filepath.SkipDir
			}
			for _, x := range excludeDirs {
				if x == path {
					return filepath.SkipDir
				}
			}

			// ignore hidden files
			if filepath.Base(path)[0] == '.' {
				return nil
			}

			if (allFiles || filepath.Ext(path) == ".go") && info.ModTime().After(startTime) {
				cb(path)
				startTime = time.Now()
				return errors.New("done")
			}

			return nil
		})
		time.Sleep(500 * time.Millisecond)
	}
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
