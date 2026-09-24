# Build targets for Darktide Settings Sync

.PHONY: all build build-windows build-linux test clean

all: test build

test:
	go test -v ./...

build: build-windows

build-windows:
	go build -ldflags "-s -w -H windowsgui" -o darktide-settings-sync.exe .

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o darktide-settings-sync-linux .

clean:
	rm -f darktide-settings-sync.exe darktide-settings-sync.bin darktide-settings-sync-linux
