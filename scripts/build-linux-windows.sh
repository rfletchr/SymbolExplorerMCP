#!/usr/bin/env bash
set -euo pipefail

# Install zig if not already on PATH
if ! command -v zig &>/dev/null; then
    ZIG_VERSION="0.13.0"
    ZIG_DIR="$HOME/zig-linux-x86_64-${ZIG_VERSION}"
    if [ ! -d "$ZIG_DIR" ]; then
        curl -fsSL "https://ziglang.org/download/${ZIG_VERSION}/zig-linux-x86_64-${ZIG_VERSION}.tar.xz" \
            | tar -xJ -C "$HOME"
    fi
    export PATH="$ZIG_DIR:$PATH"
fi

go test ./...
make dist-linux dist-windows
