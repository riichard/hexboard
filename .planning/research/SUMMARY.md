# Research Summary: Hexboard Web UI Enhancement

**Synthesized:** 2026-03-01

---

## Stack

| Component | Choice | Notes |
|-----------|--------|-------|
| SQLite driver | `mattn/go-sqlite3` | CGo, ARMv6-compatible; enable WAL mode |
| Hue API | `net/http` + local REST v1 | No extra library; HTTP not HTTPS |
| Drawing UI | Vanilla JS + inline SVG | No framework; consistent with existing UI |
| Config | Simple TOML or JSON file | `/etc/hexboard.conf` or `~/hexboard.conf` |

## Key Findings

**Stack:** All new dependencies work within existing CGo build pipeline. No cross-compilation complications — mattn/go-sqlite3 compiles on-device alongside existing drivers CGo.

**Table Stakes:**
- WAL mode on SQLite (prevent SQLITE_BUSY under concurrent read/write)
- Hue calls always in a goroutine (never block HTTP handler)
- Canonical bit-to-segment mapping shared between Go and JS (prevent mismatch)
- Config-absent = feature disabled gracefully (not a crash)

**Watch Out For:**
1. **SQLite file permissions** — use `/var/lib/hexboard/` created via systemd `ExecStartPre`
2. **Bit ordering mismatch** between JS drawing tool and Go display — define once, derive everywhere
3. **Timer goroutine leak** on rapid sends — cancel existing timer before starting new one
4. **Hue API blocking** — always fire-and-forget in goroutine

## Suggested Build Order

```
Phase 1: SQLite store (schema, CRUD, WAL mode, path)
Phase 2: Hue client + config file loading
Phase 3: Drawing screen (Go) + segment picker UI (browser)
Phase 4: Configurable timeout (global + per-message)
Phase 5: History UI + replay with duration prompt
```

## Architecture Notes

- New packages: `internal/store/`, `internal/hue/`
- Config loaded in `main.go`, passed to web handler and Hue client
- `web.go` gains: `GET /history`, `POST /draw`, `GET|POST /settings`
- Existing `POST /` extended to save to store + trigger Hue
- Drawing tool is a new HTML page (or section) with SVG grid

---

*Files: `.planning/research/STACK.md`, `FEATURES.md`, `ARCHITECTURE.md`, `PITFALLS.md`*
