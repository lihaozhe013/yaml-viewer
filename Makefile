GO_FILES := $(shell rg --files -g '*.go')
BUILD_GOOS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
RELEASE_LDFLAGS := -s -w

ifeq ($(BUILD_GOOS),windows)
RELEASE_LDFLAGS := -H=windowsgui -s -w
endif

.PHONY: run build debug release clean format

default: run

run:
	go run .

build:
	go build .


release:
	go build -tags release -trimpath -ldflags "$(RELEASE_LDFLAGS)" .

clean:
	rm -f yamlviewer yamlviewer.exe
	rm -rf dist

format:
	gofmt -w $(GO_FILES)

build-mac:
	go build -o yamlviewer .
	uv run scripts/build_macos.py

.PHONY: build-windows

build-windows:
	uv run --with Pillow scripts/build_windows.py
