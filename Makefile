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
	staticcheck ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: install-linter
install-linter:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: debug
debug:
	dlv debug --headless --api-version=2 --listen=127.0.0.1:4300 .
	# Exiting delve neglects to restore the terminal, so we do so here.
	reset

.PHONY: connect
connect:
	dlv connect 127.0.0.1:4300 .

.PHONY: install-terragrunt
install-terragrunt:
	mkdir -p ~/.local/bin
	curl -L https://github.com/gruntwork-io/terragrunt/releases/download/v0.67.0/terragrunt_linux_amd64 -o ~/.local/bin/terragrunt
	chmod +x ~/.local/bin/terragrunt

.PHONY: install-infracost
install-infracost:
	mkdir -p ~/.local/bin
	curl -L https://github.com/infracost/infracost/releases/download/v0.10.38/infracost-linux-amd64.tar.gz | tar -zxf -
	mv infracost-linux-amd64 ~/.local/bin/infracost
	chmod +x ~/.local/bin/infracost
