all: test build

build:
	go build

release:
	goreleaser build --single-target --skip-validate --rm-dist

test:
	go test github.com/mabels/zero-animal

