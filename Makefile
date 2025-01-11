APP_NAME=revid-serve
VERSION=1.0.0
MAIN_FILE=main.go
BUILD_DIR=build

# Build flags for maximum optimization
LDFLAGS=-w -s \
	-X 'main.Version=$(VERSION)' \
	-X 'main.BuildTime=$(shell date -u '+%Y-%m-%d %H:%M:%S')' \
	-extldflags "-static"

BUILD_FLAGS=-trimpath -ldflags "$(LDFLAGS)"

# Additional optimization flags
GOAMD64=v3
GOGC=off
GOOS=darwin

.PHONY: all clean build build-small build-all

all: clean build

# Clean build directory
clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)

# Build with basic optimizations
build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Optimized build complete: $(BUILD_DIR)/$(APP_NAME)"
	@ls -lh $(BUILD_DIR)/$(APP_NAME)

# Build with maximum size reduction
build-small:
	@echo "Building with maximum optimization..."
	GOGC=off CGO_ENABLED=0 GOAMD64=v3 go build -tags timetzdata \
		$(BUILD_FLAGS) \
		-o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Compressing binary..."
	upx --best --lzma $(BUILD_DIR)/$(APP_NAME) || echo "UPX compression skipped (upx not installed)"
	@echo "Final binary size:"
	@ls -lh $(BUILD_DIR)/$(APP_NAME)

# Build for multiple platforms
build-all: clean
	# Mac
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_FILE)
	# Linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(MAIN_FILE)
	# Windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "Multi-platform builds complete:"
	@ls -lh $(BUILD_DIR)/
