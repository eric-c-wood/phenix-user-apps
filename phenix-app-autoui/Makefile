SHELL := /bin/bash

# Default version number to git commit hash if not set.
DATE    := $(shell date -u)

GOSOURCES   := $(shell find src -maxdepth 1 -name '*.go')

all: clean fmt bin/auto-ui

fmt: 
	go fmt $(GOSOURCES)

clean:
	$(RM) bin/auto-ui

install:
	sudo cp bin/auto-ui /usr/local/bin

bin/auto-ui: $(GOSOURCES)	
	mkdir -p bin
	CGO_ENABLED=1 GOOS=linux go build -v -mod=vendor -ldflags="-s -w" -trimpath -o bin/auto-ui $(GOSOURCES)
