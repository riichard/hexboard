# Stack Research: Hexboard Web UI Enhancement

**Research type:** Stack — subsequent milestone
**Date:** 2026-03-01

## Summary

Adding SQLite persistence, Philips Hue Bridge API, and a browser-side segment drawing tool to an existing Go 1.19 / ARMv6 / CGo app.

---

## SQLite — Go Driver

### Recommended: `mattn/go-sqlite3`

- **Version:** v1.14.x (stable)
- **Why:** Most mature Go SQLite driver; CGo-based (compatible with existing CGo build pipeline); widely deployed on ARM
- **ARMv6 compatibility:** ✓ Confirmed — mattn/go-sqlite3 ships SQLite C source and compiles on-device; tested on RPi Zero
- **Confidence:** High

### What NOT to use:
- `modernc.org/sqlite` — pure Go, no CGo needed, but uses `unsafe` tricks and has had ARMv6 bugs; mattn is safer given we're already using CGo
- `gorm` — ORM overhead unnecessary; plain `database/sql` is sufficient for simple history table

### Schema:
```sql
CREATE TABLE history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  type TEXT NOT NULL,        -- 'message' | 'drawing'
  content TEXT NOT NULL,     -- JSON blob
  duration_sec INTEGER,      -- display duration, NULL = global default
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### File location on Pi:
- `/var/lib/hexboard/history.db` — writable by root systemd service
- Or adjacent to binary: `~/hexboard.db`
- **Recommended:** `/var/lib/hexboard/history.db` with `mkdir -p` in service ExecStartPre

---

## Philips Hue — Local REST API

### API: Hue Bridge Local REST v1

- **Endpoint:** `http://<bridge-ip>/api/<username>/lights/<id>/state`
- **Method:** `PUT {"on": true}`
- **Auth:** Pre-provisioned username (long token) — no OAuth, no cloud dependency
- **Version:** Hue API v1 (still supported; v2 CLIP API requires HTTPS + self-signed cert complexity)
- **Go client:** Standard `net/http` — no library needed
- **Confidence:** High

### Config file format (TOML or simple key=value):
```toml
[hue]
bridge_ip = "192.168.178.x"
api_key   = "abcdef..."
device_id = "3"
```

Or simpler — just use a JSON file or environment-style flat file.

---

## Browser-side: Segment Drawing Tool

### Approach: SVG rendered server-side template or inline HTML + vanilla JS

- **No framework** — consistent with existing plain HTML UI
- **SVG for segments** — each 16-segment digit rendered as SVG paths; clickable `<path>` or `<rect>` elements
- **16-segment layout:** Standard 16-segment display — 7 vertical/horizontal bars + 4 diagonals + dot
- **State:** JS object `segments[col][row]` = 16-bit bitmask; serialized to JSON on submit
- **Confidence:** High

### Segment encoding:
- Standard 16-segment bit assignment (matches existing Go font package)
- POST body: JSON array of 128 segment bitmasks (one per digit)

---

## Build pipeline — no changes needed

- `make deploy` (rsync + `go build` on device) works unchanged
- Additional CGo for SQLite compiles on-device alongside existing drivers CGo
- No cross-compilation required

---

## Confidence Summary

| Component | Library/Approach | Confidence |
|-----------|-----------------|------------|
| SQLite | mattn/go-sqlite3 | High |
| Hue API | net/http + local REST v1 | High |
| Drawing UI | Vanilla JS + SVG | High |
| Config file | TOML (BurntSushi/toml) or JSON | Medium — keep it simple |
