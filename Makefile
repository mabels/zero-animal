all: test build

build:
	go build

release:
	goreleaser build --single-target --skip-validate --rm-dist
	touch release

docker: release
	cp ./dist/zero-animal_linux_arm64/zero-animal .
	docker build --no-cache -t zero-animal:arm64 . --platform=linux/arm64/v8 
	cp ./dist/zero-animal_linux_arm_7/zero-animal .
	docker build --no-cache -t zero-animal:armv7 . --platform=linux/arm/v7 
	cp ./dist/zero-animal_linux_amd64/zero-animal .
	docker build --no-cache -t zero-animal:amd64 . --platform=linux/amd64

test:
	go test github.com/mabels/zero-animal

