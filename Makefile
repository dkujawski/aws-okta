# Goals:
# - user can build binaries on their system without having to install special tools
# - user can fork the canonical repo and expect to be able to run CircleCI checks
#
# This makefile is meant for humans
VERSION := $(shell git describe --tags --always --dirty="-dev")
LDFLAGS := -ldflags='-X "main.Version=$(VERSION)"'

test:
	GO111MODULE=on go test -mod=vendor -v ./...

all: dist/aws-okta-$(VERSION)-darwin-amd64 dist/aws-okta-$(VERSION)-linux-amd64

clean:
	rm -rf ./dist

dist/:
	mkdir -p dist
	ifeq ("$(wildcard $('.git'))","")
		vf := version
		echo ".git dir not found, reading version from file: $(vf)"
		VERSION := $(shell cat $(vf))
	endif

dist/aws-okta-$(VERSION)-darwin-amd64: | dist/
	GOOS=darwin GOARCH=amd64 GO111MODULE=on go build -mod=vendor $(LDFLAGS) -o $@

dist/aws-okta-$(VERSION)-linux-amd64: | dist/
	GOOS=linux GOARCH=amd64 GO111MODULE=on go build -mod=vendor $(LDFLAGS) -o $@

dist/aws-okta-$(VERSION).exe: | dist/
	go mod vendor
	go get -u ./...
	go get github.com/nomad-software/vend
	vend
	GOOS=windows GOARCH=amd64 GO111MODULE=on go build -mod=vendor $(LDFLAGS) -o dist/aws-okta.exe

.PHONY: clean all