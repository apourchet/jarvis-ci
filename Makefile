PWD=$(shell pwd)

default: build

build:
	docker build -t apourchet/jarvis-ci .

jarvis-ci-test:
	echo "Testing"
	echo "Done :)"
