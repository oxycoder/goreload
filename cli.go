package main

import "github.com/urfave/cli/v2"

var (
	sha1ver   string
	buildTime string
)

func cliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "goreload"
	app.Usage = "A live reload utility for Go web applications"
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
			Value:   `.go|.html`,
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
			Value:   cli.NewStringSlice("bin", ".git", "node_modules"),
			Usage:   "Relative directories to exclude",
		},
		&cli.StringFlag{
			Name:  "buildArgs",
			Usage: "Additional go build arguments",
		},
		&cli.StringFlag{
			Name:  "runArgs",
			Usage: "Additional arguments when run app",
		},
		&cli.StringFlag{
			Name:  "logPrefix",
			Usage: "Log prefix",
			Value: "",
		},
		&cli.DurationFlag{
			Name:  "delay",
			Usage: "Delay build after detect changes",
			Value: 200,
		},
		&cli.BoolFlag{
			Name:    "showWatchedFiles",
			Aliases: []string{"swf"},
			Usage:   "Verbose output",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"dlv"},
			Usage:   "Start with debugger attached, require dlv installed",
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "dlvAddress",
			Aliases: []string{"da"},
			Usage:   "dlv debug port",
			Value:   ":2345",
		},
		&cli.StringFlag{
			Name:    "watcher",
			Aliases: []string{"w"},
			Usage:   "Choose default watcher, valid value is bwatch and fsnotify, default is fsnotify",
			Value:   "fsnotify",
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
	return app
}

func verAction(c *cli.Context) error {
	version()
	return nil
}

func version() {
	logInfo(`
---------------------
     GoReload
🎉version: 1.1.5 🎉
---------------------`)
}
