PWD=$(shell pwd)

default: build

builder:
	mkdir -p bin
	docker build -f Builderfile -t jarvis-builder .
	-docker rm -f jarvis-builder
	docker run -d --name jarvis-builder \
		-v $(PWD):/go/src/github.com/apourchet/jarvis-ci \
		-v $(PWD)/bin:/jarvis-ci/bin \
		jarvis-builder /bin/sh -c "while true; do sleep 10; done"

build:
	docker exec jarvis-builder make cbuild

package: build
	docker build -f Runnerfile -t apourchet/jarvis-ci .

cbuild:
	CGO_ENABLED=0 go build -i -ldflags "-s" -o /jarvis-ci/bin/jarvis-ci .
