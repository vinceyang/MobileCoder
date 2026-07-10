#!/usr/bin/env bash
set -euo pipefail

CMD_NAME="$(basename "$0")"
SERVER="${MOBILECODER_SERVER:-121.41.69.142:8080}"
AI_TOOL="${MOBILECODER_AI:-codex}"
CLIENT="${MOBILECODER_CLIENT:-$HOME/.local/bin/mobilecoder-client}"
RUN_DIR="${MOBILECODER_RUN_DIR:-$PWD}"

export TERM="${TERM:-xterm-256color}"
if [ "$TERM" = "xterm-ghostty" ]; then
  export TERM="xterm-256color"
fi

canonical_dir() {
  cd "$1" 2>/dev/null && pwd -P
}

slugify() {
  value="${1:-dir}"
  value="$(printf '%s' "$value" | tr -c 'A-Za-z0-9._-' '-' | sed 's/--*/-/g; s/^-//; s/-$//')"
  if [ -z "$value" ]; then
    value="dir"
  fi
  printf '%s' "$value"
}

path_hash() {
  if command -v sha1sum >/dev/null 2>&1; then
    printf '%s' "$1" | sha1sum | awk '{print substr($1,1,8)}'
  elif command -v shasum >/dev/null 2>&1; then
    printf '%s' "$1" | shasum | awk '{print substr($1,1,8)}'
  else
    printf '%s' "$1" | cksum | awk '{print $1}'
  fi
}

default_session_name() {
  dir="$(canonical_dir "$1")"
  base="$(basename "$dir")"
  printf 'mobilecoder-agent-%s-%s' "$(slugify "$base")" "$(path_hash "$dir")"
}

if ! RUN_DIR="$(canonical_dir "$RUN_DIR")"; then
  echo "run directory not found: $RUN_DIR" >&2
  exit 1
fi
SESSION="${MOBILECODER_SESSION:-$(default_session_name "$RUN_DIR")}"
LOG_FILE="${MOBILECODER_LOG:-$HOME/.MobileCoder/$SESSION.log}"

add_path() {
  case ":$PATH:" in
    *":$1:"*) ;;
    *) PATH="$1:$PATH" ;;
  esac
}

add_path "$HOME/.local/bin"
add_path "$HOME/.bun/bin"
add_path "$HOME/.cargo/bin"
if [ -d "$HOME/.nvm/versions/node" ]; then
  node_bin="$(find "$HOME/.nvm/versions/node" -maxdepth 2 -type d -name bin 2>/dev/null | sort | tail -n 1 || true)"
  if [ -n "${node_bin:-}" ]; then
    add_path "$node_bin"
  fi
fi
export PATH

server_host() {
  printf '%s' "${SERVER%%:*}"
}

append_no_proxy_host() {
  host="$1"
  if [ -z "$host" ]; then
    return
  fi

  current="${NO_PROXY:-${no_proxy:-}}"
  case ",$current," in
    *",$host,"*) ;;
    *) current="${current:+$current,}$host" ;;
  esac
  export NO_PROXY="$current"
  export no_proxy="$current"
}

sync_tmux_environment() {
  append_no_proxy_host "$(server_host)"

  for name in HTTP_PROXY HTTPS_PROXY ALL_PROXY NO_PROXY http_proxy https_proxy all_proxy no_proxy; do
    if [ "${!name+x}" = "x" ]; then
      tmux set-environment -g "$name" "${!name}" >/dev/null 2>&1 || true
    else
      tmux set-environment -gu "$name" >/dev/null 2>&1 || true
    fi
  done
}

usage() {
  cat <<EOF
Usage: $CMD_NAME [start|status|logs|stop|restart|attach|attach-agent|list]

Environment overrides:
  MOBILECODER_SERVER   default: $SERVER
  MOBILECODER_AI       default: $AI_TOOL
  MOBILECODER_SESSION  default: $SESSION
  MOBILECODER_CLIENT   default: $CLIENT
  MOBILECODER_LOG      default: $LOG_FILE
  MOBILECODER_RUN_DIR  default: $RUN_DIR
  MOBILECODER_DEVICE_NAME  default: host name
EOF
}

require_tools() {
  if ! command -v tmux >/dev/null 2>&1; then
    echo "tmux not found in PATH" >&2
    exit 1
  fi
  if ! command -v "$AI_TOOL" >/dev/null 2>&1; then
    echo "$AI_TOOL not found in PATH" >&2
    exit 1
  fi
  if [ ! -x "$CLIENT" ]; then
    echo "MobileCoder client is not executable: $CLIENT" >&2
    exit 1
  fi
}

quote_arg() {
  printf '%q' "$1"
}

attach_command() {
  printf 'TERM=xterm-256color tmux attach -t %s' "$(quote_arg "$1")"
}

tool_session_name() {
  device_id_path="$HOME/.MobileCoder/device-id"
  if [ ! -r "$device_id_path" ]; then
    return 1
  fi
  device_id="$(tr -d '[:space:]' < "$device_id_path")"
  if [ "${#device_id}" -lt 6 ]; then
    return 1
  fi
  dir_name="$(basename "$RUN_DIR")"
  if [ -z "$dir_name" ] || [ "$dir_name" = "/" ]; then
    dir_name="root"
  fi
  dir_name="${dir_name//\//-}"
  dir_name="${dir_name// /_}"
  printf '%s-%s-%s' "$AI_TOOL" "${device_id:0:6}" "$dir_name"
}

print_attach_info() {
  echo
  echo "Supervisor attach:"
  echo "  $(attach_command "$SESSION")"

  if tool_session="$(tool_session_name 2>/dev/null)"; then
    echo "Tool attach ($AI_TOOL):"
    echo "  $(attach_command "$tool_session")"
    if tmux has-session -t "$tool_session" 2>/dev/null; then
      echo "Tool session: running"
    else
      echo "Tool session: missing"
      echo "Hint: run '$CMD_NAME restart' in $RUN_DIR to recreate it."
    fi
  else
    echo "Tool attach ($AI_TOOL): unavailable until the device id is created."
  fi
}

show_status() {
  echo "server:  $SERVER"
  echo "ai:      $AI_TOOL"
  echo "session: $SESSION"
  echo "client:  $CLIENT"
  echo "log:     $LOG_FILE"
  echo "run_dir: $RUN_DIR"
  echo
  if tmux has-session -t "$SESSION" 2>/dev/null; then
    echo "tmux:    running"
    pane_pid="$(tmux list-panes -t "$SESSION" -F '#{pane_pid}' 2>/dev/null | head -n 1 || true)"
    if [ -n "${pane_pid:-}" ]; then
      echo "pane:    $pane_pid"
      ps -eo pid=,ppid=,stat=,etime=,args= | while read -r pid ppid stat etime args; do
        if [ "$pid" = "$pane_pid" ] || [ "$ppid" = "$pane_pid" ]; then
          printf '%s %s %s %s %s\n' "$pid" "$ppid" "$stat" "$etime" "$args"
        fi
      done
    fi
  else
    echo "tmux:    stopped"
  fi
  print_attach_info
}

list_services() {
  tmux list-sessions 2>/dev/null | grep -E '^mobilecoder-agent(:|-)' || true
}

start_service() {
  require_tools
  mkdir -p "$(dirname "$LOG_FILE")" "$RUN_DIR"
  sync_tmux_environment

  if tmux has-session -t "$SESSION" 2>/dev/null; then
    if tool_session="$(tool_session_name 2>/dev/null)" && ! tmux has-session -t "$tool_session" 2>/dev/null; then
      echo "$SESSION is running but $tool_session is missing; restarting $SESSION."
      stop_service
    else
      echo "$SESSION is already running."
      show_status
      return
    fi
  fi

  start_cmd="PATH=$(quote_arg "$PATH")"
  if [ -n "${MOBILECODER_DEVICE_NAME:-}" ]; then
    start_cmd="$start_cmd MOBILECODER_DEVICE_NAME=$(quote_arg "$MOBILECODER_DEVICE_NAME")"
  fi
  start_cmd="$start_cmd $(quote_arg "$CLIENT") -server $(quote_arg "$SERVER") -ai $(quote_arg "$AI_TOOL") >> $(quote_arg "$LOG_FILE") 2>&1"

  tmux new-session -d -s "$SESSION" -c "$RUN_DIR" "$start_cmd"

  expected_tool_session=""
  if expected_tool_session="$(tool_session_name 2>/dev/null)"; then
    for _ in $(seq 1 30); do
      if ! tmux has-session -t "$SESSION" 2>/dev/null; then
        echo "failed to start $SESSION" >&2
        tail -n 120 "$LOG_FILE" 2>/dev/null || true
        exit 1
      fi
      if tmux has-session -t "$expected_tool_session" 2>/dev/null; then
        break
      fi
      sleep 1
    done
    if ! tmux has-session -t "$expected_tool_session" 2>/dev/null; then
      echo "failed to start tool session $expected_tool_session" >&2
      tail -n 120 "$LOG_FILE" 2>/dev/null || true
      exit 1
    fi
  else
    sleep 2
    if ! tmux has-session -t "$SESSION" 2>/dev/null; then
      echo "failed to start $SESSION" >&2
      tail -n 120 "$LOG_FILE" 2>/dev/null || true
      exit 1
    fi
  fi

  echo "$SESSION started."
  tail -n 80 "$LOG_FILE"
  print_attach_info
}

stop_service() {
  if tmux has-session -t "$SESSION" 2>/dev/null; then
    tmux kill-session -t "$SESSION"
    echo "$SESSION stopped."
  else
    echo "$SESSION is not running."
  fi
}

case "${1:-start}" in
  start)
    start_service
    ;;
  status)
    show_status
    ;;
  logs)
    mkdir -p "$(dirname "$LOG_FILE")"
    tail -n "${2:-120}" -f "$LOG_FILE"
    ;;
  stop)
    stop_service
    ;;
  restart)
    stop_service
    start_service
    ;;
  attach)
    if tool_session="$(tool_session_name 2>/dev/null)" && tmux has-session -t "$tool_session" 2>/dev/null; then
      exec tmux attach -t "$tool_session"
    fi
    echo "Tool session ($AI_TOOL) is not running for $RUN_DIR." >&2
    echo "Run '$CMD_NAME restart' to recreate it, or '$CMD_NAME attach-agent' for the supervisor." >&2
    exit 1
    ;;
  attach-agent)
    exec tmux attach -t "$SESSION"
    ;;
  list)
    list_services
    ;;
  -h|--help|help)
    usage
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac
