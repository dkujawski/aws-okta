# Goals:
# - user can build binaries on their system without having to install special tools
# - user can fork the canonical repo and expect to be able to run CircleCI checks
#
# This makefile is meant for humans

if [ ! -d ".git" ];then \
	VERSION := $(shell git describe --tags --always --dirty="-dev") \
fi

ifeq ($(strip $(VERSION)),)
	vf := version
	VERSION := $(shell cat ${vf})
endif

LDFLAGS := -ldflags='-X "main.Version=$(VERSION)"'

test:
	GO111MODULE=on go test -mod=vendor -v ./...

all: dist/aws-okta-$(VERSION)-darwin-amd64 dist/aws-okta-$(VERSION)-linux-amd64

clean:
	rm -rf ./dist

dist/:
	mkdir -p dist

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