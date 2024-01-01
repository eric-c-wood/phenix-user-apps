SHELL := /bin/bash

# Default version number to git commit hash if not set.
DATE    := $(shell date -u)
WD := $(shell pwd)

GOSOURCES   := $(shell find src -maxdepth 1 -name '*.go')

all: clean fmt bin/auto-ui install

fmt: 
	go fmt $(GOSOURCES)

clean:
	$(RM) bin/auto-ui

install:
	sudo $(RM) /usr/local/bin/auto-ui
	sudo ln -s $(WD)/bin/auto-ui /usr/local/bin/auto-ui

bin/auto-ui: $(GOSOURCES)	
	mkdir -p bin
	CGO_ENABLED=1 GOOS=linux go build -v -mod=vendor -ldflags="-s -w" -trimpath -o bin/auto-ui $(GOSOURCES)
