default: build

build:
	mkdir -p bin
	go build -o ./bin/jarvis-ci .
	docker build -f Dockerfile -t apourchet/jarvis-ci .
