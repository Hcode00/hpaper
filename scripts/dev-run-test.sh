#!/usr/bin/env bash
set -euo pipefail

# dev-run-test.sh
# Helper script to test local development version of hpaper against installed AUR version.
# It will:
#  1. Stop any running hpaper daemon (AUR or dev) via IPC + process kill fallback.
#  2. Build the local development binary into ./dist/hpaper-dev
#  3. Start the dev daemon with the test wallpapers directory: ~/wallpapers
#  4. Exercise commands: current, next, prev, reload (after switching symlink) and quit
#  5. Provide clear log style output so you can visually inspect behavior.
#
# Assumptions:
#  - Go toolchain installed in PATH.
#  - hpaper installed from AUR as `hpaper` and possibly running.
#  - Test wallpaper dirs:
#       ~/wallpapers
#       ~/wallpapers/wallpapers  (secondary set for reload test via symlink swap)
#  - You optionally have a symlink ~/wallpapers/Active that you want to repoint; if not, script will create it.
#
# Notes:
#  - The script does not daemonize the dev process permanently; ctrl+c will clean up.
#  - Socket path is /tmp/hpaper_daemon.sock per code.
#
# Usage:
#   bash scripts/dev-run-test.sh

INFO() { printf "\033[1;34m[INFO]\033[0m %s\n" "$*"; }
WARN() { printf "\033[1;33m[WARN]\033[0m %s\n" "$*"; }
ERR()  { printf "\033[1;31m[ERR ]\033[0m %s\n" "$*"; }
CMD()  { printf "\033[0;36m$ %s\033[0m\n" "$*"; }

SOCK="/tmp/hpaper_daemon.sock"
DIST_DIR="./dist"
DEV_BIN="$DIST_DIR/hpaper-dev"
WALL_BASE="$HOME/wallpapers"
WALL_ALT="$HOME/wallpapers/wallpapers"
SYMLINK_DIR="$WALL_BASE/Active"

require_go() {
  if ! command -v go >/dev/null 2>&1; then
    ERR "Go toolchain not found in PATH. Install Go to build dev version."; exit 1
  fi
}

stop_running_daemon() {
  INFO "Attempting graceful shutdown of any running hpaper daemon..."
  if [ -S "$SOCK" ]; then
    if command -v hpaper >/dev/null 2>&1; then
      CMD "hpaper quit"
      if hpaper quit 2>/dev/null; then
        INFO "Sent quit via installed hpaper client.";
      fi
    else
      # fallback raw socket write
      printf 'quit\n' | socat - unix-connect:"$SOCK" 2>/dev/null || true
    fi
    sleep 0.5
  fi

  # Fallback: kill any residual processes named hpaper (avoid killing unrelated names by matching full basename)
  if pgrep -x hpaper >/dev/null 2>&1; then
    WARN "Force killing lingering hpaper processes"
    pkill -x hpaper || true
    sleep 0.3
  fi

  # Remove stale socket if present
  if [ -S "$SOCK" ]; then
    WARN "Removing stale socket $SOCK"
    rm -f "$SOCK"
  fi
}

prepare_symlink() {
  INFO "Ensuring test symlink setup for reload scenario"
  if [ ! -d "$WALL_BASE" ]; then
    ERR "Base wallpaper directory $WALL_BASE does not exist."; exit 1
  fi
  if [ ! -d "$WALL_ALT" ]; then
    WARN "Alternate wallpaper directory $WALL_ALT not found; reload test may be limited"
  fi
  if [ ! -e "$SYMLINK_DIR" ]; then
    CMD "ln -s '$WALL_BASE' '$SYMLINK_DIR'"
    ln -s "$WALL_BASE" "$SYMLINK_DIR"
    INFO "Created initial symlink Active -> $WALL_BASE"
  elif [ ! -L "$SYMLINK_DIR" ]; then
    WARN "$SYMLINK_DIR exists and is not a symlink; reload test will still work but symlink switch step will be skipped"
  else
    INFO "Symlink already present: $(readlink "$SYMLINK_DIR")"
  fi
}

build_dev() {
  INFO "Building development binary"
  mkdir -p "$DIST_DIR"
  CMD "go build -o $DEV_BIN ."
  if go build -o "$DEV_BIN" .; then
    INFO "Dev binary built at $DEV_BIN"
  else
    ERR "Build failed"; exit 1
  fi
}

start_daemon() {
  INFO "Starting dev daemon using symlink directory: $SYMLINK_DIR"
  CMD "$DEV_BIN start $SYMLINK_DIR"
  "$DEV_BIN" start "$SYMLINK_DIR" &
  DEV_PID=$!
  INFO "Dev daemon PID $DEV_PID"; sleep 1
}

run_command() {
  local label=$1; shift
  INFO "Test: $label"
  CMD "$DEV_BIN $*"
  if ! "$DEV_BIN" "$@"; then
    ERR "Command failed: $*"
  fi
  sleep 0.4
}

reload_test() {
  if [ -L "$SYMLINK_DIR" ] && [ -d "$WALL_ALT" ]; then
    INFO "Switching symlink target to alternate directory for reload test"
    CMD "ln -sfn '$WALL_ALT' '$SYMLINK_DIR'"
    ln -sfn "$WALL_ALT" "$SYMLINK_DIR"
    ls -l "$SYMLINK_DIR" || true
    run_command "Reload after symlink switch" reload
    run_command "Get current after reload" current
  else
    WARN "Skipping symlink reload test (requirements not met)"
  fi
}

test_sequence() {
  run_command "Fetch current wallpaper" current
  run_command "Next wallpaper" next
  run_command "Previous wallpaper" prev
  reload_test
}

cleanup() {
  INFO "Cleaning up: sending quit to dev daemon"
  if kill -0 "$DEV_PID" 2>/dev/null; then
    CMD "$DEV_BIN quit"
    "$DEV_BIN" quit || true
    sleep 0.5
    if kill -0 "$DEV_PID" 2>/dev/null; then
      WARN "Force killing dev daemon PID $DEV_PID"
      kill "$DEV_PID" || true
    fi
  fi
}

main() {
  require_go
  stop_running_daemon
  prepare_symlink
  build_dev
  start_daemon
  trap cleanup EXIT INT TERM
  test_sequence
  INFO "All tests executed. You can inspect logs above."
}

main "$@"
