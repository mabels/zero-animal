BIN_NAME=./zero-animal
VERSION ?= dev
GITCOMMIT ?= $(shell git rev-list -1 HEAD)


all: test build

build:
	go build -ldflags "-s -w -X main.Version='$(VERSION)' -X main.GitCommit=$(GITCOMMIT)" -o $(BIN_NAME) .
	$(BIN_NAME) --version

release:
	goreleaser build --single-target --skip-validate --rm-dist

test:
	go test ./...

