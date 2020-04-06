package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/oxycoder/goreload/internal"
	"github.com/radovskyb/watcher"
	"github.com/urfave/cli/v2"
)

// a fallback watcher for docker in windows
func kwatch(c *cli.Context, runner internal.Runner, builder internal.Builder) {
	w := watcher.New()
	defer w.Close()
	w.IgnoreHiddenFiles(true)
	for _, x := range c.StringSlice("excludeDir") {
		w.Ignore(x)
	}

	filterPattern := fmt.Sprintf(`.*(%s)`, c.String("ext"))
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

			case <-time.After(time.Millisecond * c.Duration("delay")):
				if haveModified {
					runner.Kill()
					start(builder, runner, logger, c.Bool("debug"))
					haveModified = false
				}
			}
		}
	}()
	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 200); err != nil {
		logger.Fatalln(err)
	}
}

// fsnotify watcher
func fwatcher(c *cli.Context, runner internal.Runner, builder internal.Builder) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatal(err)
	}
	defer w.Close()
	dirs, err := getAllDir(".", c.StringSlice("excludeDir")...)
	if err != nil {
		logger.Fatal(err)
	}
	for _, dir := range dirs {
		logInfo("watching dir: %s", dir)
		w.Add(dir)
	}
	exts := strings.Split(c.String("ext"), "|")
	haveModified := false
	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			// check file extentions
			if containExt(path.Ext(event.Name), exts) {
				switch event.Op {
				case fsnotify.Create:
					if isDir(event.Name) {
						if c.Bool("showWatchedFiles") {
							logSuccess("%s added to watch list", event.Name)
						}
						w.Add(event.Name)
					}
					haveModified = true
				case fsnotify.Remove:
					if isDir(event.Name) {
						w.Remove(event.Name)
					}
					haveModified = true
				case fsnotify.Chmod:
				default:
					logInfo("modified file: %s", event.Name)
					haveModified = true
				}
			}

		case <-time.After(time.Millisecond * c.Duration("delay")):
			if haveModified {
				runner.Kill()
				start(builder, runner, logger, c.Bool("debug"))
				haveModified = false
			}

		case err, ok := <-w.Errors:
			haveModified = false
			logError("error: %s", err)
			if !ok {
				return
			}
		}
	}
}

func getAllDir(pathname string, excludes ...string) ([]string, error) {
	var allDir []string
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		logError("read dir fail: %s", err)
		return allDir, err
	}

	allDir = append(allDir, pathname)

	for _, fi := range rd {
		if !fi.IsDir() {
			continue
		}
		fullDir := pathname + "/" + fi.Name()
		skip := false
		for _, ex := range excludes {
			if strings.Contains(fullDir, ex) {
				skip = true
			}
		}
		if skip {
			continue
		}
		subdir, err := getAllDir(fullDir)
		if err != nil {
			logError("read dir fail: %s", err)
			return allDir, err
		}
		allDir = append(allDir, subdir...)
	}
	return allDir, nil
}

func isDir(path string) bool {
	absPath, _ := filepath.Abs(path)
	fi, err := os.Stat(absPath)
	if err != nil {
		logError(err.Error())
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
