# 2026-07-10 yongtang Codex proxy failure

## Symptom

Codex sessions on `yongtang` could start, but model requests failed or hung.

## Root Cause

`~/.local/bin/codex` selected `127.0.0.1:7891` whenever the TCP port was open.
On this host `7891` was still listening via `privoxy`, but its upstream route was
broken: HTTPS CONNECT to `api.openai.com` ended with TLS EOF. The host's `mihomo`
proxy on `127.0.0.1:7890` was healthy.

Existing tmux sessions also preserved the old proxy environment, so already
running Codex panes had to be restarted to pick up the corrected wrapper.

## Fix

Added `deploy/codex-wrapper.sh` and deployed it to:

`/home/yongtang/.local/bin/codex`

The wrapper probes `api.openai.com` through candidate local proxies and prefers
the working `127.0.0.1:7890` route before falling back to `7891`.

## Verification

- `curl -x http://127.0.0.1:7890 https://api.openai.com/v1/models` returned the
  expected OpenAI `401`, proving network reachability.
- `curl -x http://127.0.0.1:7891 https://api.openai.com/v1/models` failed with
  TLS EOF, reproducing the bad route.
- `codex exec --ephemeral ... "只输出 OK"` on `yongtang` returned `OK`.
- Restarted `~/Code/wmdlx`; the new `codex-9d3256-wmdlx` process environment
  contains `HTTPS_PROXY=http://127.0.0.1:7890`.

## Follow-up

`codex-9d3256-project-governance-plan` was not restarted because it had a
long-running `pnpm start` child process. It will pick up the fixed wrapper after
that session is restarted.

## 2026-07-10 Update: Mac Outline Transport

The user reported that Codex was still unstable through the target host's
existing routes and asked to install the same Outline connection used by the
MacBook on `yongtang`, without depending on the MacBook as a proxy.

Actions:

- Parsed the MacBook Outline NetworkExtension transport locally without printing
  the password.
- Installed a separate `outline-mac` user service on `yongtang`:
  - SOCKS: `127.0.0.1:7893`
  - HTTP via privoxy: `127.0.0.1:7894`
- Updated `deploy/codex-wrapper.sh` to probe and prefer `127.0.0.1:7894`, then
  fall back to `7890` and older routes.
- Restarted `codex-9d3256-wmdlx` and `codex-9d3256-yongtang`; both now inherit
  `HTTPS_PROXY=http://127.0.0.1:7894`.

Verification:

- Five consecutive `curl -x http://127.0.0.1:7894 https://api.openai.com/v1/models`
  checks returned the expected OpenAI `401`.
- `codex exec --ephemeral ... "只输出 OK"` returned `OK`.
- `outline-mac-sslocal.service` and `outline-mac-http-proxy.service` are active.

Still not restarted:

- `codex-9d3256-project-governance-plan` remains on `7891` because it has a
  long-running `pnpm start` child process.
