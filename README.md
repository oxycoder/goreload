# Goreload

`goreload` forks from codegangsta/gin and remove unused features.

Just run `goreload` in your app directory.
`goreload` will automatically recompile your code when it
detects a change.

goreload using fsnotify (Cross-platform file system notifications for Go)

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
   --bin value, -b value         Path to generated binary file (default: "./bin/goreload")
   --path value, -t value        Path to watch files from (default: ".")
   --build value, -d value       Path to build files from (defaults to same value as --path)
   --ext value, -e value         File extention to watch changes (default: .go)
   --excludeDir value, -x value  Relative directories to exclude
   --buildArgs value             Additional go build arguments
   --logPrefix value             Setup custom log prefix
   --help, -h                    show help
   --version, -v                 print the version
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