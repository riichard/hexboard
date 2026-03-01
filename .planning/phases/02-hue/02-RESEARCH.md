# Phase 2: Hue - Research

**Researched:** 2026-03-01
**Domain:** Philips Hue API v1 CLIP, Go HTTP client, TOML config parsing
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Trigger scope**
- All message inputs fire Hue — web form POST and TCP port 8080 both trigger the light
- Hook must be placed in `showMessage()` (in `display`, `main.go`) so all callers (web handler and TCP listener) get it automatically
- Not web-handler-only; the "any message" in the phase goal is literal

**Config format and location**
- TOML file at `/var/lib/hexboard/hue.toml`
- Fields: bridge IP, API key, device ID (matching HUE-01 requirements)
- Read at startup; no hot-reload (restart required to pick up changes — consistent with systemd service model)
- If file is absent or incomplete: silently disabled, log once at startup

**Failure handling**
- Runtime Hue failures (bridge unreachable, key revoked): silent log to journalctl only
- No web UI feedback for Hue status (that's v2 per requirements)
- Consistent with existing fail-soft pattern: store failures already log + continue

**Light target state**
- Just turn the light on — no brightness or color temperature targeting
- Color control is explicitly out of scope for this milestone

### Claude's Discretion

- Hue API version (v1 clip vs v2)
- Package structure (new `internal/hue` package vs inline)
- Goroutine approach for non-blocking call (fire-and-forget goroutine is the obvious choice)
- Config struct design and TOML field names

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| HUE-01 | Hue Bridge IP, API key, and device ID are configured via a config file on the Pi | TOML config at `/var/lib/hexboard/hue.toml` parsed with BurntSushi/toml; Config struct with three string fields |
| HUE-02 | When any message or drawing is sent to the display, the Hue device is automatically turned on | `showMessage()` fires goroutine calling Hue v1 CLIP API: PUT `http://{bridge_ip}/api/{api_key}/lights/{device_id}/state` with `{"on":true}` |
| HUE-03 | Hue API call is non-blocking — a Hue failure does not delay or prevent the message from displaying | `go func() { hueConf.TurnOn() }()` pattern; `http.Client{Timeout: 5*time.Second}` prevents goroutine leak on unreachable bridge |
| HUE-04 | If Hue config is absent or incomplete, the feature is silently disabled (logged, not crashed) | `os.IsNotExist(err)` on DecodeFile failure; validate all three fields non-empty after decode; `nil` Config pointer = disabled |
</phase_requirements>

---

## Summary

Phase 2 adds Hue light automation to hexboard: whenever a message is displayed, the Hue bridge turns the configured light on. The implementation is a small `internal/hue` package that loads a TOML config at startup and provides a single `TurnOn()` method fired as a goroutine from `showMessage()`.

The core API call is the Hue v1 CLIP endpoint: a plain HTTP PUT to `http://{bridge_ip}/api/{api_key}/lights/{device_id}/state` with body `{"on":true}`. This requires no TLS configuration, no new dependencies beyond a TOML parser, and works on all Hue Bridge generations. A 5-second timeout on the HTTP client ensures goroutines cannot leak if the bridge is unreachable.

The only new dependency is `github.com/BurntSushi/toml` v1.3.2 (go 1.16 minimum, compatible with device's Go 1.19). The `internal/hue` package follows the same pattern as `internal/store`: a package-level path constant, a load function that returns an error, and the `display` struct in `main.go` holding a `*hue.Config` that is `nil` when Hue is disabled.

**Primary recommendation:** Use Hue v1 CLIP API over plain HTTP with `net/http` stdlib + BurntSushi/toml v1.3.2 for TOML parsing; no other new dependencies needed.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `net/http` (stdlib) | Go 1.19 | HTTP PUT to Hue bridge | Already in project; STATE.md explicitly decided "plain net/http"; no new dep |
| `github.com/BurntSushi/toml` | v1.3.2 | Parse `/var/lib/hexboard/hue.toml` | 4905 stars; `DecodeFile` is one call; go 1.16 minimum (device has 1.19) |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `encoding/json` (stdlib) | Go 1.19 | Marshal `{"on":true}` request body | Always — no external lib needed for this single-field payload |
| `log` (stdlib) | Go 1.19 | Log startup disable, runtime failures | Already used throughout project |
| `os` (stdlib) | Go 1.19 | `os.IsNotExist()` check for missing config file | Distinguishes "no config" from "bad config" |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hue v1 CLIP (plain HTTP) | Hue v2 CLIP (HTTPS) | v2 requires `crypto/tls` with `InsecureSkipVerify` (bridge self-signed cert); more complex; not needed for LAN use |
| BurntSushi/toml v1.3.2 | pelletier/go-toml v2 | go-toml v2 requires Go 1.21 — device has Go 1.19, **not compatible** |
| BurntSushi/toml v1.3.2 | manual `key=value` parser | TOML format is a locked decision; hand-rolling is the anti-pattern |
| `internal/hue` package | inline in `main.go` | Package is cleaner, follows existing `internal/*` structure, testable in isolation |

**Installation:**
```bash
# Run locally (not on device — the go.mod update syncs via rsync)
go get github.com/BurntSushi/toml@v1.3.2
# Then deploy normally
cd gohexdump && make deploy
```

---

## Architecture Patterns

### Recommended Project Structure

```
gohexdump/
├── cmd/hexboard/
│   ├── main.go          # Add: load hue config; set d.hueConf; wire into newDisplay or set after
│   └── web.go           # No changes needed (send() calls showMessage)
└── internal/
    ├── hue/
    │   └── hue.go       # NEW: Config struct, LoadConfig(), TurnOn()
    └── store/
        └── store.go     # Unchanged (pattern reference)
```

### Pattern 1: Config Load at Startup (mirrors store.OpenDB)

**What:** Load TOML at startup in `main()`, pass result to display. `nil` config means disabled.
**When to use:** Any optional external service with file-based config.

```go
// internal/hue/hue.go

package hue

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/BurntSushi/toml"
)

const configPath = "/var/lib/hexboard/hue.toml"

// Config holds the Hue connection parameters read from hue.toml.
// All three fields must be non-empty for Hue to be active.
type Config struct {
    BridgeIP string `toml:"bridge_ip"`
    APIKey   string `toml:"api_key"`
    DeviceID string `toml:"device_id"`
}

// LoadConfig reads /var/lib/hexboard/hue.toml and returns a Config.
// Returns (nil, nil) if the file is absent — Hue is silently disabled.
// Returns (nil, err) if the file exists but is malformed or incomplete.
// The caller logs the result and continues in all cases.
func LoadConfig() (*Config, error) {
    var cfg Config
    if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
        return nil, err
    }
    if cfg.BridgeIP == "" || cfg.APIKey == "" || cfg.DeviceID == "" {
        return nil, fmt.Errorf("hue.toml: bridge_ip, api_key, and device_id are all required")
    }
    return &cfg, nil
}

var hueClient = &http.Client{Timeout: 5 * time.Second}

// TurnOn sends a non-blocking PUT to the Hue bridge to turn the configured light on.
// Logs and returns silently on any error — never blocks the caller.
// Call as: go cfg.TurnOn()
func (c *Config) TurnOn() {
    body, err := json.Marshal(map[string]bool{"on": true})
    if err != nil {
        log.Printf("hue: marshal: %v", err)
        return
    }
    url := fmt.Sprintf("http://%s/api/%s/lights/%s/state", c.BridgeIP, c.APIKey, c.DeviceID)
    req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
    if err != nil {
        log.Printf("hue: request: %v", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := hueClient.Do(req)
    if err != nil {
        log.Printf("hue: put: %v", err)
        return
    }
    resp.Body.Close() // must drain/close to release connection
}
```

### Pattern 2: Wiring into main.go (mirrors db setup)

**What:** Load Hue config in `main()`, set on display struct, fire in `showMessage()`.
**When to use:** Any optional feature that hooks into a central chokepoint.

```go
// cmd/hexboard/main.go — additions

// In main():
hueCfg, err := hue.LoadConfig()
if err != nil {
    if os.IsNotExist(err) {
        log.Printf("hue: config not found — Hue disabled")
    } else {
        log.Printf("hue: config error: %v — Hue disabled", err)
    }
    // hueCfg remains nil — disabled
}

d := newDisplay()
d.hueConf = hueCfg   // nil is safe; showMessage checks before firing
d.cursor.SetCursor(0, 0)
```

```go
// display struct — add field:
type display struct {
    rain     screen.Screen
    ripple   screen.Screen
    text     screen.TextScreen
    cursor   screen.Cursor
    hueConf  *hue.Config   // nil when Hue is disabled
}
```

```go
// showMessage — add goroutine trigger:
func (d *display) showMessage(msg string, screenChan chan<- screen.Screen, timeout time.Duration) {
    d.text.Clear()
    for row, line := range strings.SplitN(strings.ToUpper(msg), "\n", 4) {
        runes := []rune(line)
        if len(runes) > 32 {
            runes = runes[:32]
        }
        d.text.WriteAt(string(runes), 0, row)
    }
    screenChan <- d.ripple
    if d.hueConf != nil {
        go d.hueConf.TurnOn()   // fire-and-forget, consistent with existing goroutine pattern
    }
    go func() {
        time.Sleep(timeout)
        screenChan <- d.rain
    }()
}
```

### Pattern 3: TOML Config File (user-created on Pi)

```toml
# /var/lib/hexboard/hue.toml
# Created manually on the Pi; restart service after editing.
bridge_ip = "192.168.1.100"
api_key   = "abcdef1234567890abcdef1234567890abcdef12"
device_id = "1"
```

Notes:
- `bridge_ip`: IP address of the Hue bridge (no `http://` prefix, no trailing slash)
- `api_key`: The "username" registered with the bridge (create via bridge button + POST to `/api`)
- `device_id`: The light's integer ID as shown in the bridge (query `GET http://{bridge_ip}/api/{api_key}/lights` to discover)

### Anti-Patterns to Avoid

- **Blocking the display for Hue:** `showMessage()` MUST return immediately. `TurnOn()` must always run in a goroutine.
- **Missing HTTP client timeout:** The default `http.DefaultClient` has no timeout. An unreachable bridge will hang the goroutine indefinitely. Always use `http.Client{Timeout: 5 * time.Second}`.
- **Not closing response body:** Even for fire-and-forget, `resp.Body.Close()` is required to return the TCP connection to the pool. Goroutines accumulate if skipped.
- **Panicking on nil Config:** Check `d.hueConf != nil` before calling `TurnOn()`. Nil Config means gracefully disabled.
- **Hardcoding bridge URL in main.go:** Keep config path constant inside `internal/hue` package (same pattern as `dbPath` in `internal/store`).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| TOML parsing | Custom `key=value` line parser | `github.com/BurntSushi/toml` | TOML has strings, multi-line, escaping, nested tables — edge cases in hand-rolled parsers |
| Hue client library | Custom retry/polling logic | `net/http` with simple PUT | "Turn on" is one PUT call; a full Hue client library adds unused API surface and CGo-incompatible dependencies |

**Key insight:** The Hue v1 "turn on" use case is literally one HTTP PUT — there is no meaningful complexity that justifies an external Hue library.

---

## Common Pitfalls

### Pitfall 1: Goroutine Leak on Unreachable Bridge

**What goes wrong:** If the bridge IP is wrong or bridge is offline, `http.Client.Do()` blocks until the OS TCP timeout (~2 minutes on Linux). A goroutine accumulates per message sent while bridge is down.
**Why it happens:** `http.DefaultClient` has no timeout.
**How to avoid:** Use a package-level client: `var hueClient = &http.Client{Timeout: 5 * time.Second}`. Five seconds is ample for a LAN request.
**Warning signs:** `runtime.NumGoroutine()` climbing; journalctl showing many pending Hue calls.

### Pitfall 2: Forgetting to Close Response Body

**What goes wrong:** `net/http` keeps the TCP connection open until Body is closed. Each goroutine holds a connection until GC. Under high message rate, exhausts file descriptors.
**Why it happens:** Response body is non-nil even when you don't need the response content.
**How to avoid:** Always `resp.Body.Close()` after `hueClient.Do()`. For fire-and-forget: `io.Copy(io.Discard, resp.Body); resp.Body.Close()` or simply `resp.Body.Close()` if the response is small (it is — Hue API response is ~50 bytes).
**Warning signs:** "too many open files" errors in journalctl.

### Pitfall 3: Device ID Integer vs String

**What goes wrong:** Hue v1 light IDs are integers (1, 2, 3...) but appear as strings in JSON and as config file values. Using `int` in the Config struct causes TOML parsing issues if the user writes `device_id = 1` (int) vs `device_id = "1"` (string).
**Why it happens:** TOML distinguishes integer from string types.
**How to avoid:** Use `string` for `DeviceID` in Config struct; format into URL with `%s`. TOML string `"1"` is unambiguous. Document that `device_id` must be a quoted string in the TOML file.
**Warning signs:** `toml: cannot unmarshal integer into Go struct field` error at startup.

### Pitfall 4: Missing File vs Bad File (HUE-04)

**What goes wrong:** Treating both "file not found" and "bad config" the same way fails to give the user meaningful feedback.
**Why it happens:** `DecodeFile` returns an error in both cases.
**How to avoid:** Use `os.IsNotExist(err)` to distinguish. Missing file: log "Hue disabled (no config)" at INFO level. Bad config: log "Hue disabled (config error: %v)" with the error detail.
**Warning signs:** User confused about why Hue isn't working with a malformed file.

### Pitfall 5: BridgeIP with http:// Prefix

**What goes wrong:** User writes `bridge_ip = "http://192.168.1.100"` in TOML, causing malformed URL in `fmt.Sprintf`.
**Why it happens:** Users often copy-paste full URLs.
**How to avoid:** The `TurnOn()` function constructs the URL with `http://` prefix. Document clearly that `bridge_ip` is an IP or hostname only. Optionally strip trailing slash: `strings.TrimRight(c.BridgeIP, "/")`.
**Warning signs:** `hue: put: unsupported protocol scheme "http://http"` in journalctl.

---

## Code Examples

Verified patterns from official sources and project conventions:

### BurntSushi/toml DecodeFile

```go
// Source: https://raw.githubusercontent.com/BurntSushi/toml/v1.3.2/decode.go
// Returns error wrapping os.PathError if file not found — check with os.IsNotExist()
var cfg hue.Config
_, err := toml.DecodeFile("/var/lib/hexboard/hue.toml", &cfg)
if err != nil {
    if os.IsNotExist(err) {
        // file absent — disable silently
    } else {
        // file exists but malformed — log and disable
    }
}
```

### net/http PUT with JSON body and timeout

```go
// Source: Go stdlib net/http documentation + project pattern (web.go uses net/http)
var hueClient = &http.Client{Timeout: 5 * time.Second}

body, _ := json.Marshal(map[string]bool{"on": true})
url := fmt.Sprintf("http://%s/api/%s/lights/%s/state", bridgeIP, apiKey, deviceID)
req, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
resp, err := hueClient.Do(req)
if err != nil {
    log.Printf("hue: put: %v", err)
    return
}
resp.Body.Close()
```

### Fire-and-forget goroutine (project pattern)

```go
// Source: main.go:76-80 (timeout goroutine) — established project pattern
if d.hueConf != nil {
    go d.hueConf.TurnOn()
}
```

### Hue v1 CLIP API endpoint

```
PUT http://{bridge_ip}/api/{api_key}/lights/{device_id}/state
Content-Type: application/json

{"on": true}

Success response: [{"success": {"/lights/1/state/on": true}}]
Error response:   [{"error": {"type": 3, "address": "/lights/1", "description": "resource not available"}}]
```

Source: go-hue library (https://github.com/heatxsink/go-hue/blob/master/lights/lights.go) confirms URL structure.

### Discover light IDs on Hue bridge

```bash
# Run this on the Pi or from any host on the LAN to see all light IDs
curl http://BRIDGE_IP/api/API_KEY/lights | python3 -m json.tool
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Hue v1 CLIP (HTTP) | Hue v2 CLIP (HTTPS + EventStream) | ~2020 (bridge v2 firmware) | v2 adds TLS complexity; v1 still works on all bridges |
| `encoding/toml` (stdlib) | `github.com/BurntSushi/toml` | Go 1.23 added `encoding/toml` | Device has Go 1.19; stdlib option NOT available |

**Deprecated/outdated:**
- Hue v1 CLIP API: officially deprecated for *new bridge registrations* as of late 2024 but continues to function on all existing bridges indefinitely. For a Pi Zero home appliance on LAN, v1 is appropriate.
- `pelletier/go-toml` v2: requires Go 1.21 — not compatible with device Go 1.19.

---

## Open Questions

1. **Hue API v1 vs v2 — is the user's bridge v2-capable?**
   - What we know: v2 bridges use HTTPS with self-signed cert and UUID-based light IDs; v1 uses plain HTTP with integer IDs; v1 works on all bridges including new ones
   - What's unclear: which bridge generation the user has
   - Recommendation: Implement v1 CLIP API. Document in TOML comment that `device_id` is an integer string (e.g. `"1"`). If user has v2-only setup and needs UUID-based ID, upgrade path is a new TurnOn implementation with InsecureSkipVerify and PUT to `/clip/v2/resource/light/{uuid}` with `hue-application-key` header.

2. **Should device_id support groups/rooms instead of individual lights?**
   - What we know: Hue v1 group endpoint is `PUT /api/{key}/groups/{id}/action` (different URL pattern from lights); CONTEXT says "device ID" without specifying
   - What's unclear: whether the user wants to control a group vs individual light
   - Recommendation: Implement for individual lights first (REQUIREMENTS.md says "device" not "group"). The architecture allows adding group support later by parameterising the URL template.

---

## Sources

### Primary (HIGH confidence)
- `https://raw.githubusercontent.com/heatxsink/go-hue/master/lights/lights.go` — Hue v1 API URL pattern and HTTP method confirmed from library source
- `https://raw.githubusercontent.com/BurntSushi/toml/v1.3.2/decode.go` — `DecodeFile` API signature and behaviour confirmed from source
- `https://raw.githubusercontent.com/BurntSushi/toml/v1.3.2/go.mod` — go 1.16 minimum version confirmed
- `https://raw.githubusercontent.com/BurntSushi/toml/master/go.mod` — v1.6.0 requires go 1.18 confirmed

### Secondary (MEDIUM confidence)
- `https://raw.githubusercontent.com/open-hue/spec/main/spec.yaml` — Hue v2 CLIP API uses `/clip/v2/resource/light/{id}` and `hue-application-key` header confirmed
- `https://raw.githubusercontent.com/pelletier/go-toml/v2/go.mod` — go 1.21 requirement confirmed (rules out this library)
- GitHub API (`api.github.com`) — star counts, release dates, repo metadata for library selection

### Tertiary (LOW confidence)
- Home Assistant `homeassistant/components/hue/const.py` — confirms v2 is current HA approach, v1 still present as legacy; inference about v1 deprecation status

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — TOML library version and Go compatibility verified from go.mod files; HTTP pattern verified from existing codebase and go-hue source
- Architecture: HIGH — follows established project patterns (store.go, main.go goroutines); hook point confirmed from CONTEXT.md code analysis
- Pitfalls: HIGH — HTTP timeout and body-close issues are documented Go stdlib behaviours; TOML type pitfall verified from library source

**Research date:** 2026-03-01
**Valid until:** 2026-04-01 (Hue v1 API stable; BurntSushi/toml v1.x stable)
