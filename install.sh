#!/usr/bin/env bash
set -euo pipefail

REPO="arimatakao/mdx"
BIN_NAME="mdx"
INSTALL_DIR="${INSTALL_DIR:-}"
INSTALL_MODE="tar"
VERSION_INPUT="latest"

usage() {
  cat <<'EOF'
Usage:
  bash install.sh [--pkg] [version]

Options:
  --pkg                Install on Linux via package manager (apt/dnf/yum/apk/pacman).
                       If package manager is missing, script exits with error.
  --tar                Force tar.gz installation mode (default).
  -h, --help           Show this help.

Examples:
  bash install.sh
  bash install.sh v1.12.0
  bash install.sh --pkg
  bash install.sh --pkg v1.12.0
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Error: required command '$1' is not installed." >&2
    exit 1
  fi
}

parse_args() {
  local positional=()
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --pkg)
        INSTALL_MODE="pkg"
        shift
        ;;
      --tar)
        INSTALL_MODE="tar"
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      -*)
        echo "Error: unknown option '$1'." >&2
        usage
        exit 1
        ;;
      *)
        positional+=("$1")
        shift
        ;;
    esac
  done

  if [ "${#positional[@]}" -gt 1 ]; then
    echo "Error: too many positional arguments." >&2
    usage
    exit 1
  fi

  if [ "${#positional[@]}" -eq 1 ]; then
    VERSION_INPUT="${positional[0]}"
  fi
}

path_contains_dir() {
  case ":${PATH:-}:" in
    *":$1:"*) return 0 ;;
    *) return 1 ;;
  esac
}

append_line_if_missing() {
  local file="$1"
  local line="$2"

  mkdir -p "$(dirname "$file")"
  touch "$file"
  if ! grep -Fqx "$line" "$file"; then
    printf "\n%s\n" "$line" >> "$file"
  fi
}

normalize_os() {
  local os
  os="$(uname -s)"
  case "$os" in
    Linux) echo "linux" ;;
    Darwin) echo "darwin" ;;
    *)
      echo "Error: unsupported OS '$os' (only Linux and macOS are supported)." >&2
      exit 1
      ;;
  esac
}

normalize_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    i386|i686) echo "386" ;;
    *)
      echo "Error: unsupported architecture '$arch'." >&2
      exit 1
      ;;
  esac
}

resolve_version() {
  if [ "$VERSION_INPUT" = "latest" ]; then
    local latest_url
    latest_url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest")"
    basename "$latest_url"
  else
    if [[ "$VERSION_INPUT" == v* ]]; then
      echo "$VERSION_INPUT"
    else
      echo "v${VERSION_INPUT}"
    fi
  fi
}

install_binary() {
  local src="$1"
  local dst="${INSTALL_DIR}/${BIN_NAME}"

  if [ ! -d "$INSTALL_DIR" ]; then
    if [ -w "$(dirname "$INSTALL_DIR")" ]; then
      mkdir -p "$INSTALL_DIR"
    else
      echo "Error: cannot create '${INSTALL_DIR}' (permission denied)." >&2
      echo "Set a writable directory, e.g.: INSTALL_DIR=\$HOME/.local/bin bash install.sh" >&2
      exit 1
    fi
  fi

  if [ ! -w "$INSTALL_DIR" ]; then
    echo "Error: '${INSTALL_DIR}' is not writable." >&2
    echo "Set a writable directory, e.g.: INSTALL_DIR=\$HOME/.local/bin bash install.sh" >&2
    exit 1
  fi

  install -m 0755 "$src" "$dst"
}

ensure_path_setup() {
  if path_contains_dir "$INSTALL_DIR"; then
    return 0
  fi

  local shell_name line
  shell_name="$(basename "${SHELL:-}")"
  line="export PATH=\"${INSTALL_DIR}:\$PATH\""

  append_line_if_missing "${HOME}/.profile" "$line"

  case "$shell_name" in
    bash)
      append_line_if_missing "${HOME}/.bashrc" "$line"
      append_line_if_missing "${HOME}/.bash_profile" "$line"
      ;;
    zsh)
      append_line_if_missing "${HOME}/.zshrc" "$line"
      append_line_if_missing "${HOME}/.zprofile" "$line"
      ;;
  esac
}

detect_pkg_manager() {
  if command -v apt >/dev/null 2>&1; then
    echo "apt"
  elif command -v dnf >/dev/null 2>&1; then
    echo "dnf"
  elif command -v yum >/dev/null 2>&1; then
    echo "yum"
  elif command -v apk >/dev/null 2>&1; then
    echo "apk"
  elif command -v pacman >/dev/null 2>&1; then
    echo "pacman"
  else
    echo "Error: no supported Linux package manager found (apt/dnf/yum/apk/pacman)." >&2
    exit 1
  fi
}

arch_regex() {
  case "$1" in
    amd64) echo "amd64|x86_64" ;;
    arm64) echo "arm64|aarch64" ;;
    386) echo "386|i386|i686" ;;
    *) echo "$1" ;;
  esac
}

package_ext_regex() {
  case "$1" in
    apt) echo "\\.deb$" ;;
    dnf|yum) echo "\\.rpm$" ;;
    apk) echo "\\.apk$" ;;
    pacman) echo "\\.pkg\\.tar\\.zst$" ;;
    *)
      echo "Error: unsupported package manager '$1'." >&2
      exit 1
      ;;
  esac
}

find_package_asset_url() {
  local version="$1"
  local arch="$2"
  local manager="$3"
  local ext_re arch_re release_json
  ext_re="$(package_ext_regex "$manager")"
  arch_re="$(arch_regex "$arch")"

  release_json="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/tags/${version}")"

  printf '%s\n' "$release_json" \
    | sed -n 's/^[[:space:]]*"browser_download_url":[[:space:]]*"\(.*\)",\?/\1/p' \
    | grep -E "${ext_re}" \
    | grep -E "(^|[._-])(${arch_re})([._-]|$)" \
    | head -n1
}

install_with_package_manager() {
  local version="$1"
  local arch="$2"
  local tmp_dir="$3"
  local manager package_url package_file

  manager="$(detect_pkg_manager)"
  package_url="$(find_package_asset_url "$version" "$arch" "$manager" || true)"
  if [ -z "$package_url" ]; then
    echo "Error: no matching package asset found for manager '${manager}' and arch '${arch}' in release '${version}'." >&2
    exit 1
  fi

  if [ "$(id -u)" -ne 0 ]; then
    echo "Error: package-manager installation requires root privileges." >&2
    echo "Run with root, for example: sudo bash install.sh --pkg ${version}" >&2
    exit 1
  fi

  package_file="${tmp_dir}/$(basename "$package_url")"
  echo "Downloading package $(basename "$package_url")..."
  curl -fL "$package_url" -o "$package_file"

  case "$manager" in
    apt)
      apt install -y "$package_file"
      ;;
    dnf)
      dnf install -y "$package_file"
      ;;
    yum)
      yum install -y "$package_file"
      ;;
    apk)
      apk add --allow-untrusted "$package_file"
      ;;
    pacman)
      pacman -U --noconfirm "$package_file"
      ;;
  esac
}

install_with_tar() {
  local version="$1"
  local os="$2"
  local arch="$3"
  local tmp_dir="$4"
  local archive url

  if [ -z "$INSTALL_DIR" ]; then
    INSTALL_DIR="${HOME}/.local/bin"
  fi

  archive="${BIN_NAME}_${version}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${version}/${archive}"

  echo "Downloading ${archive}..."
  curl -fL "$url" -o "${tmp_dir}/${archive}"

  tar -xzf "${tmp_dir}/${archive}" -C "$tmp_dir"
  if [ ! -f "${tmp_dir}/${BIN_NAME}" ]; then
    echo "Error: '${BIN_NAME}' was not found in archive ${archive}." >&2
    exit 1
  fi

  install_binary "${tmp_dir}/${BIN_NAME}"
  ensure_path_setup

  echo "Installed '${BIN_NAME}' to ${INSTALL_DIR}."
  "${INSTALL_DIR}/${BIN_NAME}" --version 2>/dev/null || "${INSTALL_DIR}/${BIN_NAME}" version 2>/dev/null || true
  if ! path_contains_dir "$INSTALL_DIR"; then
    echo "Added ${INSTALL_DIR} to your shell profile. Restart terminal or run:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi
}

main() {
  parse_args "$@"

  require_cmd curl
  require_cmd tar
  require_cmd install

  if [ -z "${PATH:-}" ]; then
    PATH="/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
    export PATH
  fi

  local os arch version tmp_dir
  os="$(normalize_os)"
  arch="$(normalize_arch)"
  version="$(resolve_version)"

  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT

  if [ "$INSTALL_MODE" = "pkg" ]; then
    if [ "$os" != "linux" ]; then
      echo "Error: --pkg mode is supported only on Linux." >&2
      exit 1
    fi
    install_with_package_manager "$version" "$arch" "$tmp_dir"
    command -v "${BIN_NAME}" >/dev/null 2>&1 && "${BIN_NAME}" --version 2>/dev/null || true
  else
    install_with_tar "$version" "$os" "$arch" "$tmp_dir"
  fi
}

main "$@"
