package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"strings"

	"github.com/fsnotify/fsnotify"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/oxycoder/goreload/internal"
	"github.com/urfave/cli"

	"log"
	"os"
	"os/signal"
	"path/filepath"
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
		cli.StringFlag{
			Name:  "bin,b",
			Value: "./bin/goreload",
			Usage: "path to generated binary file",
		},
		cli.StringSliceFlag{
			Name:  "ext,e",
			Value: &cli.StringSlice{".go"},
			Usage: "File extention to watch changes",
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
		cli.StringFlag{
			Name:  "buildArgs",
			Usage: "Additional go build arguments",
		},
		cli.StringFlag{
			Name:  "logPrefix",
			Usage: "Log prefix",
			Value: "goreload",
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
	// all := c.GlobalBool("all")
	logPrefix := c.GlobalString("logPrefix")

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
	builder := internal.NewBuilder(buildPath, c.GlobalString("bin"), wd, buildArgs)
	runner := internal.NewRunner(filepath.Join(wd, builder.Binary()), c.Args()...)
	runner.SetWriter(os.Stdout)

	shutdown(runner)

	// build right now
	build(builder, runner, logger)
	log.Println(runtime.GOOS)
	if runtime.GOOS == "windows" {
		// scan for changes
		scanChanges(c.GlobalString("path"), c.GlobalStringSlice("excludeDir"), func() {
			runner.Kill()
			build(builder, runner, logger)
		})
	} else {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()
		dirs, err := getAllDir(c.GlobalString("path"))
		if err != nil {
			log.Fatal(err)
		}
		for _, dir := range dirs {
			if strings.Contains(dir, ".git") {
				continue
			}
			for _, x := range c.GlobalStringSlice("excludeDir") {
				if x == dir {
					continue
				}
			}
			log.Println("watching dir: ", dir)
			watcher.Add(dir)
		}
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// check file extentions
				if containExt(path.Ext(event.Name), c.GlobalStringSlice("ext")) {
					switch event.Op {
					case fsnotify.Create:
						if isDir(event.Name) {
							log.Println("added dir: ", event.Name, " to watch list")
							watcher.Add(event.Name)
						}
					case fsnotify.Chmod:

					default:
						log.Println("modified file:", event.Name)
						runner.Kill()
						build(builder, runner, logger)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}

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

func getAllDir(pathname string) ([]string, error) {
	var allDir []string
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		log.Print("read dir fail:", err)
		return allDir, err
	}

	allDir = append(allDir, pathname)

	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			subdir, err := getAllDir(fullDir)
			allDir = append(allDir, subdir...)
			if err != nil {
				log.Print("read dir fail:", err)
				return allDir, err
			}
		}
	}
	return allDir, nil
}

func scanChanges(watchPath string, excludeDirs []string, cb func()) {
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

			if (filepath.Ext(path) == ".go") && info.ModTime().After(startTime) {
				cb()
				startTime = time.Now()
				return errors.New("done")
			}

			return nil
		})
		time.Sleep(500 * time.Millisecond)
	}
}

func isDir(path string) bool {
	absPath, _ := filepath.Abs(path)
	fi, err := os.Stat(absPath)
	if err != nil {
		log.Print(err)
	}
	return fi.IsDir()
}

func containExt(ext string, exts []string) bool {
	for _, a := range exts {
		if a == ext {
			return true
		}
	}
	return false
}
