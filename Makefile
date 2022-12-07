SHELL := /bin/bash

.PHONY:	all
all:	clean build

build:
	goreleaser build --snapshot --rm-dist --single-target --debug

test:
	./hack/test.sh

vendor:
	go mod vendor

tidy:
	go mod tidy

run:
	go run cmd/main.go

lint:
	golangci-lint run

# codecov:	test
# 	bash <(curl -s https://codecov.io/bash)

clean:
	go clean
	rm -rf dist