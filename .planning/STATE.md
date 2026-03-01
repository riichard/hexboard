# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-01)

**Core value:** Every message or drawing sent from the web UI should reach the display — even if it was off — and be replayable from history indefinitely.
**Current focus:** Phase 1 - Store

## Current Position

Phase: 1 of 5 (Store)
Plan: 2 of 2 in current phase (checkpoint: awaiting human verify)
Status: In progress
Last activity: 2026-03-01 — Completed 01-02 tasks (store wired into web, deployed to Pi)

Progress: [██░░░░░░░░] 10%

## Performance Metrics

**Velocity:**
- Total plans completed: 2
- Average duration: ~15 min
- Total execution time: ~31 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-store | 2 | 31 min | ~15 min |

**Recent Trend:**
- Last 5 plans: 01-01 (1 min), 01-02 (30 min)
- Trend: 30 min dominated by sqlite3 CGo compilation on Pi Zero ARMv6

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
- [01-01] SetMaxOpenConns(1) serialises writes; no store-level mutex needed
- [01-01] schema type column (DEFAULT 'text') added now for Phase 3 drawing support without migration
- [01-01] fail-soft pattern: Save returns error, caller logs and continues
- [01-02] Removed sync.Mutex from webHandler: database/sql handles concurrency, store enforces single-connection writes
- [01-02] log.Fatalf on store.OpenDB failure: missing directory surfaces at startup, not on first request
- [01-02] No db.Close(): binary runs until killed, OS reclaims file handles on exit

### Pending Todos

None yet.

### Blockers/Concerns

- Bit ordering between JS drawing tool and Go display must be defined once and derived everywhere — critical to get right in Phase 3
- Timer goroutine leak risk on rapid sends — existing issue to address during Duration phase

## Session Continuity

Last session: 2026-03-01
Stopped at: 01-02-PLAN.md checkpoint:human-verify (Tasks 1+2 complete, deployed to Pi)
Resume file: None
