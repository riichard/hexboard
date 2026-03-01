# Roadmap: Hexboard Web UI Enhancement

## Overview

Five phases deliver persistent history, drawing capability, configurable duration, and Hue smart-switch integration onto an existing Go web app running on a Raspberry Pi Zero. Each phase builds on the previous: the SQLite store lands first (everything else writes to it), Hue wiring comes next, drawing tool follows, duration controls come fourth, and history UI closes the loop.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Store** - SQLite persistence layer for messages and drawings
- [x] **Phase 2: Hue** - Smart-switch config and auto-on integration
- [ ] **Phase 3: Drawing** - Segment picker UI + drawing screen in Go
- [ ] **Phase 4: Duration** - Global default and per-send duration controls
- [ ] **Phase 5: History UI** - Unified history list with replay

## Phase Details

### Phase 1: Store
**Goal**: Every message sent to the display is durably persisted and survives server restarts
**Depends on**: Nothing (first phase)
**Requirements**: HIST-01, HIST-02
**Success Criteria** (what must be TRUE):
  1. Sending a message via the web form writes a record to the SQLite DB on the Pi
  2. After `make deploy` restarts the service, previously sent messages still exist in the DB
  3. The DB file lives at a stable path under `/var/lib/hexboard/` with correct permissions
  4. Concurrent reads and writes do not produce SQLITE_BUSY errors (WAL mode active)
**Plans**: 2 plans

Plans:
- [x] 01-01-PLAN.md — Create internal/store package (OpenDB, Save, Recent) with SQLite + WAL
- [x] 01-02-PLAN.md — Wire store into web handler and main, deploy with systemd ExecStartPre

### Phase 2: Hue
**Goal**: Sending any message or drawing automatically turns on the Hue-connected display light
**Depends on**: Phase 1
**Requirements**: HUE-01, HUE-02, HUE-03, HUE-04
**Success Criteria** (what must be TRUE):
  1. A config file on the Pi holding bridge IP, API key, and device ID is read at startup
  2. Sending a message to the display triggers the Hue device to turn on within a few seconds
  3. A Hue failure (unreachable bridge, bad key) does not delay or block the message from displaying
  4. If the config file is absent or incomplete, the service starts normally with Hue silently disabled
**Plans**: 1 plan

Plans:
- [x] 02-01-PLAN.md — Create internal/hue package, wire into main.go, deploy and verify

### Phase 3: Drawing
**Goal**: Users can compose a segment-level drawing in the browser and send it to the display
**Depends on**: Phase 1
**Requirements**: DRAW-01, DRAW-02, DRAW-03, DRAW-04, DRAW-05
**Success Criteria** (what must be TRUE):
  1. The web UI shows a 4x32 grid of digit cells, each with 16 individually clickable segments
  2. Clicking a segment toggles it on/off and the grid updates visually as a live preview
  3. A "Clear" button resets all segments to off
  4. Sending a drawing pushes it to the physical display for the specified duration
  5. The sent drawing is saved to the SQLite DB alongside text messages
**Plans**: TBD

### Phase 4: Duration
**Goal**: Users control how long any message or drawing stays on the display
**Depends on**: Phase 1
**Requirements**: DUR-01, DUR-02, DUR-03, DUR-04
**Success Criteria** (what must be TRUE):
  1. A settings page lets the user set a global default display duration, which persists across restarts
  2. Each message or drawing send form includes an optional duration override field
  3. When a duration override is given, that duration is used instead of the global default
  4. When no override is given, the global default duration governs how long the display shows the content
**Plans**: TBD

### Phase 5: History UI
**Goal**: Users can browse all past messages and drawings and re-send any of them
**Depends on**: Phase 1, Phase 4
**Requirements**: HIST-03, HIST-04, HIST-05
**Success Criteria** (what must be TRUE):
  1. The web UI shows a unified chronological list of all past messages and drawings from the DB
  2. Each history item has a re-send button that pushes it back to the display
  3. Re-sending a history item prompts the user for a duration (pre-filled with the stored or default duration)
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Store | 2/2 | Complete   | 2026-03-01 |
| 2. Hue | 1/1 | Complete   | 2026-03-01 |
| 3. Drawing | 0/? | Not started | - |
| 4. Duration | 0/? | Not started | - |
| 5. History UI | 0/? | Not started | - |
