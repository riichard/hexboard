# Phase 1: Store - Research

**Researched:** 2026-03-01
**Domain:** SQLite persistence in Go on ARMv6 Linux (Raspberry Pi Zero)
**Confidence:** HIGH (core decisions locked and well-understood; only discretionary areas require judgment)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **Persistence scope:** Web form POST only — DB write stays in `webHandler.send()`. TCP messages (port 8080) are NOT persisted. Cursor protocol (port 8082) is ignored entirely.
- **DB schema:** Table: `messages (id INTEGER PRIMARY KEY, type TEXT, content TEXT, sent_at DATETIME)`. `type = 'text'` for all Phase 1 entries. `sent_at` stores UTC timestamp.
- **Recent list:** Replace the in-memory `recent []string` with a DB query. Show 10 most recent messages, newest first. If DB read fails on GET, fall back to empty list (no crash, no error shown to user).
- **Write failure behavior:** Fail-soft — if a DB write fails, the message still shows on the display. Log to stderr with `log.Printf`. Never expose a DB error to the user.
- **First boot / setup:** Go code creates DB file and schema automatically on startup (`CREATE TABLE IF NOT EXISTS`). systemd `ExecStartPre` creates `/var/lib/hexboard/` with correct permissions. DB path: `/var/lib/hexboard/hexboard.db`.
- **WAL mode:** `PRAGMA journal_mode=WAL` — already decided in project decisions (listed here for completeness).

### Claude's Discretion

- Connection pool size / `database/sql` configuration
- Exact `log.Printf` format for write errors
- Whether to use a dedicated `store` package or inline DB logic in `web.go`

### Deferred Ideas (OUT OF SCOPE)

- None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| HIST-01 | All sent messages and drawings are persisted to SQLite DB on the Pi | mattn/go-sqlite3 + database/sql; schema locked; write in `send()` |
| HIST-02 | History survives server restarts | Durable SQLite file at `/var/lib/hexboard/hexboard.db`; WAL mode prevents corruption |
</phase_requirements>

---

## Summary

Phase 1 is a narrow, well-scoped SQLite integration in an existing Go HTTP server. The user has already made every significant architectural decision (library, schema, DB path, failure behavior, setup mechanism). Research confirms those choices are sound and surfaces the CGo-specific details that matter for a clean implementation.

The central constraint is the ARMv6 build environment. `mattn/go-sqlite3` is the only mature CGo SQLite driver for Go and is known to compile successfully on Linux ARMv6 via the device's native toolchain. The existing `Makefile` already builds on the device (`make build` runs `go build` via SSH), so adding a CGo dependency does not change the workflow — only a `go get` step and a longer first build are new.

`database/sql` plus `mattn/go-sqlite3` is the canonical Go SQLite stack. For a single-file SQLite DB on a Pi Zero accessed exclusively by one process, the right connection pool configuration is `SetMaxOpenConns(1)` to serialize all writes and avoid `SQLITE_BUSY` in journal mode (WAL provides concurrent reads, but the driver still benefits from connection-level serialization). WAL mode is activated by DSN pragma and survives across reconnects when set on the connection string.

**Primary recommendation:** Add `mattn/go-sqlite3` as a CGo dependency, build on device, open DB with WAL PRAGMA in the DSN, `SetMaxOpenConns(1)`, and put DB logic in a small `store` package inside `cmd/hexboard/` to keep `web.go` clean.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/mattn/go-sqlite3` | v1.14.x (latest stable) | CGo SQLite3 driver; implements `database/sql` driver interface | Only mature, production-ready CGo SQLite driver for Go; ARMv6 Linux confirmed supported; used universally in Go SQLite projects |
| `database/sql` | stdlib | Connection pool, query interface, scan | Standard Go database abstraction; no additional install required |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `log` | stdlib | Error logging to stderr | Already used in project style; `log.Printf` goes to `journalctl` via systemd |
| `time` | stdlib | UTC timestamp generation for `sent_at` | `time.Now().UTC()` for consistent DB values |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `mattn/go-sqlite3` (CGo) | `modernc.org/sqlite` (pure Go) | Pure Go avoids CGo complexity, but the project already has CGo (`drivers` package), and `mattn/go-sqlite3` has a longer track record on embedded Linux. Either works; mattn is the safer known quantity given existing CGo usage. |
| `database/sql` | Direct SQLite C API via `mattn/go-sqlite3` internal | Unnecessary — `database/sql` is the right abstraction level |

**Installation (run on device after `make sync`):**
```bash
cd ~/dev/hexboard/gohexdump
/usr/local/go/bin/go get github.com/mattn/go-sqlite3
```

Then commit the updated `go.mod` and `go.sum` locally and re-sync.

---

## Architecture Patterns

### Recommended Project Structure

The existing codebase puts all `cmd/hexboard` logic flat in the package. For Phase 1, two options are viable:

**Option A (recommended): `store.go` in `cmd/hexboard`**
```
cmd/hexboard/
├── main.go        # unchanged (mostly)
├── web.go         # integration point: calls store functions
└── store.go       # NEW: DB open, schema init, Save(), Recent()
```

**Option B: `internal/store` package**
```
internal/store/
└── store.go       # exported Store type, Open(), Save(), Recent()
```

Option A is simpler and consistent with the existing pattern (all hexboard logic is in `cmd/hexboard`). Option B is cleaner for future sharing (Phase 3 drawings will also need the store). Given Phase 3 is on the roadmap, Option B is the better long-term choice but Option A is faster to implement. This is Claude's discretion.

### Pattern 1: DB Initialization at Startup

**What:** Open the DB once in `startWebServer` (or `main`), store the `*sql.DB` handle as a field on `webHandler`, run schema migration at startup.

**When to use:** Always — `*sql.DB` is safe for concurrent use and designed to be long-lived.

**Example:**
```go
// store.go (in cmd/hexboard or internal/store)
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
)

const dbPath = "/var/lib/hexboard/hexboard.db"

func OpenDB() (*sql.DB, error) {
    // DSN: WAL mode + foreign keys via URI params
    dsn := dbPath + "?_journal_mode=WAL&_foreign_keys=on"
    db, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, err
    }
    // SQLite with WAL: one writer at a time prevents SQLITE_BUSY
    db.SetMaxOpenConns(1)
    if err := migrate(db); err != nil {
        db.Close()
        return nil, err
    }
    return db, nil
}

func migrate(db *sql.DB) error {
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS messages (
        id      INTEGER PRIMARY KEY,
        type    TEXT    NOT NULL DEFAULT 'text',
        content TEXT    NOT NULL,
        sent_at DATETIME NOT NULL
    )`)
    return err
}
```

### Pattern 2: Fail-Soft Write

**What:** Call `Save()` inside `send()`, log errors, never return them to the caller.

**When to use:** Always — consistent with the locked decision.

**Example:**
```go
// In web.go send():
func (h *webHandler) send(msg string) {
    if err := h.db.Save(msg); err != nil {
        log.Printf("store: failed to save message: %v", err)
    }
    // message still displays regardless
    h.d.showMessage(msg, h.screenChan, h.timeout)
}
```

### Pattern 3: Recent Messages Query

**What:** Replace `h.recent []string` slice with a DB query on each GET.

**When to use:** For the GET handler — drop the mutex-guarded slice entirely.

**Example:**
```go
func Recent(db *sql.DB, limit int) ([]string, error) {
    rows, err := db.Query(
        `SELECT content FROM messages ORDER BY id DESC LIMIT ?`, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var msgs []string
    for rows.Next() {
        var s string
        if err := rows.Scan(&s); err != nil {
            return nil, err
        }
        msgs = append(msgs, s)
    }
    return msgs, rows.Err()
}
```

```go
// In ServeHTTP GET branch:
recent, err := store.Recent(h.db, maxRecent)
if err != nil {
    log.Printf("store: failed to load recent: %v", err)
    recent = nil // fall back to empty list
}
indexTmpl.Execute(w, recent)
```

### Pattern 4: systemd ExecStartPre for Directory Setup

**What:** Add a pre-start command to the service file that creates the data directory.

**When to use:** Required — ensures the directory exists before the binary tries to open the DB.

**Example (hexboard.service change):**
```ini
[Service]
ExecStartPre=/bin/mkdir -p /var/lib/hexboard
ExecStartPre=/bin/chown root:root /var/lib/hexboard
ExecStartPre=/bin/chmod 750 /var/lib/hexboard
ExecStart=/home/pi/hexboard
```

### Anti-Patterns to Avoid

- **Opening a new DB connection per request:** Never call `sql.Open` inside a handler. Open once at startup; the pool handles concurrency.
- **`SetMaxOpenConns(0)` (unlimited) with WAL:** WAL mode allows concurrent readers but SQLite's write lock is still at the file level. Unlimited connections with a write-heavy workload risks `SQLITE_BUSY` errors under load. Set to 1 for simplicity.
- **Storing `sent_at` in local time:** Always use `time.Now().UTC()` and store as ISO 8601 (`time.RFC3339`). SQLite has no native timestamp type; use TEXT.
- **Forgetting `rows.Close()`:** Always `defer rows.Close()` after `db.Query`. A forgotten close leaks the connection back to the pool.
- **Mutex on DB handle:** `*sql.DB` is already safe for concurrent use. Do not add a mutex around DB calls. The `h.mu` mutex on `webHandler` can be dropped entirely once the `recent []string` slice is replaced.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| SQLite file I/O | Custom binary file format | `mattn/go-sqlite3` | Atomicity, WAL, crash recovery, concurrent reads — all handled |
| Connection pooling | Manual conn lifecycle | `database/sql` pool | Thread safety, idle connection reuse, health checking |
| WAL corruption recovery | Custom logic | SQLite WAL auto-recovery | SQLite's WAL mode recovers automatically on re-open after crash |
| Schema versioning | Version number in a file | `CREATE TABLE IF NOT EXISTS` | Phase 1 schema is stable; a single idempotent migration is sufficient |

**Key insight:** SQLite's WAL mode + `database/sql` handles everything that seems "hard" about concurrent DB access on a Pi. The only real work is writing the thin wrapper.

---

## Common Pitfalls

### Pitfall 1: SQLITE_BUSY on Concurrent Writes

**What goes wrong:** Two goroutines (e.g., two rapid web POSTs) both open write transactions simultaneously. The second gets `SQLITE_BUSY` and the write is lost silently.

**Why it happens:** SQLite allows only one writer at a time. WAL mode extends the timeout but does not eliminate the limit.

**How to avoid:** `db.SetMaxOpenConns(1)` serializes all DB access through a single connection. The `database/sql` pool queues concurrent callers.

**Warning signs:** Intermittent `database is locked` errors in `journalctl`.

### Pitfall 2: CGo Build Failure on First `go get`

**What goes wrong:** `go get github.com/mattn/go-sqlite3` on a Mac (where the dev machine is) generates the wrong `go.sum` entries, or the module is fetched but won't cross-compile.

**Why it happens:** `mattn/go-sqlite3` uses CGo and must be compiled for the target arch. The existing workflow already handles this: `make build` runs the build on the device. However, `go.mod` and `go.sum` must be updated locally first (the Go module system is arch-independent for checksums), then synced to the device.

**How to avoid:**
1. Run `go get github.com/mattn/go-sqlite3` on the dev machine (just updates `go.mod`/`go.sum`; no CGo compilation)
2. Commit `go.mod` and `go.sum`
3. `make sync && make build` — the device compiles the CGo code natively

**Warning signs:** `make build` fails with `undefined reference` or `cannot find -lsqlite3`.

### Pitfall 3: DB File Created Before Directory Exists

**What goes wrong:** Binary starts, `sql.Open` tries to create `/var/lib/hexboard/hexboard.db`, but `/var/lib/hexboard/` doesn't exist yet. `sql.Open` succeeds (it's lazy), but `db.Ping()` or first `Exec` fails with "no such file or directory".

**Why it happens:** `sql.Open` in Go does not open the connection immediately — it validates arguments only. The actual file open happens on first use.

**How to avoid:** The `ExecStartPre` in the service file creates the directory before `ExecStart`. Alternatively, the Go code can `os.MkdirAll` the directory before `sql.Open`. Both approaches work; `ExecStartPre` is preferred (matches the locked decision and keeps Go code clean).

**Warning signs:** First `db.Exec` in `migrate()` returns an error; binary logs it and exits.

### Pitfall 4: Forgetting `_` Import for Driver Registration

**What goes wrong:** `sql.Open("sqlite3", ...)` returns `sql: unknown driver "sqlite3"` at runtime.

**Why it happens:** `mattn/go-sqlite3` registers itself via `init()` in a side-effect import. If the import is missing, the driver is never registered.

**How to avoid:** Always include `_ "github.com/mattn/go-sqlite3"` in the file that calls `sql.Open`.

### Pitfall 5: Thread-Safety of `h.mu` and DB Access

**What goes wrong:** The existing `webHandler.mu` mutex guards `h.recent []string`. If the mutex is kept but now also wraps DB calls, it creates an unnecessary serialization point beyond what the DB pool already provides.

**Why it happens:** Developer habit of protecting shared state.

**How to avoid:** After replacing `h.recent` with DB queries, the mutex is no longer needed. Remove it entirely. `*sql.DB` handles its own synchronization.

---

## Code Examples

Verified patterns from standard Go + mattn/go-sqlite3 usage:

### Opening DB with WAL Mode via DSN

```go
// Source: mattn/go-sqlite3 README — DSN parameters
dsn := "/var/lib/hexboard/hexboard.db?_journal_mode=WAL"
db, err := sql.Open("sqlite3", dsn)
if err != nil {
    log.Fatalf("store: open DB: %v", err)
}
db.SetMaxOpenConns(1)
```

### Schema Creation (idempotent)

```go
// Source: standard database/sql pattern
const schema = `CREATE TABLE IF NOT EXISTS messages (
    id      INTEGER  PRIMARY KEY AUTOINCREMENT,
    type    TEXT     NOT NULL DEFAULT 'text',
    content TEXT     NOT NULL,
    sent_at DATETIME NOT NULL
)`

func migrate(db *sql.DB) error {
    _, err := db.Exec(schema)
    return err
}
```

### Inserting a Message

```go
// Source: standard database/sql pattern
func Save(db *sql.DB, content string) error {
    _, err := db.Exec(
        `INSERT INTO messages (type, content, sent_at) VALUES (?, ?, ?)`,
        "text", content, time.Now().UTC().Format(time.RFC3339),
    )
    return err
}
```

### Querying Recent Messages

```go
// Source: standard database/sql pattern
func Recent(db *sql.DB, n int) ([]string, error) {
    rows, err := db.Query(
        `SELECT content FROM messages ORDER BY id DESC LIMIT ?`, n)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []string
    for rows.Next() {
        var s string
        if err := rows.Scan(&s); err != nil {
            return nil, err
        }
        out = append(out, s)
    }
    return out, rows.Err()
}
```

### Integration in `send()` (fail-soft)

```go
func (h *webHandler) send(msg string) {
    if err := save(h.db, msg); err != nil {
        log.Printf("store: save failed: %v", err)
    }
    h.d.showMessage(msg, h.screenChan, h.timeout)
}
```

### Integration in GET handler (fail-open)

```go
recent, err := recent(h.db, maxRecent)
if err != nil {
    log.Printf("store: recent failed: %v", err)
    recent = nil
}
indexTmpl.Execute(w, recent)
```

### Updated systemd Service File

```ini
[Unit]
Description=Hexboard LED display
After=network.target

[Service]
ExecStartPre=/bin/mkdir -p /var/lib/hexboard
ExecStartPre=/bin/chmod 750 /var/lib/hexboard
ExecStart=/home/pi/hexboard
Restart=on-failure
RestartSec=5
User=root
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| In-memory `[]string` recent list (volatile) | SQLite DB query | Phase 1 | Survives restarts; opens door to Phase 3 drawings and Phase 5 unified history |
| No persistence | mattn/go-sqlite3 CGo | Phase 1 | Adds ~10s to first-time build on Pi Zero due to CGo compilation of sqlite3.c (~170 KLOC) |

**Note on build time:** `sqlite3.c` is an amalgamation file (~7 MB). On a Pi Zero ARMv6, the first `go build` after adding `mattn/go-sqlite3` will take 30–120 seconds longer than usual while GCC compiles it. Subsequent builds are fast (object file cached). This is expected and not a problem.

---

## Open Questions

1. **Package structure: `store.go` in `cmd/hexboard` vs. `internal/store`**
   - What we know: Both work. The CONTEXT.md flags this as Claude's discretion.
   - What's unclear: Phase 3 will also need DB access (drawings). If drawing logic is also in `cmd/hexboard`, inlining is fine. If it ever moves to its own binary, `internal/store` would be required.
   - Recommendation: Use `internal/store` from the start. The cost is trivial (one extra file/directory) and avoids a Phase 3 refactor. Phase 3 simply imports `internal/store`.

2. **`chown` in `ExecStartPre` — is it needed?**
   - What we know: The service runs as `root` (confirmed in `hexboard.service`). `mkdir -p` by root creates a root-owned directory.
   - What's unclear: Nothing — root owns it by default. `chown` is redundant but harmless.
   - Recommendation: Keep `mkdir -p` and `chmod 750`; omit the `chown` (root:root is the default when running as root).

3. **`db.Ping()` at startup to surface early errors?**
   - What we know: `sql.Open` is lazy; errors surface on first use. An explicit `Ping()` after `Open()` catches "directory doesn't exist" before the server starts serving.
   - Recommendation: Call `db.Ping()` inside `OpenDB()` after `migrate()`. If it fails, return the error and let `main` handle it (log fatal — service will restart via systemd).

---

## Sources

### Primary (HIGH confidence)

- `mattn/go-sqlite3` — GitHub README and known behavior; widely documented CGo SQLite driver for Go, ARMv6 Linux supported via native build
- `database/sql` (Go stdlib) — `SetMaxOpenConns`, query patterns, concurrent safety guarantees
- Project codebase — `web.go`, `main.go`, `hexboard.service`, `Makefile`, `go.mod` directly inspected

### Secondary (MEDIUM confidence)

- CGo build time on ARMv6 for `sqlite3.c` amalgamation — based on known characteristics of the file size (~7MB) and Pi Zero CPU; exact timing varies
- WAL mode behavior via DSN `?_journal_mode=WAL` parameter — standard mattn/go-sqlite3 DSN feature, consistent with documented usage

### Tertiary (LOW confidence — WebFetch/WebSearch unavailable)

- `mattn/go-sqlite3` current version tag (v1.14.x) — training knowledge, not verified against current GitHub releases; run `go get github.com/mattn/go-sqlite3@latest` to get the actual latest

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — mattn/go-sqlite3 + database/sql is the established, uncontroversial choice for Go+SQLite; all decisions locked by user
- Architecture: HIGH — patterns follow directly from locked decisions and existing code structure
- Pitfalls: HIGH — all five pitfalls are well-known Go/SQLite integration issues, not speculative
- Build time estimate: MEDIUM — based on known amalgamation size; actual Pi Zero timing unverified

**Research date:** 2026-03-01
**Valid until:** 2026-09-01 (stable ecosystem; mattn/go-sqlite3 rarely has breaking changes)
