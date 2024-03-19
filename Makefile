VERSION = $(shell git describe --tags --dirty --always)
GIT_COMMIT = $(shell git rev-parse HEAD)
RANDOM_SUFFIX := $(shell cat /dev/urandom | tr -dc 'a-z0-9' | head -c5)
PROJECT = github.com/leg100/pug
LD_FLAGS = " \
    -s -w \
	-X '$(PROJECT)/internal.Version=$(VERSION)'      \
	-X '$(PROJECT)/internal.Commit=$(GIT_COMMIT)'	 \
	-X '$(PROJECT)/internal.Built=$(shell date +%s)' \
	" \

.PHONY: build
build:
	CGO_ENABLED=0 go build -o _build/ -ldflags $(LD_FLAGS) ./...
	chmod -R +x _build/*

.PHONY: install
install:
	go install -ldflags $(LD_FLAGS) ./...

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	go list ./... | xargs staticcheck

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: install-linter
install-linter:
	go install honnef.co/go/tools/cmd/staticcheck@latest
