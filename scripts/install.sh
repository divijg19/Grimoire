#!/usr/bin/env sh
set -eu

REPO="${GRIMOIRE_REPO:-divijg19/Grimoire}"
BIN_NAME="${GRIMOIRE_BIN:-grimoire}"
VERSION="${GRIMOIRE_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

os_raw="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch_raw="$(uname -m)"

case "$os_raw" in
  linux*) os="linux" ;;
  darwin*) os="darwin" ;;
  *)
    echo "Unsupported OS: $os_raw" >&2
    exit 1
    ;;
esac

case "$arch_raw" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $arch_raw" >&2
    exit 1
    ;;
esac

artifact="${BIN_NAME}-${os}-${arch}"
if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/${artifact}"
else
  url="https://github.com/${REPO}/releases/download/${VERSION}/${artifact}"
fi

tmp="$(mktemp "${TMPDIR:-/tmp}/grimoire-installer.XXXXXX")"
cleanup() {
  rm -f "$tmp"
}
trap cleanup EXIT

echo "Downloading ${url} ..."
curl -fsSL "$url" -o "$tmp"
chmod +x "$tmp"

if [ ! -d "$INSTALL_DIR" ]; then
  mkdir -p "$INSTALL_DIR" 2>/dev/null || true
fi

if [ ! -w "$INSTALL_DIR" ]; then
  fallback_dir="${HOME}/.local/bin"
  mkdir -p "$fallback_dir"
  INSTALL_DIR="$fallback_dir"
  echo "Install dir not writable, using ${INSTALL_DIR}" >&2
fi

target="${INSTALL_DIR}/${BIN_NAME}"
mv "$tmp" "$target"

echo "Installed ${BIN_NAME} to ${target}"
if [ "$INSTALL_DIR" = "${HOME}/.local/bin" ]; then
  echo "Ensure ${HOME}/.local/bin is in your PATH."
fi
