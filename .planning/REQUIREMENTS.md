# Requirements: Hexboard Web UI Enhancement

**Defined:** 2026-03-01
**Core Value:** Every message or drawing sent from the web UI should reach the display — even if it was off — and be replayable from history indefinitely.

## v1 Requirements

### History & Persistence

- [x] **HIST-01**: All sent messages and drawings are persisted to SQLite DB on the Pi
- [x] **HIST-02**: History survives server restarts
- [ ] **HIST-03**: Web UI displays a unified chronological list of all past messages and drawings
- [ ] **HIST-04**: User can re-send any history item to the display
- [ ] **HIST-05**: Re-sending from history prompts for display duration (default: stored duration or global default)

### Drawing Tool

- [ ] **DRAW-01**: Web UI includes a segment picker: a 4×32 grid of digit cells, each showing the 16 segments of that digit
- [ ] **DRAW-02**: User can click individual segments to toggle them on/off
- [ ] **DRAW-03**: Drawing grid reflects the current state as a visual preview
- [ ] **DRAW-04**: User can clear all segments (reset the grid)
- [ ] **DRAW-05**: User can send a drawing to the display with a specified duration

### Display Duration

- [ ] **DUR-01**: Global default display duration is configurable in the web UI
- [ ] **DUR-02**: Global default is persisted across server restarts
- [ ] **DUR-03**: Each message or drawing send includes an optional duration override field
- [ ] **DUR-04**: If no per-send duration is given, the global default is used

### Hue Integration

- [ ] **HUE-01**: Hue Bridge IP, API key, and device ID are configured via a config file on the Pi
- [ ] **HUE-02**: When any message or drawing is sent to the display, the Hue device is automatically turned on
- [ ] **HUE-03**: Hue API call is non-blocking — a Hue failure does not delay or prevent the message from displaying
- [ ] **HUE-04**: If Hue config is absent or incomplete, the feature is silently disabled (logged, not crashed)

## v2 Requirements

### History

- **HIST-V2-01**: Pagination of history list (for large histories)
- **HIST-V2-02**: Delete individual history entries
- **HIST-V2-03**: Filter history by type (message / drawing)

### Drawing

- **DRAW-V2-01**: Import text into drawing grid (convert font → segments)
- **DRAW-V2-02**: Multi-frame animation drawing

### Hue

- **HUE-V2-01**: Status indicator in web UI (Hue reachable / not)
- **HUE-V2-02**: Manual Hue on/off button in web UI
- **HUE-V2-03**: Auto-off after display returns to idle

## Out of Scope

| Feature | Reason |
|---------|--------|
| Hue auto-discovery | Manual config is simpler; single-device setup |
| Real-time collaboration | Single-user app |
| Cloud sync | Local-only device |
| Mobile-native app | Web UI is sufficient |
| Drawing animations (multi-frame) | Out of scope for this milestone |
| Color control | Monochrome display only |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| HIST-01 | Phase 1 | Complete |
| HIST-02 | Phase 1 | Complete |
| HUE-01 | Phase 2 | Pending |
| HUE-02 | Phase 2 | Pending |
| HUE-03 | Phase 2 | Pending |
| HUE-04 | Phase 2 | Pending |
| DRAW-01 | Phase 3 | Pending |
| DRAW-02 | Phase 3 | Pending |
| DRAW-03 | Phase 3 | Pending |
| DRAW-04 | Phase 3 | Pending |
| DRAW-05 | Phase 3 | Pending |
| DUR-01 | Phase 4 | Pending |
| DUR-02 | Phase 4 | Pending |
| DUR-03 | Phase 4 | Pending |
| DUR-04 | Phase 4 | Pending |
| HIST-03 | Phase 5 | Pending |
| HIST-04 | Phase 5 | Pending |
| HIST-05 | Phase 5 | Pending |

**Coverage:**
- v1 requirements: 18 total
- Mapped to phases: 18
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-01*
*Last updated: 2026-03-01 after roadmap creation*
