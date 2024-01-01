SHELL := /bin/bash

# Default version number to git commit hash if not set.
DATE    := $(shell date -u)

GOSOURCES   := $(shell find src -maxdepth 1 -name '*.go')

all: clean fmt bin/phenix-app-autoui install

fmt: 
	go fmt $(GOSOURCES)

clean:
	$(RM) bin/phenix-app-autoui

install:
	sudo cp bin/phenix-app-autoui /usr/local/bin

bin/phenix-app-autoui: $(GOSOURCES)	
	mkdir -p bin
	CGO_ENABLED=1 GOOS=linux go build -v -mod=vendor -ldflags="-s -w" -trimpath -o bin/phenix-app-autoui $(GOSOURCES)
