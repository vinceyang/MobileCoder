#!/usr/bin/env bash
set -euo pipefail

NODE_BIN="$HOME/.nvm/versions/node/v22.22.1/bin"
CODEX_BIN="$NODE_BIN/codex"

export PATH="$HOME/.local/bin:$HOME/.bun/bin:$NODE_BIN:$PATH"

# This service provides 127.0.0.1:7891 on some machines, but the port can be
# open while its upstream route is broken. Always probe the route before use.
systemctl --user start outline-codex-http-proxy.service >/dev/null 2>&1 || true

append_no_proxy() {
  current="${NO_PROXY:-${no_proxy:-localhost,127.0.0.1,::1}}"
  export NO_PROXY="$current"
  export no_proxy="$current"
}

proxy_reaches_openai() {
  proxy_url="$1"
  env -u HTTP_PROXY -u HTTPS_PROXY -u ALL_PROXY \
    -u http_proxy -u https_proxy -u all_proxy \
    curl -I -sS --max-time 8 -x "$proxy_url" \
      https://api.openai.com/v1/models >/dev/null 2>&1
}

use_http_proxy() {
  proxy_url="$1"
  export HTTP_PROXY="$proxy_url"
  export HTTPS_PROXY="$proxy_url"
  export http_proxy="$proxy_url"
  export https_proxy="$proxy_url"
}

append_no_proxy

if proxy_reaches_openai "http://127.0.0.1:7890"; then
  use_http_proxy "http://127.0.0.1:7890"
  export ALL_PROXY="socks5h://127.0.0.1:7890"
  export all_proxy="$ALL_PROXY"
elif proxy_reaches_openai "http://127.0.0.1:7891"; then
  use_http_proxy "http://127.0.0.1:7891"
  export ALL_PROXY="http://127.0.0.1:7891"
  export all_proxy="$ALL_PROXY"
else
  if [ -x "$HOME/.local/bin/select-codex-proxy-node" ]; then
    python3 "$HOME/.local/bin/select-codex-proxy-node" >/dev/null 2>&1 || true
  fi

  if [ -z "${HTTPS_PROXY:-}${https_proxy:-}${ALL_PROXY:-}${all_proxy:-}" ]; then
    use_http_proxy "http://127.0.0.1:7890"
    export ALL_PROXY="socks5h://127.0.0.1:7890"
    export all_proxy="$ALL_PROXY"
  fi
fi

exec "$CODEX_BIN" "$@"
