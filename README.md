# Goreload

========

`goreload` is a simple command line utility for live-reloading Go web applications.
Just run `goreload` in your app directory.
`goreload` will automatically recompile your code when it
detects a change.

`goreload` adheres to the "silence is golden" principle, so it will only complain
if there was a compiler error or if you succesfully compile after an error.

## Installation

Assuming you have a working Go environment and `GOPATH/bin` is in your
`PATH`, `goreload` is a breeze to install:

```shell
go get github.com/acoshift/goreload
```

Then verify that `goreload` was installed correctly:

```shell
goreload -h
```

## Basic usage

```shell
goreload run main.go
```

Options

```txt
   --bin value, -b value         name of generated binary file (default: ".goreload")
   --path value, -t value        Path to watch files from (default: ".")
   --build value, -d value       Path to build files from (defaults to same value as --path)
   --excludeDir value, -x value  Relative directories to exclude
   --all                         reloads whenever any file changes, as opposed to reloading only on .go file change
   --godep, -g                   use godep when building
   --buildArgs value             Additional go build arguments
   --logPrefix value             Setup custom log prefix
   --notifications               enable desktop notifications
   --help, -h                    show help
   --version, -v                 print the version
```
