.PHONY: default run debug build build-macos build-windows test format clean

default: run

run:
	go run .

debug:
	go run -tags debug . 2> debug.log

build:
	uv run scripts/build.py

build-macos:
	uv run scripts/build_macos.py

build-windows:
	uv run --with Pillow scripts/build_windows.py

test:
	go test ./...
	go test -tags debug ./...
	uv run -m unittest discover -s scripts -p 'test_*.py'

format:
	go fmt ./...

clean:
	rm -f yamlviewer yamlviewer.exe debug.log
	rm -rf dist
