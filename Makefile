DIST    := dist
ZIG     ?= zig
LDFLAGS := -ldflags="-s -w"

.PHONY: all build dist test clean

all: build

# Build for the current platform.
build:
	@mkdir -p $(DIST)
	go build $(LDFLAGS) -o $(DIST)/symex-mcp ./cmd/symex-mcp
	go build $(LDFLAGS) -o $(DIST)/symex     ./cmd/symex

test:
	go test ./...

clean:
	rm -rf $(DIST)

# Cross-compilation.
#
# Linux targets use zig cc for fully static musl binaries and can be built
# from any platform (requires zig in PATH or ZIG=/path/to/zig).
#
# Darwin targets link against Apple frameworks (CoreFoundation, Security,
# resolv) which are not available outside the macOS SDK. Build these on a
# macOS host or a macOS CI runner where the native toolchain is present.
#
# Usage:
#   make dist-linux   — build Linux artifacts (any platform + zig)
#   make dist-darwin  — build macOS artifacts (macOS only)
#   make dist         — both of the above

define zigbuild
	@mkdir -p $(DIST)
	CGO_ENABLED=1 GOOS=$(1) GOARCH=$(2) \
	  CC="$(ZIG) cc -target $(3)" \
	  CXX="$(ZIG) c++ -target $(3)" \
	  go build $(LDFLAGS) -o $(4) $(5)
endef

define gobuild
	@mkdir -p $(DIST)
	CGO_ENABLED=1 GOOS=$(1) GOARCH=$(2) \
	  go build $(LDFLAGS) -o $(3) $(4)
endef

dist: dist-linux dist-darwin dist-windows

dist-linux: \
	$(DIST)/symex-mcp-linux-amd64 \
	$(DIST)/symex-mcp-linux-arm64 \
	$(DIST)/symex-linux-amd64     \
	$(DIST)/symex-linux-arm64

dist-darwin: \
	$(DIST)/symex-mcp-darwin-amd64 \
	$(DIST)/symex-mcp-darwin-arm64 \
	$(DIST)/symex-darwin-amd64     \
	$(DIST)/symex-darwin-arm64

$(DIST)/symex-mcp-linux-amd64: ; $(call zigbuild,linux,amd64,x86_64-linux-musl,$@,./cmd/symex-mcp)
$(DIST)/symex-mcp-linux-arm64: ; $(call zigbuild,linux,arm64,aarch64-linux-musl,$@,./cmd/symex-mcp)
$(DIST)/symex-linux-amd64:     ; $(call zigbuild,linux,amd64,x86_64-linux-musl,$@,./cmd/symex)
$(DIST)/symex-linux-arm64:     ; $(call zigbuild,linux,arm64,aarch64-linux-musl,$@,./cmd/symex)

$(DIST)/symex-mcp-darwin-amd64: ; $(call gobuild,darwin,amd64,$@,./cmd/symex-mcp)
$(DIST)/symex-mcp-darwin-arm64: ; $(call gobuild,darwin,arm64,$@,./cmd/symex-mcp)
$(DIST)/symex-darwin-amd64:     ; $(call gobuild,darwin,amd64,$@,./cmd/symex)
$(DIST)/symex-darwin-arm64:     ; $(call gobuild,darwin,arm64,$@,./cmd/symex)

dist-windows: \
	$(DIST)/symex-mcp-windows-amd64.exe \
	$(DIST)/symex-mcp-windows-arm64.exe \
	$(DIST)/symex-windows-amd64.exe     \
	$(DIST)/symex-windows-arm64.exe

$(DIST)/symex-mcp-windows-amd64.exe: ; $(call zigbuild,windows,amd64,x86_64-windows-gnu,$@,./cmd/symex-mcp)
$(DIST)/symex-mcp-windows-arm64.exe: ; $(call zigbuild,windows,arm64,aarch64-windows-gnu,$@,./cmd/symex-mcp)
$(DIST)/symex-windows-amd64.exe:     ; $(call zigbuild,windows,amd64,x86_64-windows-gnu,$@,./cmd/symex)
$(DIST)/symex-windows-arm64.exe:     ; $(call zigbuild,windows,arm64,aarch64-windows-gnu,$@,./cmd/symex)
