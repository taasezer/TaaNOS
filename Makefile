.PHONY: all build-linux build-windows build-arm64 clean

BINARY_NAME_LINUX=taanos-linux
BINARY_NAME_ARM64=taanos-arm64
BINARY_NAME_WINDOWS=taanos.exe
BINARY_NAME_DARWIN_AMD64=taanos-darwin-amd64
BINARY_NAME_DARWIN_ARM64=taanos-darwin-arm64

MAIN_PATH=./cmd/taanos

all: build-linux build-arm64 build-windows build-darwin-amd64 build-darwin-arm64

build-linux:
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(BINARY_NAME_LINUX) $(MAIN_PATH)

build-arm64:
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o $(BINARY_NAME_ARM64) $(MAIN_PATH)

build-windows:
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o $(BINARY_NAME_WINDOWS) $(MAIN_PATH)

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME_LINUX) $(BINARY_NAME_ARM64) $(BINARY_NAME_WINDOWS) $(BINARY_NAME_DARWIN_AMD64) $(BINARY_NAME_DARWIN_ARM64)

build-darwin-amd64:
	@echo "Building for macOS (Intel amd64)..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o $(BINARY_NAME_DARWIN_AMD64) $(MAIN_PATH)

build-darwin-arm64:
	@echo "Building for macOS (Apple Silicon arm64)..."
	GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o $(BINARY_NAME_DARWIN_ARM64) $(MAIN_PATH)
