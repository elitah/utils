
all: fmt build

build: utils

fmt:
	@go fmt ./...

utils:
	@go build -ldflags "-w -s" -o bin/utils

clean:
	@rm -rf bin

distclean:
	@rm -rf bin
	@go clean
	@go clean --modcache
