# Phase 2: Hue - Context

**Gathered:** 2026-03-01
**Status:** Ready for planning

<domain>
## Phase Boundary

Automatically turn on the Hue display light whenever any message is sent to the hexboard. Config-driven via a TOML file on the Pi. Non-blocking: a Hue failure must not delay or prevent the message from displaying. Degrades gracefully when config is absent.

Drawing sends (Phase 3) will also trigger Hue — the hook must be placed where both text and future drawing sends pass through.

</domain>

<decisions>
## Implementation Decisions

### Trigger scope
- All message inputs fire Hue — web form POST and TCP port 8080 both trigger the light
- Hook must be placed in `showMessage()` (in `display`, `main.go`) so all callers (web handler and TCP listener) get it automatically
- Not web-handler-only; the "any message" in the phase goal is literal

### Config format and location
- TOML file at `/var/lib/hexboard/hue.toml`
- Fields: bridge IP, API key, device ID (matching HUE-01 requirements)
- Read at startup; no hot-reload (restart required to pick up changes — consistent with systemd service model)
- If file is absent or incomplete: silently disabled, log once at startup

### Failure handling
- Runtime Hue failures (bridge unreachable, key revoked): silent log to journalctl only
- No web UI feedback for Hue status (that's v2 per requirements)
- Consistent with existing fail-soft pattern: store failures already log + continue

### Light target state
- Just turn the light on — no brightness or color temperature targeting
- Color control is explicitly out of scope for this milestone

### Claude's Discretion
- Hue API version (v1 clip vs v2)
- Package structure (new `internal/hue` package vs inline)
- Goroutine approach for non-blocking call (fire-and-forget goroutine is the obvious choice)
- Config struct design and TOML field names

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `display.showMessage()` (`main.go:66`): the single chokepoint where all message sends flow — this is where the Hue trigger hook belongs
- `store.Save()` (`store/store.go:53`): existing fail-soft pattern to follow — log the error, continue regardless
- `/var/lib/hexboard/` directory: already exists (created by Phase 1 systemd setup), correct place for `hue.toml`

### Established Patterns
- Fail-soft on external dependency: `web.go:26-29` logs store errors and continues — Hue should do the same
- Startup flags: `main.go:129-134` — config file path could be a flag or hardcoded; given the store uses a hardcoded path, a hardcoded TOML path is consistent
- Goroutines for non-blocking work: `main.go:76-80` (timeout goroutine) — fire-and-forget goroutine is the established pattern

### Integration Points
- `display.showMessage()` is called by: `webHandler.send()` (`web.go:29`) and `tcpListener` goroutine (`main.go:97`) — hooking here covers all current and future callers
- `main()` opens the DB at startup (`main.go:139`); Hue config should be loaded the same way and passed into `display` or `webHandler`

</code_context>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches for the Hue API call.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 02-hue*
*Context gathered: 2026-03-01*
