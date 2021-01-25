# Disable echoing of commands
MAKEFLAGS += --silent

# Add build-time variables
PREFIX := $(shell go list ./version)
VERSION := v0.1.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
GO_VERSION := $(shell go version)
COMPILE_TIME := $(shell LC_ALL=en_US date)

BUILD_VARIABLES := -X "$(PREFIX).Version=$(VERSION)" -X "$(PREFIX).Commit=$(COMMIT)" -X "$(PREFIX).GoVersion=$(GO_VERSION)" -X "$(PREFIX).CompileTime=$(COMPILE_TIME)"
BUILD_FLAGS := -ldflags '$(BUILD_VARIABLES)'

# Configure build output
binary = larch
build = GOOS=$(1) GOARCH=$(2) go build $(BUILD_FLAGS) -o build/$(binary)$(3) ./cmd/larch/larch.go
tar = cd build && tar -czf $(1)_$(2).tar.gz $(binary)$(3) && rm $(binary)$(3)
zip = cd build && zip $(1)_$(2).zip $(binary)$(3) && rm $(binary)$(3)

source := $(shell find . -type f -name '*.go' -not -iname '*_test.go')
test_source := $(shell find . -type f -name '*_test.go')

.PHONY: help build package windows darwin linux format lint test clean

# Produce a short description of available make commands
help:
	pcregrep -Mo '^(#.*\n)+^[^# ]+:' Makefile | sed "s/^\([^# ]\+\):/> \1/g" | sed "s/^#\s\+\(.\+\)/\1/g" | GREP_COLORS='ms=1;34' grep -E --color=always '^>.*|$$' | GREP_COLORS='ms=1;37' grep -E --color=always '^[^>].*|$$'

# Build for the native platform
build: build/larch

# Package for all platforms
package: ./build/frontend.zip windows darwin linux

# Format Go code
format: $(source) $(test_source) Makefile
	gofmt -l -s -w .

# Lint Go code
lint: $(source) $(test_source) Makefile
	golint .

# Test Go code
test: $(source) $(test_source) Makefile
	go test -v github.com/AlexGustafsson/larch/warc

# Build for the native platform. For cross-platform builds, see "package" below
build/larch: $(source) Makefile
	go build $(BUILD_FLAGS) -o $@ cmd/larch/larch.go

# Build for Linux
linux: build/linux_arm.tar.gz build/linux_arm64.tar.gz build/linux_386.tar.gz build/linux_amd64.tar.gz
build/linux_386.tar.gz: $(sources)
	$(call build,linux,386,)
	$(call tar,linux,386)
build/linux_amd64.tar.gz: $(sources)
	$(call build,linux,amd64,)
	$(call tar,linux,amd64)
build/linux_arm.tar.gz: $(sources)
	$(call build,linux,arm,)
	$(call tar,linux,arm)
build/linux_arm64.tar.gz: $(sources)
	$(call build,linux,arm64,)
	$(call tar,linux,arm64)

# Build for Windows
windows: build/windows_386.zip build/windows_amd64.zip
build/windows_386.zip: $(sources)
	$(call build,windows,386,.exe)
	$(call zip,windows,386,.exe)
build/windows_amd64.zip: $(sources)
	$(call build,windows,amd64,.exe)
	$(call zip,windows,amd64,.exe)

# Build for macOS
darwin: build/darwin_amd64.tar.gz
build/darwin_amd64.tar.gz: $(sources)
	$(call build,darwin,amd64,)
	$(call tar,darwin,amd64)

# Clean all dynamically created files
clean:
	rm -rf ./build &> /dev/null || true
