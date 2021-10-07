# vim: set ft=make ffs=unix fenc=utf8:
# vim: set noet ts=4 sw=4 tw=72 list:
#
PRVVER != git describe --tags --abbrev=0
BRANCH != git rev-parse --symbolic-full-name --abbrev-ref HEAD
GITHASH != git rev-parse --short HEAD

all: release

release: release_freebsd release_linux

release_freebsd: build
	@echo "Building static FreeBSD ...."
	@env CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go install -tags osusergo,netgo -ldflags "-X main.privprodVersion=$(PRVVER)-$(GITHASH)/$(BRANCH)" ./...

release_linux: build
	@echo "Building static Linux ...."
	@env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -tags osusergo,netgo -ldflags "-X main.privprodVersion=$(PRVVER)-$(GITHASH)/$(BRANCH)" ./...
	
build: generate
	@go build ./...

generate:
	@go generate ./cmd/...
