#!/usr/bin/env bash
set -Eeuo pipefail

# Install a single-node K3s server for Archon Base student practice.
# This script is intentionally explicit: use --check-only to inspect a host,
# --dry-run to print the install command, and run without either flag only
# when the host is ready and the operator intends to change the system.

readonly MIN_RAM_KIB=$((8 * 1024 * 1024))
readonly MIN_DISK_BYTES=$((20 * 1024 * 1024 * 1024))
readonly MIN_CPUS=2
readonly MIN_GPU_MB=1024
readonly DEFAULT_CHANNEL="stable"
readonly K3S_URL="https://get.k3s.io"

channel="$DEFAULT_CHANNEL"
check_only=false
dry_run=false
skip_confirmation=false

usage() {
  cat <<'EOF'
Usage: install-k3s-student.sh [options]

Install a single-node K3s server for the Archon Base enthusiast/student profile.

Options:
  --check-only          Check prerequisites without changing the host.
  --dry-run             Print the install command without running it.
  --channel CHANNEL     K3s channel (default: stable).
  --yes                 Skip the interactive confirmation prompt.
  -h, --help            Show this help.

The target profile is 8 GB RAM, 2 CPU cores, 20 GB free disk, and optional
1 GB graphics memory for lightweight GPU experiments. GPU absence is a warning,
not a failure: CPU-only K3s practice is supported.
EOF
}

log() { printf '[k3s-student] %s\n' "$*"; }
warn() { printf '[k3s-student] WARNING: %s\n' "$*" >&2; }
fail() { printf '[k3s-student] ERROR: %s\n' "$*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only) check_only=true ;;
    --dry-run) dry_run=true ;;
    --channel) [[ $# -ge 2 ]] || fail "--channel requires a value"; channel="$2"; shift ;;
    --yes) skip_confirmation=true ;;
    -h|--help) usage; exit 0 ;;
    *) fail "unknown option: $1" ;;
  esac
  shift
done

require_command() { command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"; }

check_host() {
  [[ "$(uname -s)" == "Linux" ]] || fail "K3s student installer requires Linux"
  if ! "$check_only" && ! "$dry_run" && (( EUID != 0 )); then
    fail "run the installer as root (for example: sudo $0)"
  fi
  require_command awk
  require_command curl
  require_command df
  require_command free
  require_command grep
  require_command systemctl

  local cpus ram_kib disk_available
  cpus="$(getconf _NPROCESSORS_ONLN 2>/dev/null || true)"
  [[ "$cpus" =~ ^[0-9]+$ ]] || fail "could not determine CPU count"
  if (( cpus < MIN_CPUS )); then
    fail "host has ${cpus} CPU cores; ${MIN_CPUS} are recommended"
  fi

  ram_kib="$(awk '/^MemTotal:/ {print $2}' /proc/meminfo)"
  [[ "$ram_kib" =~ ^[0-9]+$ ]] || fail "could not determine host memory"
  if (( ram_kib < MIN_RAM_KIB )); then
    fail "host has ${ram_kib} KiB RAM; at least 8 GB is recommended"
  fi

  disk_available="$(df -B1 --output=avail /var/lib 2>/dev/null | tail -n1 | tr -d '[:space:]')"
  [[ "$disk_available" =~ ^[0-9]+$ ]] || fail "could not determine free space under /var/lib"
  if (( disk_available < MIN_DISK_BYTES )); then
    fail "less than 20 GB is available under /var/lib"
  fi

  if command -v nvidia-smi >/dev/null 2>&1; then
    local gpu_mb
    gpu_mb="$(nvidia-smi --query-gpu=memory.total --format=csv,noheader,nounits 2>/dev/null | head -n1 | tr -d '[:space:]' || true)"
    if [[ "$gpu_mb" =~ ^[0-9]+$ ]] && (( gpu_mb < MIN_GPU_MB )); then
      warn "GPU reports ${gpu_mb} MB; 1 GB graphics memory is recommended for GPU experiments"
    else
      log "NVIDIA GPU detected (${gpu_mb:-unknown} MB)"
    fi
  else
    warn "nvidia-smi not found; continuing with CPU-only K3s practice"
  fi

  log "prerequisites passed: ${cpus} CPU cores, $((ram_kib / 1024 / 1024)) GB RAM, at least 20 GB available under /var/lib"
}

install_k3s() {
  local install_command
  install_command="curl -sfL ${K3S_URL} | INSTALL_K3S_CHANNEL=${channel@Q} sh -s - server --write-kubeconfig-mode 600"
  if "$dry_run"; then
    log "dry run; would execute: ${install_command}"
    return
  fi
  if ! "$skip_confirmation"; then
    printf 'This will install K3s and enable a system service on this host. Continue? [y/N] '
    read -r answer
    [[ "$answer" =~ ^[Yy]$ ]] || { log "installation cancelled"; return; }
  fi
  log "installing K3s channel ${channel}"
  curl -sfL "$K3S_URL" | INSTALL_K3S_CHANNEL="$channel" sh -s - server --write-kubeconfig-mode 600
  systemctl enable --now k3s
  install -d -m 0700 /root/.kube
  ln -sfn /etc/rancher/k3s/k3s.yaml /root/.kube/config
  log "K3s is installed; kubeconfig is available at /etc/rancher/k3s/k3s.yaml"
  log "verify with: k3s kubectl get nodes"
}

check_host
if "$check_only"; then exit 0; fi
install_k3s
