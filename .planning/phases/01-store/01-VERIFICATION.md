---
phase: 01-store
verified: 2026-03-01T17:00:00Z
status: passed
score: 7/7 must-haves verified
re_verification: false
---

# Phase 1: Store Verification Report

**Phase Goal:** Every message sent to the display is durably persisted and survives server restarts
**Verified:** 2026-03-01T17:00:00Z
**Status:** PASSED
**Re-verification:** No — initial verification

---

## Goal Achievement

The ROADMAP.md defines four success criteria for Phase 1. All four are satisfied.

### Observable Truths (from ROADMAP.md Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Sending a message via the web form writes a record to the SQLite DB on the Pi | VERIFIED | `web.go:26` calls `store.Save(h.db, msg)` inside `send()`; `store.go:53-58` executes INSERT with UTC timestamp |
| 2 | After `make deploy` restarts the service, previously sent messages still exist in the DB | VERIFIED | WAL-mode SQLite file at stable path; `OpenDB` re-opens existing DB; schema is `CREATE TABLE IF NOT EXISTS`; SUMMARY-02 records human-verified restart test |
| 3 | The DB file lives at a stable path under `/var/lib/hexboard/` with correct permissions | VERIFIED | `store.go:14` hardcodes `dbPath = "/var/lib/hexboard/hexboard.db"`; `hexboard.service:6-7` runs `ExecStartPre=/bin/mkdir -p /var/lib/hexboard` and `ExecStartPre=/bin/chmod 750 /var/lib/hexboard` |
| 4 | Concurrent reads and writes do not produce SQLITE_BUSY errors (WAL mode active) | VERIFIED | `store.go:28` DSN contains `?_journal_mode=WAL`; `store.go:33` calls `db.SetMaxOpenConns(1)`; single-connection pool serialises writes |

**Score: 4/4 success criteria satisfied**

---

## Required Artifacts

All five artifacts declared across both PLANs exist, are substantive, and are wired.

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `gohexdump/internal/store/store.go` | SQLite store: OpenDB, Save, Recent | VERIFIED | 79 lines; exports OpenDB, Save, Recent; contains WAL DSN, SetMaxOpenConns, schema, driver import; commit `03fe990` |
| `gohexdump/go.mod` | Module dependency on mattn/go-sqlite3 | VERIFIED | Line 7: `github.com/mattn/go-sqlite3 v1.14.34 // indirect`; commit `27b8bb1` |
| `gohexdump/cmd/hexboard/web.go` | DB-backed send and recent list | VERIFIED | `webHandler.db *sql.DB` field present; `send()` calls `store.Save`; GET branch calls `store.Recent`; old `mu sync.Mutex` and `recent []string` fields removed; commit `7e8de60` |
| `gohexdump/cmd/hexboard/main.go` | DB open at startup, passed to webHandler | VERIFIED | `store.OpenDB()` at line 139 with `log.Fatalf` on error; `db` passed to `startWebServer` at line 152; commit `7e8de60` |
| `gohexdump/hexboard.service` | ExecStartPre creates /var/lib/hexboard/ | VERIFIED | Lines 6-7: `ExecStartPre=/bin/mkdir -p /var/lib/hexboard` and `ExecStartPre=/bin/chmod 750 /var/lib/hexboard`; commit `ff49dea` |

---

## Key Link Verification

All six key links from both PLANs are wired.

| From | To | Via | Status | Evidence |
|------|----|-----|--------|----------|
| `store.go OpenDB` | `/var/lib/hexboard/hexboard.db` | `sql.Open` with WAL DSN | WIRED | `store.go:28`: `dsn := dbPath + "?_journal_mode=WAL"` — pattern `_journal_mode=WAL` present |
| `store.go OpenDB` | `database/sql` | `SetMaxOpenConns(1)` | WIRED | `store.go:33`: `db.SetMaxOpenConns(1)` — pattern confirmed |
| `web.go send()` | `store.go Save()` | `store.Save(h.db, msg)` | WIRED | `web.go:26`: `if err := store.Save(h.db, msg); err != nil` — pattern `store\.Save` confirmed |
| `web.go ServeHTTP GET` | `store.go Recent()` | `store.Recent(h.db, maxRecent)` | WIRED | `web.go:58`: `recent, err := store.Recent(h.db, maxRecent)` — pattern `store\.Recent` confirmed |
| `main.go` | `store.go OpenDB()` | `store.OpenDB()` in main() | WIRED | `main.go:139`: `db, err := store.OpenDB()` — pattern `store\.OpenDB` confirmed |
| `main.go` | `web.go startWebServer` | `db` passed as parameter | WIRED | `main.go:152`: `go startWebServer(":"+*webport, screenChan, d, *timeout, db)` |

---

## Requirements Coverage

Both requirement IDs declared in both PLAN frontmatter files are accounted for.

| Requirement | Source Plan(s) | Description | Status | Evidence |
|-------------|----------------|-------------|--------|----------|
| HIST-01 | 01-01, 01-02 | All sent messages and drawings are persisted to SQLite DB on the Pi | SATISFIED | `store.Save` called in `web.go send()` on every POST; INSERT to messages table with content and UTC timestamp |
| HIST-02 | 01-01, 01-02 | History survives server restarts | SATISFIED | WAL-mode SQLite file at stable path `/var/lib/hexboard/hexboard.db`; `CREATE TABLE IF NOT EXISTS` schema; human-verified restart test confirmed in SUMMARY-02 |

No orphaned requirements: REQUIREMENTS.md maps HIST-01 and HIST-02 to Phase 1 only. Both are claimed by both plans. No Phase-1 requirements appear unclaimed.

---

## Anti-Patterns Found

No blockers or warnings detected.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `web.go` | 136, 230 | `placeholder` | Info | HTML/CSS `textarea::placeholder` and `placeholder=` attribute — not code stubs; part of the UI design |

The in-memory `recent []string` slice and `mu sync.Mutex` fields have been fully removed from `webHandler`. No TODO/FIXME/HACK comments. No empty handlers. No stub return values in the store package.

---

## Commit Verification

All four task commits referenced in SUMMARYs exist and touched the declared files:

| Commit | Message | Files Changed |
|--------|---------|---------------|
| `27b8bb1` | `chore(01-01): add mattn/go-sqlite3 dependency` | `go.mod`, `go.sum` |
| `03fe990` | `feat(01-01): create internal/store package` | `internal/store/store.go` (+79 lines) |
| `7e8de60` | `feat(01-02): wire store package into web handler and main` | `web.go`, `main.go` |
| `ff49dea` | `feat(01-02): add ExecStartPre directives and deploy to device` | `hexboard.service` |

---

## Human Verification Required

One item was gated on human verification in Plan 02 (checkpoint task). Per SUMMARY-02, this was approved:

### 1. Persistent Recent List Across Restart

**Test:** Open `http://txt`, send a test message, restart service with `ssh txt 'sudo systemctl restart hexboard'`, refresh UI
**Expected:** Previously sent message still appears in RECENT section
**Status:** APPROVED — recorded in 01-02-SUMMARY.md ("Human verify checkpoint: approved — web UI shows persistent recent list after restart")
**Why human:** Visual browser behaviour and end-to-end device state cannot be verified programmatically from dev machine

---

## Summary

Phase 1 goal is fully achieved. Every artifact is substantive (not a placeholder), every key link is wired (not orphaned), and both requirements are satisfied with direct code evidence.

The implementation follows a clean fail-soft pattern: `store.Save` failures are logged but do not prevent the message from reaching the display. The single-connection WAL configuration eliminates SQLITE_BUSY risk. The systemd `ExecStartPre` ensures the data directory exists idempotently on every boot with no manual setup required.

---

_Verified: 2026-03-01T17:00:00Z_
_Verifier: Claude (gsd-verifier)_
