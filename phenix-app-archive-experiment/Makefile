SHELL := /bin/bash

# Default version number to git commit hash if not set.
VER     := $(or $(VER),$(shell git log -1 --format="%h"))
COMMIT  := $(or $(COMMIT),$(shell git log -1 --format="%h - %ae"))
DATE    := $(shell date -u)
VERSION := $(VER) (commit $(COMMIT)) $(DATE)

GOSOURCES   := $(shell find src \( -name '*.go' \))

all: clean fmt bin/phenix-app-archive-experiment install

fmt: 
	go fmt $(GOSOURCES)

clean:
	$(RM) bin/phenix-app-archive-experiment

install:
	sudo cp bin/phenix-app-archive-experiment /usr/local/bin

bin/phenix-app-archive-experiment: $(GOSOURCES)
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -v -mod=vendor -ldflags="-X 'phenix-app-archive-experiment/version.Version=$(VERSION)' -s -w" -trimpath -o bin/phenix-app-archive-experiment $(GOSOURCES)