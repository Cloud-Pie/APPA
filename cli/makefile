# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=appa
BUILD_DIR=build
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME)_win

.PHONY: all test clean server

compile:
	$(GOBUILD)

test:
	$(GOCLEAN) -testcache
	$(GOTEST)

clean:
	$(GOCLEAN) -testcache
	- rm  build/*
	
dist:
	go build -o build/$(BINARY_NAME)
	env GOOS=linux GOARCH=amd64 go build -o build/$(BINARY_UNIX)
	env GOOS=windows GOARCH=386 go build -o build/$(BINARY_WINDOWS)
	