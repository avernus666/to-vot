#!/usr/bin/env sh
set -eu
command -v go >/dev/null 2>&1 || { echo "Go is required"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Docker is required"; exit 1; }
command -v cargo >/dev/null 2>&1 || { echo "Rust/Cargo is required"; exit 1; }
go version
docker compose version
cargo --version
