#!/usr/bin/env bash
set -euo pipefail

REPO="arimatakao/mdx"
BIN_NAME="mdx"
INSTALL_DIR="${INSTALL_DIR:-}"
INSTALL_MODE="tar"
VERSION_INPUT="latest"
AUTO_YES="false"
TMP_DIR=""
REINSTALL_CONFIRMED="false"

cleanup_tmp_dir() {
  if [ -n "${TMP_DIR:-}" ]; then
    rm -rf "$TMP_DIR"
  fi
}

log_step() {
  echo "==> $1"
}

usage() {
  cat <<'EOF'
Usage:
  bash install.sh [--pkg] [version]

Options:
  --pkg                Install on Linux via package manager (apt/dnf/yum/apk/pacman).
                       If package manager is missing, script exits with error.
  --tar                Force tar.gz installation mode (default).
  -y, --yes            Skip confirmation prompt.
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
      -y|--yes)
        AUTO_YES="true"
        shift
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

confirm_install() {
  local message="$1"
  local answer=""

  if [ "$AUTO_YES" = "true" ]; then
    return 0
  fi

  if [ -t 0 ]; then
    read -r -p "${message} [y/N]: " answer
  elif [ -r /dev/tty ]; then
    read -r -p "${message} [y/N]: " answer < /dev/tty
  else
    echo "Error: interactive confirmation is required, but no TTY is available." >&2
    echo "Use --yes to continue non-interactively." >&2
    exit 1
  fi

  case "$answer" in
    y|Y|yes|YES) ;;
    *)
      echo "Installation cancelled."
      exit 0
      ;;
  esac
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

normalize_version() {
  local input="${1#v}"
  input="${input%%[-+]*}"
  echo "$input"
}

version_to_sort_key() {
  local version major minor patch build rest
  version="$(normalize_version "$1")"
  IFS='.' read -r major minor patch build rest <<EOF
$version
EOF
  major="${major:-0}"
  minor="${minor:-0}"
  patch="${patch:-0}"
  build="${build:-0}"
  printf '%010d%010d%010d%010d\n' "$major" "$minor" "$patch" "$build"
}

is_version_newer() {
  local target_key current_key
  target_key="$(version_to_sort_key "$1")"
  current_key="$(version_to_sort_key "$2")"
  [ "$target_key" \> "$current_key" ]
}

get_installed_version() {
  local version_output candidate_dir candidate_path match
  local -a candidates=()

  if [ "$INSTALL_MODE" = "tar" ]; then
    candidate_dir="${INSTALL_DIR:-${HOME}/.local/bin}"
    candidate_path="${candidate_dir}/${BIN_NAME}"
    if [ -x "$candidate_path" ]; then
      candidates+=("$candidate_path")
    fi
  fi

  if candidate_path="$(command -v "${BIN_NAME}" 2>/dev/null)"; then
    candidates+=("$candidate_path")
  fi

  if [ "${#candidates[@]}" -eq 0 ]; then
    return 1
  fi

  for candidate_path in "${candidates[@]}"; do
    version_output="$("$candidate_path" -v 2>/dev/null || "$candidate_path" --version 2>/dev/null || "$candidate_path" version 2>/dev/null || true)"
    if [ -z "$version_output" ]; then
      continue
    fi

    match="$(
      printf '%s\n' "$version_output" \
        | sed -E 's/\x1B\[[0-9;]*[mK]//g' \
        | grep -Eo 'v?[0-9]+(\.[0-9]+){1,3}([-.+][0-9A-Za-z.-]+)?' \
        | head -n1
    )"
    if [ -n "$match" ]; then
      echo "$match"
      return 0
    fi
  done

  return 1
}

confirm_upgrade_if_needed() {
  local target_version="$1"
  local installed_version=""

  log_step "Checking existing ${BIN_NAME} installation"
  installed_version="$(get_installed_version || true)"
  if [ -z "$installed_version" ]; then
    echo "No existing ${BIN_NAME} installation detected."
    return 0
  fi

  if is_version_newer "$target_version" "$installed_version"; then
    confirm_install "${BIN_NAME} is already installed (version ${installed_version}). Do you want to update to ${target_version}?"
    return 0
  fi

  if [ "$(normalize_version "$target_version")" = "$(normalize_version "$installed_version")" ]; then
    confirm_install "${BIN_NAME} is already installed (version ${installed_version}). Do you want to reinstall ${target_version}?"
    REINSTALL_CONFIRMED="true"
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

  log_step "Installing binary to ${dst}"
  install -m 0755 "$src" "$dst"
}

ensure_path_setup() {
  if path_contains_dir "$INSTALL_DIR"; then
    return 0
  fi

  local shell_name target_file line
  shell_name="$(basename "${SHELL:-}")"

  case "$shell_name" in
    bash)
      target_file="${HOME}/.bashrc"
      ;;
    zsh)
      target_file="${HOME}/.zshrc"
      ;;
    *)
      target_file="${HOME}/.profile"
      ;;
  esac

  if [ -f "$target_file" ] && grep -Fq "${INSTALL_DIR}" "$target_file"; then
    return 0
  fi

  line="case \":\$PATH:\" in *\":${INSTALL_DIR}:\"*) ;; *) export PATH=\"${INSTALL_DIR}:\$PATH\" ;; esac"
  log_step "Adding ${INSTALL_DIR} to PATH in ${target_file}"
  append_line_if_missing "$target_file" "$line"
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
  if [ "$REINSTALL_CONFIRMED" != "true" ]; then
    confirm_install "Install ${BIN_NAME} ${version} via ${manager} using package $(basename "$package_url")?"
  fi
  log_step "Downloading package $(basename "$package_url")"
  curl -fsSL "$package_url" -o "$package_file"
  log_step "Installing package via ${manager}"

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

  if [ "$REINSTALL_CONFIRMED" != "true" ]; then
    confirm_install "Install ${BIN_NAME} ${version} from ${archive} to ${INSTALL_DIR}?"
  fi
  log_step "Downloading ${archive}"
  curl -fsSL "$url" -o "${tmp_dir}/${archive}"

  log_step "Extracting ${archive}"
  tar -xzf "${tmp_dir}/${archive}" -C "$tmp_dir"
  if [ ! -f "${tmp_dir}/${BIN_NAME}" ]; then
    echo "Error: '${BIN_NAME}' was not found in archive ${archive}." >&2
    exit 1
  fi

  install_binary "${tmp_dir}/${BIN_NAME}"
  ensure_path_setup

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

  local os arch version invoke_cmd
  os="$(normalize_os)"
  arch="$(normalize_arch)"
  log_step "Resolving target version"
  version="$(resolve_version)"
  echo "Target version: ${version}"
  confirm_upgrade_if_needed "$version"

  TMP_DIR="$(mktemp -d)"
  log_step "Created temporary directory ${TMP_DIR}"
  trap cleanup_tmp_dir EXIT

  if [ "$INSTALL_MODE" = "pkg" ]; then
    if [ "$os" != "linux" ]; then
      echo "Error: --pkg mode is supported only on Linux." >&2
      exit 1
    fi
    install_with_package_manager "$version" "$arch" "$TMP_DIR"
  else
    install_with_tar "$version" "$os" "$arch" "$TMP_DIR"
  fi

  if command -v "${BIN_NAME}" >/dev/null 2>&1; then
    invoke_cmd="${BIN_NAME} --help"
  else
    invoke_cmd="${INSTALL_DIR}/${BIN_NAME} --help"
  fi

  echo
  echo "${BIN_NAME} has been installed successfully."
  echo "Run: ${invoke_cmd}"
}

main "$@"
