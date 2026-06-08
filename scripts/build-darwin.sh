#!/usr/bin/env bash
set -euo pipefail

go test ./...
make dist-darwin
