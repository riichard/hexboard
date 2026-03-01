---
phase: 02-hue
plan: 01
subsystem: infra
tags: [hue, philips-hue, iot, smart-home, toml, http-client]

# Dependency graph
requires:
  - phase: 01-store
    provides: "Fail-soft pattern for optional integrations (LoadConfig returns (nil, nil) when disabled)"
provides:
  - "internal/hue package with Config, LoadConfig, TurnOn"
  - "Fire-and-forget goroutine pattern for non-blocking external service calls"
  - "TOML config at /var/lib/hexboard/hue.toml for optional Hue integration"
affects: [03-drawing, future-notification-integrations]

# Tech tracking
tech-stack:
  added: [github.com/BurntSushi/toml v1.3.2]
  patterns:
    - "Fire-and-forget goroutine with timeout: go d.hueConf.TurnOn() in showMessage"
    - "Fail-soft config: LoadConfig returns (nil, nil) when file absent; nil Config safe to use"
    - "5-second HTTP timeout prevents goroutine leak when bridge unreachable"

key-files:
  created:
    - gohexdump/internal/hue/hue.go
  modified:
    - gohexdump/cmd/hexboard/main.go
    - gohexdump/go.mod
    - gohexdump/go.sum

key-decisions:
  - "BurntSushi/toml for TOML parsing (pure Go, no CGo, ARMv6 compatible)"
  - "Hue config path: /var/lib/hexboard/hue.toml (matches hexboard.db location)"
  - "Fire-and-forget goroutine: always non-blocking, 5s timeout, logs errors but never fails"
  - "Fail-soft startup: missing or invalid config logs and continues with Hue disabled"

patterns-established:
  - "Pattern 1: Fire-and-forget external service call — go d.hueConf.TurnOn() with nil check"
  - "Pattern 2: Fail-soft config loading — (nil, nil) return for absent config, caller logs and continues"
  - "Pattern 3: Timeout-protected HTTP client — hueClient with 5s timeout prevents goroutine leak"

requirements-completed: [HUE-01, HUE-02, HUE-03, HUE-04]

# Metrics
duration: 16min
completed: 2026-03-01
---

# Phase 2 Plan 01: Hue Integration Summary

**Philips Hue light automation via fire-and-forget goroutine using BurntSushi/toml and Hue v1 CLIP API with 5-second timeout**

## Performance

- **Duration:** 16 min
- **Started:** 2026-03-01T19:44:31Z
- **Completed:** 2026-03-01T20:00:52Z
- **Tasks:** 3 (2 automated, 1 human-verify checkpoint)
- **Files modified:** 4

## Accomplishments
- Complete internal/hue package with Config struct, LoadConfig, and TurnOn method
- Hue integration wired into main.go display struct and showMessage function
- Fire-and-forget goroutine triggered on every message send (TCP or web form)
- Fail-soft behavior: service starts cleanly with no config or invalid config
- Deployed to Pi Zero and verified in production

## Task Commits

Each task was committed atomically:

1. **Task 1: Create internal/hue package and add BurntSushi/toml dependency** - `0073f74` (feat)
2. **Task 2: Wire hue into main.go display struct and showMessage** - `666f83e` (feat)
3. **Task 3: Deploy to Pi and verify Hue integration** - Checkpoint approved (deployment executed, service verified active with "hue: disabled" log line confirming fail-soft behavior)

**Plan metadata:** (will be committed after this summary)

## Files Created/Modified
- `gohexdump/internal/hue/hue.go` - Complete Hue package: Config struct, LoadConfig (reads /var/lib/hexboard/hue.toml), TurnOn (fire-and-forget PUT to Hue bridge v1 API with 5s timeout)
- `gohexdump/cmd/hexboard/main.go` - Added hueConf field to display struct, LoadConfig call in main() with fail-soft logging, goroutine trigger in showMessage
- `gohexdump/go.mod` - Added github.com/BurntSushi/toml v1.3.2
- `gohexdump/go.sum` - Updated by go get

## Decisions Made
- **BurntSushi/toml v1.3.2:** Pure Go TOML parser (no CGo), ARMv6 compatible, well-established library
- **Config path /var/lib/hexboard/hue.toml:** Matches existing hexboard.db location pattern from Phase 1
- **Fire-and-forget pattern:** `go d.hueConf.TurnOn()` after screenChan push — never blocks message display (HUE-03)
- **5-second timeout on hueClient:** Prevents goroutine leak when bridge is unreachable
- **Fail-soft LoadConfig:** Returns (nil, nil) when file absent; returns (nil, err) when malformed — caller logs and continues with nil hueConf (HUE-04)
- **Nil-safe showMessage:** `if d.hueConf != nil { go d.hueConf.TurnOn() }` — no action when Hue disabled

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None. Build succeeded on first attempt (BurntSushi/toml is pure Go, no CGo compilation required). Deployment via `make deploy` completed in ~2 minutes. Startup logs confirmed fail-soft behavior working correctly.

## User Setup Required

**Hue bridge configuration optional.** To enable Hue integration:

1. Create `/var/lib/hexboard/hue.toml` on the Pi with:
   ```toml
   bridge_ip = "YOUR_BRIDGE_IP"
   api_key   = "YOUR_API_KEY"
   device_id = "YOUR_DEVICE_ID"
   ```
2. Restart service: `sudo systemctl restart hexboard`
3. Verify in logs: `sudo journalctl -u hexboard -n 10 | grep hue`
   - Expected: "hue: enabled (bridge=... device=...)"
4. Send a test message: `echo "test" | nc txt 8080`
   - Expected: Configured light turns on within 5 seconds

If file is absent or incomplete, service starts normally with "hue: disabled" log line.

## Next Phase Readiness

- Phase 2 complete — Hue integration deployed and verified
- Ready for Phase 3 (Drawing) — drawing sends will also trigger Hue light-on via the same showMessage goroutine
- Fire-and-forget pattern established for future integrations (notifications, webhooks, etc.)

**No blockers.** Drawing phase can begin immediately.

## Self-Check: PASSED

All deliverables verified:

**Files created:**
- ✓ gohexdump/internal/hue/hue.go
- ✓ .planning/phases/02-hue/02-01-SUMMARY.md

**Commits present:**
- ✓ 0073f74 (Task 1: Create internal/hue package)
- ✓ 666f83e (Task 2: Wire hue into main.go)
- ✓ 7fcfab6 (Planning docs)

**Dependencies:**
- ✓ BurntSushi/toml in go.mod

---
*Phase: 02-hue*
*Completed: 2026-03-01*
