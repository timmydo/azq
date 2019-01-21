HASH := $(shell git rev-parse --short HEAD)

.PHONY: all
all: build

.PHONY: build
build:
	go build -o azq

.PHONY: version
version:
	echo $(HASH)

.PHONY: docker
docker:
	docker build -t timmydo/azq:git-$(HASH) .

.PHONY: dev
dev:
	docker run --rm -it -e GO111MODULE=on --workdir /go/src/github.com/timmydo/azq --volume $(CURDIR):/go/src/github.com/timmydo/azq quay.io/deis/go-dev:latest
