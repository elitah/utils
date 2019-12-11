
.PHONY: all
all: fmt build

.PHONY: build
build: utils

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: utils
utils:
	@go build -ldflags "-w -s" -o bin/utils

.PHONY: clean
clean:
	@rm -rf bin

.PHONY: distclean
distclean:
	@rm -rf bin
	@go clean --modcache
	@go clean
