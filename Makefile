BINARY_NAME=plario
VERSION=$(shell git describe --tags --always)

OS_ARCHS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

all: build

build:
	@for target in $(OS_ARCHS); do \
		OS=$${target%/*}; \
		ARCH=$${target#*/}; \
		echo "Building $$OS/$$ARCH..."; \
		OUT_NAME=$(BINARY_NAME); \
		if [ $$OS = "windows" ]; then OUT_NAME=$${OUT_NAME}.exe; fi; \
		GOOS=$$OS GOARCH=$$ARCH go build -o bin/$$OS_$$ARCH/$$OUT_NAME .; \
	done

