# Phase 1: Store - Context

**Gathered:** 2026-03-01
**Status:** Ready for planning

<domain>
## Phase Boundary

Add SQLite persistence for text messages sent via the web form. Every web POST writes to the DB; the RECENT section reads from the DB instead of an in-memory slice. No new UI surfaces — the existing home page layout stays the same. Drawing persistence is Phase 3's responsibility.

</domain>

<decisions>
## Implementation Decisions

### Persistence scope
- Web form POST only — DB write stays in `webHandler.send()`
- TCP messages (port 8080) are NOT persisted — they're a developer/tooling shortcut
- Cursor protocol (port 8082) is not a message; ignore it entirely

### DB schema
- Table: `messages (id INTEGER PRIMARY KEY, type TEXT, content TEXT, sent_at DATETIME)`
- `type = 'text'` for all Phase 1 entries; column exists now so Phase 3 drawings slot in without a migration
- `sent_at` stores UTC timestamp

### Recent list
- Replace the in-memory `recent []string` (volatile, max 10) with a DB query
- Show the 10 most recent messages, newest first — same count and order as today
- If DB read fails on GET, fall back to an empty list (no crash, no error shown to user)

### Write failure behavior
- Fail-soft: if a DB write fails, the message still shows on the display
- Log the error to stderr with `log.Printf` so it appears in `journalctl -fu hexboard`
- Never expose a DB error to the user (no HTTP 500, no UI change)

### First boot / setup
- Go code creates the DB file and schema automatically on startup if missing (`CREATE TABLE IF NOT EXISTS`)
- systemd `ExecStartPre` creates `/var/lib/hexboard/` with correct permissions — zero manual setup required
- DB path: `/var/lib/hexboard/hexboard.db`

### Claude's Discretion
- WAL mode enablement (`PRAGMA journal_mode=WAL`) — already decided in project decisions
- Connection pool size / `database/sql` configuration
- Exact `log.Printf` format for write errors
- Whether to use a dedicated `store` package or inline DB logic in `web.go`

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `webHandler.send(msg string)` in `web.go`: central integration point — DB write goes here
- `webHandler.mu sync.Mutex`: already guards the recent list; can coordinate DB access if needed
- `indexTmpl.Execute(w, recent)`: template already iterates over `[]string`; changing the data source to a DB query requires no template changes
- `maxRecent = 10`: constant to reuse as the query LIMIT

### Established Patterns
- Errors in goroutines are logged to stderr via `fmt.Println` / `return` — match this style with `log.Printf`
- No external dependencies beyond `go-serial` and `golang.org/x/term`; adding `mattn/go-sqlite3` is a significant new CGo dep — must build on device (ARMv6 constraint)
- `webHandler` is constructed once in `startWebServer`; a DB handle (`*sql.DB`) fits naturally as a field

### Integration Points
- `webHandler.send()` — add `store.Save(msg)` or equivalent call before/after `showMessage`
- `webHandler.ServeHTTP` GET branch — replace `h.recent` slice copy with DB query result
- `hexboard.service` `ExecStartPre` — add `mkdir -p /var/lib/hexboard && chown root:root /var/lib/hexboard`
- `go.mod` — add `github.com/mattn/go-sqlite3` dependency

</code_context>

<specifics>
## Specific Ideas

- No specific references — straightforward SQLite integration following the existing Go style
- Keep the store minimal: one table, one write function, one read function

</specifics>

<deferred>
## Deferred Ideas

- None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-store*
*Context gathered: 2026-03-01*
