#!/bin/sh
set -eu

REPO="rodneyxr/taco"
PROJECT="taco"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: $1 is required" >&2
    exit 1
  fi
}

need curl
need uname
need mktemp
need tr

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  linux|darwin) ;;
  msys*|mingw*|cygwin*) os="windows" ;;
  *)
    echo "error: unsupported OS: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "error: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

ext=""
if [ "$os" = "windows" ]; then
  ext=".exe"
fi

asset="${PROJECT}_${os}_${arch}${ext}"
if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
else
  url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT INT TERM

echo "Downloading ${url}"
curl -fsSL "$url" -o "$tmp"
chmod +x "$tmp"

dest="${INSTALL_DIR}/${PROJECT}${ext}"

if mkdir -p "$INSTALL_DIR" 2>/dev/null && [ -w "$INSTALL_DIR" ]; then
  mv "$tmp" "$dest"
else
  need sudo
  sudo mkdir -p "$INSTALL_DIR"
  sudo mv "$tmp" "$dest"
fi

echo "Installed ${PROJECT} to ${dest}"
