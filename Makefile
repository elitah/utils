
.PHONY: all
all: build

.PHONY: build
build: init fmt utils

.PHONY: init
init:
	@mkdir -p bin

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: utils
utils:
	@go build -ldflags "-w -s" -o bin/$@

.PHONY: clean
clean:
	@go clean -i -n -x -cache
	@rm -rf bin go.sum

.PHONY: distclean
distclean:
	@go clean -i -n -x --modcache
	@rm -rf bin go.sum
