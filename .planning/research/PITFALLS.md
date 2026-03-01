# Pitfalls Research: Hexboard Web UI Enhancement

**Research type:** Pitfalls — subsequent milestone
**Date:** 2026-03-01

---

## SQLite on ARMv6 with CGo

### Pitfall 1: mattn/go-sqlite3 requires CGO_ENABLED=1 and gcc on device
- **Warning sign:** `go build` fails with "cgo: C compiler not found"
- **Prevention:** Verify `gcc` is installed on the Pi (`apt install gcc`). Already implied by existing CGo use — check `make build` succeeds with new dep.
- **Phase:** Phase 1 (store)

### Pitfall 2: SQLite file permissions
- **Warning sign:** `database is locked` or `unable to open database` at runtime
- **Prevention:** Service runs as root — use `/var/lib/hexboard/` (create with `ExecStartPre=mkdir -p /var/lib/hexboard`). Add to systemd unit.
- **Phase:** Phase 1 (store)

### Pitfall 3: WAL mode vs default journal mode
- **Warning sign:** Concurrent reads from web handler + writes from send goroutine cause `SQLITE_BUSY`
- **Prevention:** Enable WAL mode (`PRAGMA journal_mode=WAL`) at DB open. Also use `_busy_timeout=5000` DSN parameter.
- **Phase:** Phase 1 (store)

### Pitfall 4: go.sum / go.mod sync
- **Warning sign:** `go: missing go.sum entry`
- **Prevention:** After adding `mattn/go-sqlite3` dep locally, run `go mod tidy` before `make deploy` so go.sum is committed and synced. The Pi has internet access for `go get` if needed.
- **Phase:** Phase 1 (store)

---

## Philips Hue API

### Pitfall 5: Hue API call blocking the send response
- **Warning sign:** Web POST takes 1-2s when Hue unreachable
- **Prevention:** Always call `hue.TurnOn()` in a goroutine. Never await it in the HTTP handler.
- **Phase:** Phase 2 (hue)

### Pitfall 6: Hue API v1 vs v2 (CLIP)
- **Warning sign:** HTTPS certificate errors if using v2 endpoints
- **Prevention:** Use v1 REST (`http://`, not `https://`). Hue v1 is unencrypted local LAN — fine for home use.
- **Phase:** Phase 2 (hue)

### Pitfall 7: Hue config not set — nil pointer on startup
- **Warning sign:** Panic when Hue client tries to use empty bridge IP
- **Prevention:** If config is empty/unset, skip Hue integration entirely (feature is disabled, not broken). Log a startup warning.
- **Phase:** Phase 2 (hue)

---

## Drawing Tool

### Pitfall 8: 16-segment bit ordering mismatch between JS and Go
- **Warning sign:** Segments appear wrong on physical display
- **Prevention:** Define canonical bit-to-segment mapping in one place (Go constant file). Document it. Generate or derive JS constant from the same source. Write a round-trip test: encode in JS, decode in Go, verify segments match.
- **Phase:** Phase 3 (drawing screen + UI)

### Pitfall 9: SVG click targets too small on mobile
- **Warning sign:** Users can't reliably click individual segments
- **Prevention:** Make each segment path have a minimum hit area (use `<rect>` with transparent fill as a click proxy). Or accept desktop-only for the drawing tool.
- **Phase:** Phase 3 (drawing UI)

### Pitfall 10: Large JSON payload for 128 × uint16
- **Warning sign:** Slow POST on slow Wi-Fi
- **Prevention:** 128 uint16s = 256 bytes as hex — trivially small. No pagination or chunking needed.
- **Phase:** Phase 3

---

## Display Duration / Timeout

### Pitfall 11: Timer goroutine leak on rapid message sends
- **Warning sign:** Memory grows; old messages reappear after new ones clear
- **Prevention:** Cancel existing timer before starting a new one. Use a timer channel + select pattern, or `time.AfterFunc` with a mutex-protected cancel.
- **Phase:** Phase 4 (timeout)

### Pitfall 12: Duration of 0 causing immediate idle return
- **Warning sign:** Message flashes briefly then disappears
- **Prevention:** Treat duration=0 as "use global default". Use a sentinel value (e.g. -1) for "display indefinitely".
- **Phase:** Phase 4

---

## General

### Pitfall 13: HTTP handler goroutine and screenChan race
- **Warning sign:** Occasional missed messages under rapid sends
- **Prevention:** `screenChan` is already `chan screen.Screen` used from multiple goroutines — this is by design in the existing code. No changes needed; just don't add locks around channel sends.
- **Phase:** All phases

### Pitfall 14: Config file path — not found silently
- **Warning sign:** Hue doesn't work but no error visible
- **Prevention:** Log at startup: "config file not found at /etc/hexboard.conf — Hue disabled" vs "config loaded, Hue bridge: 192.168.x.x". Make it obvious.
- **Phase:** Phase 2
