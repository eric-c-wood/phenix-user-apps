SHELL := /bin/bash

# Default version number to git commit hash if not set.
VER     := $(or $(VER),$(shell git log -1 --format="%h"))
COMMIT  := $(or $(COMMIT),$(shell git log -1 --format="%h - %ae"))
DATE    := $(shell date -u)
VERSION := $(VER) (commit $(COMMIT)) $(DATE)

PREFIX := $(shell find $(GOPATH) -name *.yml | grep api/config/default | cut -d $$'\n' -f 1)
PREFIX := $(shell echo $(PREFIX) | sed 's/[/]api[^\\\n\]\+//')

GOSOURCES   := $(shell find . \( -name '*.go' \))
CONFIGS   := $(shell find $(PREFIX)/api/config/default \( -name '*' \))
TEMPLATES := $(shell find $(PREFIX)/tmpl/templates \( -name '*' \))

THISFILE := $(lastword $(MAKEFILE_LIST))
THISDIR  := $(shell dirname $(realpath $(THISFILE)))
GOBIN    := $(THISDIR)/bin

# Prepend this repo's bin directory to our path since we'll want to
# install some build tools there for use during the build process.
PATH := $(GOBIN):$(PATH)

# Export GOBIN env variable so `go install` picks it up correctly.
export GOBIN

all: $(shell sudo chmod -R 0777 $(PREFIX)) clean fmt bin/phenix-app-archive-experiment

bin/go-bindata:
	go install github.com/go-bindata/go-bindata/v3/go-bindata

api/config/bindata.go: $(CONFIGS) bin/go-bindata
	$(GOBIN)/go-bindata -pkg config -prefix $(PREFIX)/api/config/default -o $(PREFIX)/api/config/bindata.go $(PREFIX)/api/config/default/...

generate-bindata: api/config/bindata.go tmpl/bindata.go

tmpl/bindata.go: $(TEMPLATES) bin/go-bindata
	$(GOBIN)/go-bindata -pkg tmpl -prefix $(PREFIX)/tmpl/templates -o $(PREFIX)/tmpl/bindata.go $(PREFIX)/tmpl/templates/...

fmt: 
	go fmt $(GOSOURCES)

clean:
	$(RM) bin/phenix-app-archive-experiment
	$(RM) bin/go-bindata

bin/phenix-app-archive-experiment: $(GOSOURCES)	generate-bindata 
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-X 'phenix-app-archive-experiment/version.Version=$(VERSION)' -s -w" -trimpath -o bin/phenix-app-archive-experiment $(GOSOURCES)