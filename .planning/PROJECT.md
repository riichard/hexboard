# Hexboard Web UI Enhancement

## What This Is

An enhancement to the hexboard wall-display web app running on a Raspberry Pi Zero. The existing app lets users send text messages to a 4×32 grid of 16-segment LED digits; this milestone adds persistent history (messages + drawings), a segment-level drawing tool, configurable display duration, and automatic Hue smart-switch power control.

## Core Value

Every message or drawing sent from the web UI should reach the display — even if the display was off — and be replayable from history indefinitely.

## Requirements

### Validated

- ✓ Web UI sends text messages to the display — existing
- ✓ Multi-line messages (up to 4 rows) — existing
- ✓ Last 10 messages kept in memory (in-memory only) — existing
- ✓ Display timeout (30s default) returns to idle — existing
- ✓ TCP message port (8080) and cursor port (8082) — existing

### Active

- [ ] Persist all sent messages and drawings to SQLite DB on the Pi
- [ ] Show unified history of messages and drawings in the web UI
- [ ] Drawing tool: segment picker in the web UI (toggle individual segments per digit cell)
- [ ] Re-send from history with per-send duration prompt
- [ ] Configurable display duration: global default + per-message override in web UI
- [ ] Auto-turn on Hue smart switch when any message or drawing is sent
- [ ] Hue Bridge config via config file on the Pi (bridge IP + API key)

### Out of Scope

- Real-time collaboration — single-user web UI is sufficient
- Mobile-native app — web UI is responsive enough
- Hue auto-discovery — manual config file is simpler for a single-device setup
- Drawing animations / multi-frame sequences — static drawn frames only for now

## Context

- **Target device:** Raspberry Pi Zero (ARMv6, Linux), Go 1.19
- **Existing stack:** Go backend (`cmd/hexboard`), plain HTML form web UI, systemd service running as root
- **Display:** 4 rows × 32 columns of 16-segment digits (128 digits total)
- **CGo constraint:** `drivers` package uses Linux CGo — must build on device, not locally
- **Sync workflow:** `make deploy` (rsync + build + restart) — no GitHub SSH on device
- **Hue:** Philips Hue Bridge on local network; REST API with pre-provisioned API key

## Constraints

- **Platform:** ARMv6 Linux — SQLite must use a Go SQLite driver that works on this arch (CGo-based, e.g. `mattn/go-sqlite3`)
- **Build:** Must compile on device; CGo already in use — additional CGo (SQLite) is fine
- **Network:** Hue Bridge reachable on LAN; config file stores bridge IP and API key
- **UI:** No JS framework — keep plain HTML/CSS/minimal JS consistent with existing web UI

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| SQLite for history persistence | Simple, no external deps, widely supported on ARMv6 via CGo | — Pending |
| Config file for Hue credentials | Simpler than flags for a deployed service; survives restarts | — Pending |
| Segment picker drawing tool | Full control over 16-segment display; matches the hardware reality | — Pending |
| Per-send duration prompt on replay | Avoids stale durations; user always chooses when replaying | — Pending |
| Global default + per-message timeout | Covers both "set it and forget it" and one-off overrides | — Pending |

---
*Last updated: 2026-03-01 after initialization*
