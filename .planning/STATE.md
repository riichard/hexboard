# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-01)

**Core value:** Every message or drawing sent from the web UI should reach the display — even if it was off — and be replayable from history indefinitely.
**Current focus:** Phase 1 - Store

## Current Position

Phase: 1 of 5 (Store)
Plan: 0 of ? in current phase
Status: Ready to plan
Last activity: 2026-03-01 — Roadmap created

Progress: [░░░░░░░░░░] 0%

## Performance Metrics

**Velocity:**
- Total plans completed: 0
- Average duration: -
- Total execution time: 0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| - | - | - | - |

**Recent Trend:**
- Last 5 plans: -
- Trend: -

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- SQLite via mattn/go-sqlite3 (CGo, ARMv6-compatible); WAL mode required
- DB path: /var/lib/hexboard/; created via systemd ExecStartPre
- Hue API: plain net/http, always fire-and-forget goroutine
- Config: simple file at ~/hexboard.conf or /etc/hexboard.conf
- Drawing UI: vanilla JS + inline SVG; bit-to-segment mapping shared between Go and JS

### Pending Todos

None yet.

### Blockers/Concerns

- Bit ordering between JS drawing tool and Go display must be defined once and derived everywhere — critical to get right in Phase 3
- Timer goroutine leak risk on rapid sends — existing issue to address during Duration phase

## Session Continuity

Last session: 2026-03-01
Stopped at: Roadmap created; no phases planned yet
Resume file: None
