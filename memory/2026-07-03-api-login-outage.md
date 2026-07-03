# 2026-07-03 API Login Outage

## Symptom

The mobile app could not log in. Public `http://121.41.69.142:8080/health`
and `POST /api/auth/login` accepted TCP connections but returned an empty HTTP
reply. The H5 frontend on `:3001` still served pages.

## Root Cause

The `agentapi` container was still running, but the Go API process inside it had
panicked and exited. The container stayed alive because the Next.js process was
still running in the foreground.

The panic was:

```text
panic: send on closed channel
github.com/mobile-coder/cloud/internal/ws.(*Hub).SendLastOutput
```

`WSHubHandler.HandleConnection` starts `SendLastOutput` in a goroutine for new
viewers. If the viewer disconnects and unregisters first, `Unregister` closes
`client.Send`; `SendLastOutput` could then send to that closed channel and crash
the API process.

## Fix

`Hub.SendLastOutput` now verifies under the hub read lock that the client is
still registered under the expected device/session key before sending cached
output.

Regression tests cover:

- unregistered closed clients do not panic
- registered clients still receive cached output
- clients registered under a different key do not receive output

## Operational Action

The live container was restarted to restore service, then `/app/server` inside
the running container was hotpatched with a rebuilt Linux amd64 binary from this
fix and the API process was restarted.

This hotpatch survives `docker restart agentapi`, but not container recreation
from the old image. A follow-up immutable image release should include this fix.

## Verification

Commands run:

```bash
cd cloud && go test ./internal/ws -run 'TestSendLastOutput' -count=1
cd cloud && go test ./...
curl -fsS --max-time 12 http://121.41.69.142:8080/health
curl -i --max-time 12 -H 'Content-Type: application/json' \
  -d '{"email":"probe@example.com","password":"bad-password"}' \
  http://121.41.69.142:8080/api/auth/login
```

Expected production results after the fix:

- `/health` returns `{"status":"ok"}`
- invalid login returns HTTP 401 `invalid email or password`

Status: DONE_WITH_CONCERNS. The live service is restored and hotpatched; publish
a new Docker image before recreating the container.
