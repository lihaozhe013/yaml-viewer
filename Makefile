GO_FILES := $(shell rg --files -g '*.go')

.PHONY: run build clean format

default: run

run:
	go run .

build:
	go build -o yamlviewer .

clean:
	rm -f yamlviewer yamlviewer.exe

format:
	gofmt -w $(GO_FILES)
