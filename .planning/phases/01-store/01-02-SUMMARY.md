---
phase: 01-store
plan: 02
subsystem: web-persistence
tags: [sqlite, go, web, systemd, CGo, deployment]

# Dependency graph
requires:
  - "internal/store package (OpenDB, Save, Recent) from Plan 01"
provides:
  - "DB-backed web handler: send() persists to SQLite, GET reads from SQLite"
  - "Deployed hexboard binary with CGo sqlite3 on Pi"
  - "systemd ExecStartPre creates /var/lib/hexboard/ on every boot"
affects:
  - "All subsequent phases that depend on message history being durable"

# Tech tracking
tech-stack:
  added:
    - "database/sql import in web.go (via store package)"
    - "CGo sqlite3 linked into deployed binary (12MB vs 8.5MB previously)"
  patterns:
    - "Fail-soft DB writes in send(): log.Printf on error, message still displays"
    - "DB-first recent list: GET handler queries store.Recent, nil on error"
    - "No mutex in webHandler: database/sql.DB is concurrency-safe, store uses SetMaxOpenConns(1)"
    - "ExecStartPre for idempotent directory setup: mkdir -p is safe on restart"

key-files:
  created: []
  modified:
    - gohexdump/cmd/hexboard/web.go
    - gohexdump/cmd/hexboard/main.go
    - gohexdump/hexboard.service

key-decisions:
  - "Removed sync.Mutex from webHandler: database/sql handles concurrency, store enforces single-connection writes"
  - "log.Fatalf on store.OpenDB failure in main: missing directory surfaces at startup, not on first request"
  - "No db.Close(): binary runs until killed, OS reclaims file handles on exit"
  - "chmod 750 on /var/lib/hexboard/: root-only write, no pi user access needed since service runs as root"

# Metrics
duration: 30min
completed: 2026-03-01
---

# Phase 1 Plan 02: Store Integration Summary

**Wired internal/store into web handler and main, replaced in-memory recent slice with SQLite reads/writes, added ExecStartPre to service file, deployed and verified persistence on Pi Zero**

## Performance

- **Duration:** ~30 min (dominated by sqlite3.c CGo compilation on ARMv6: ~24 min)
- **Started:** 2026-03-01T15:15:05Z
- **Completed:** 2026-03-01T15:45:28Z
- **Tasks:** 2 (+ 1 checkpoint awaiting human verify)
- **Files modified:** 3

## Accomplishments

- Updated `web.go`: removed `mu sync.Mutex` and `recent []string` from `webHandler`; added `db *sql.DB`; `send()` now calls `store.Save`; GET branch calls `store.Recent`; `startWebServer` signature extended to accept `*sql.DB`
- Updated `main.go`: imports `store` package; calls `store.OpenDB()` before starting web server with `log.Fatalf` on error; passes `db` to `startWebServer`
- Updated `hexboard.service`: added two `ExecStartPre` lines (`/bin/mkdir -p /var/lib/hexboard` and `/bin/chmod 750 /var/lib/hexboard`)
- Deployed to Pi (txt / 192.168.178.67): synced via rsync, built on device with CGo
- Verified end-to-end persistence: web POST writes DB row, restart preserves row, WAL mode active

## Task Commits

Each task was committed atomically:

1. **Task 1: Update web.go and main.go to use store package** - `7e8de60` (feat)
2. **Task 2: Update hexboard.service with ExecStartPre and deploy to device** - `ff49dea` (feat)

## Files Created/Modified

- `gohexdump/cmd/hexboard/web.go` - DB-backed send and recent list; removed in-memory slice and mutex
- `gohexdump/cmd/hexboard/main.go` - Added store.OpenDB() at startup, passes db to startWebServer
- `gohexdump/hexboard.service` - ExecStartPre lines for mkdir and chmod

## Key Changes: web.go diff

**Before:** `webHandler` had `mu sync.Mutex` and `recent []string`; `send()` managed slice with lock; GET branch copied slice under lock.

**After:**
```go
type webHandler struct {
    screenChan chan<- screen.Screen
    d          *display
    timeout    time.Duration
    db         *sql.DB          // was: mu sync.Mutex + recent []string
}

func (h *webHandler) send(msg string) {
    if err := store.Save(h.db, msg); err != nil {
        log.Printf("store: save failed: %v", err)
    }
    h.d.showMessage(msg, h.screenChan, h.timeout)
}

// GET branch:
recent, err := store.Recent(h.db, maxRecent)
if err != nil {
    log.Printf("store: recent failed: %v", err)
    recent = nil
}
```

## Key Changes: main.go diff

Added after `flag.Parse()`:
```go
db, err := store.OpenDB()
if err != nil {
    log.Fatalf("store: open DB: %v", err)
}
```

Changed `startWebServer` call:
```go
go startWebServer(":"+*webport, screenChan, d, *timeout, db)
```

## Key Changes: hexboard.service diff

Added before `ExecStart`:
```ini
ExecStartPre=/bin/mkdir -p /var/lib/hexboard
ExecStartPre=/bin/chmod 750 /var/lib/hexboard
```

## Deploy Outcome

- Build time: ~24 minutes on Pi Zero ARMv6 (sqlite3.c amalgamation ~7MB compiled by GCC)
- Binary size: 12MB (up from 8.5MB — CGo sqlite3 statically linked)
- Service file deployed to `/etc/systemd/system/hexboard.service` via `make install`
- `ExecStartPre` runs `mkdir -p` and `chmod 750` on each service start (idempotent)

## Verification Results

All plan verification criteria met:

1. `systemctl is-active hexboard` → `active`
2. `/var/lib/hexboard/hexboard.db` exists with WAL shm/wal sidecar files
3. `curl -d "message=verify" http://192.168.178.67/` → `303`
4. `SELECT content FROM messages ORDER BY id DESC LIMIT 1;` → `verify`
5. After `sudo systemctl restart hexboard`, step 4 still returns `verify` (HIST-02 verified)
6. `PRAGMA journal_mode;` → `wal`
7. Human verify checkpoint: awaiting confirmation of web UI behavior

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing tool] sqlite3 CLI not installed on Pi**
- **Found during:** Task 2 verification
- **Issue:** `sudo sqlite3` returned "command not found" — sqlite3 CLI not pre-installed on Raspbian
- **Fix:** `sudo apt-get update && sudo apt-get install -y sqlite3` (26 seconds)
- **Files modified:** None (system package install only)
- **Commit:** N/A (system change, not committed)

**2. [Rule 1 - Column name] Plan specified `ts` column, schema uses `sent_at`**
- **Found during:** Task 2 verification
- **Issue:** Plan verification commands referenced `ts` column; actual store schema from Plan 01 uses `sent_at`
- **Fix:** Used correct column name in verification queries
- **Files modified:** None (schema is correct from Plan 01)
- **Commit:** N/A (documentation issue only)

**3. [Rule 3 - Network] `http://txt/` DNS not resolving from dev machine**
- **Found during:** Task 2 verification
- **Issue:** `curl http://txt/` returned exit 6 (DNS resolve failure); hostname `txt` not in local `/etc/hosts`
- **Fix:** Used IP address `192.168.178.67` directly for curl verification commands
- **Files modified:** None

## Issues Encountered

- `make install` via zsh shell had a `make` function collision in the shell (zsh autoload stub). Fixed by using full path `/usr/bin/make install`
- Raspbian package mirror had a stale 404 for `sqlite3` package; fixed by running `apt-get update` first

## Next Phase Readiness

- Phase 1 is functionally complete pending human verification of web UI at http://192.168.178.67
- The persistence checkpoint (HIST-02) is verified at the CLI level
- Upon human approval, STATE.md will be updated to mark Phase 1 complete and advance to Phase 2

## Self-Check: PASSED

- FOUND: gohexdump/cmd/hexboard/web.go (db field, store.Save, store.Recent)
- FOUND: gohexdump/cmd/hexboard/main.go (store.OpenDB, log.Fatalf, db passed to startWebServer)
- FOUND: gohexdump/hexboard.service (ExecStartPre lines)
- FOUND commit: 7e8de60 (feat(01-02): wire store package into web handler and main)
- FOUND commit: ff49dea (feat(01-02): add ExecStartPre directives and deploy to device)
- DEVICE: service active, DB exists, WAL confirmed, persistence verified

---
*Phase: 01-store*
*Completed: 2026-03-01 (pending human checkpoint)*
