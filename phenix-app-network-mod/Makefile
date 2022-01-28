SHELL := /bin/bash

# Default version number to git commit hash if not set.
VER     := $(or $(VER),$(shell git log -1 --format="%h"))
COMMIT  := $(or $(COMMIT),$(shell git log -1 --format="%h - %ae"))
DATE    := $(shell date -u)
VERSION := $(VER) (commit $(COMMIT)) $(DATE)

GOSOURCES   := $(shell find src \( -name '*.go' \))

all: clean fmt bin/phenix-app-network-mod install

fmt: 
	go fmt $(GOSOURCES)

clean:
	$(RM) bin/phenix-app-network-mod

install:
	sudo cp bin/phenix-app-network-mod /usr/local/bin

bin/phenix-app-network-mod: $(GOSOURCES)	
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -v -mod=vendor -ldflags="-X 'phenix-app-network-mod/version.Version=$(VERSION)' -s -w" -trimpath -o bin/phenix-app-network-mod $(GOSOURCES)
