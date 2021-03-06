# Goreload

`goreload` forks from codegangsta/gin and remove unused features.

Just run `goreload` in your app directory.
`goreload` will automatically recompile your code when it
detects a change.

goreload using fsnotify as default tool to detect changes, radovskyb/watcher can be configured by args.

## Installation

```shell
go get github.com/oxycoder/goreload
```

## Basic usage

```shell
# auto look for main.go
goreload
# run abc.go file
goreload abc.go
# watch .tmpl and .css extension, .go is default
goreload --ext .tmpl --ext .scss
# store binary file in path
goreload --bin ./bin/myproject
```

## Options

```txt
   --bin value, -b value         string      Path to generated binary file (default: "./bin/gorl")
   --path value, -t value        string      Path to watch files from (default: ".")
   --ext value, -e value         string      File extention to watch changes, seperate by `|` character, default `.go|.html`
   --excludeDir value, -x value  []string    Relative directories to exclude
   --buildArgs value             string      Arguments passed to `go build` command
   --runArgs value               string      Arguments passed when run program
   --logPrefix value             string      Setup custom log prefix
   --delay                       int         Delay build after detect files change, default value is 400
   --help, -h                    bool        show help
   --version, -v                 string      print the version
   --showWatchedFiles, -swf      bool        Print watched files
   --debug, -dlv                 bool        Enable debug
   --dlvAddr                     string      dlv server address, default :2345
   --watcher                     string      Watcher, default `fsnotify`, available: `fsnotify` and `bwatcher`
```

### Use with Docker

```Dockerfile
# Dockerfile
FROM golang:alpine as build
RUN apk update && apk add --no-cache git
EXPOSE 8000
# for live reload
RUN go get github.com/oxycoder/goreload

WORKDIR /myapp
ENTRYPOINT goreload --bin ./bin/myapp
```

```yml
# docker-compose.yml
version: '3'
services:
  web:
    build:
      context: .
      dockerfile: Dockerfile # build image from Dockerfile
    volumes:
      - ./:/myapp 
      - $GOPATH/pkg/mod/cache:/go/pkg/mod/cache # mount host go cache folder to speed up download
    working_dir: /myapp  
    ports:
      - "8000:8000"
```

### Build with version
```sh
./build.sh
```