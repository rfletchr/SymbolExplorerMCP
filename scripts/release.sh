#!/usr/bin/env bash
set -euo pipefail

gh release create "${GITHUB_REF_NAME}" \
    --generate-notes \
    --title "${GITHUB_REF_NAME}" \
    dist/*
