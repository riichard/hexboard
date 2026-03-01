# Features Research: Hexboard Web UI Enhancement

**Research type:** Features — subsequent milestone
**Date:** 2026-03-01

---

## History / Persistence

### Table Stakes
- **Persistent storage** — history survives server restarts
- **Unified list** — messages and drawings in one chronological list
- **Re-send from history** — one click to re-display a past item
- **Timestamp display** — show when each item was sent

### Differentiators
- **Pagination** — if history grows large
- **Search/filter** — filter by type (message/drawing)
- **Delete entries** — remove items from history

### Anti-features
- **Sync to cloud** — unnecessary complexity for a local device
- **User accounts** — single-user app
- **Infinite scroll** — overkill for typical usage volume

---

## Drawing Tool

### Table Stakes
- **Segment toggle** — click to toggle individual segments on each of the 128 digit cells
- **Grid layout** — visual 4×32 grid matching physical display layout
- **Preview** — see what the drawing looks like before sending
- **Clear button** — reset all segments
- **Send with duration** — specify display duration when submitting

### Differentiators
- **Fill cell** — set all segments in a cell on at once
- **Clear cell** — turn off all segments in a cell
- **Import text** — populate from a text string (convert font → segments)
- **Save as drawing** — send and save simultaneously (or always save on send)

### Anti-features
- **Animation/multi-frame** — out of scope for now
- **Color picking** — monochrome display only

---

## Configurable Display Duration

### Table Stakes
- **Global default setting** — one setting that applies to all sends
- **Per-send override** — field in send form to override duration for one message
- **Persist global default** — saved across restarts

### Differentiators
- **Per-history-item duration stored** — replay with original duration as default
- **"Stay on" mode** — never return to idle (duration = 0 or infinity)

---

## Hue Integration

### Table Stakes
- **Auto-on on send** — turn on Hue device whenever a message/drawing is sent
- **Non-blocking** — Hue API call must not delay the display response
- **Graceful failure** — if Hue unreachable, send still works (log error only)
- **Config file** — bridge IP and API key not hardcoded

### Differentiators
- **Status indicator in UI** — show whether Hue is reachable
- **Manual on/off button** — separate control in UI

### Anti-features
- **Auto-off** — turning off after timeout is complex and not requested
- **Scene control** — out of scope
- **Multi-switch** — single device for now

---

## Feature Dependencies

```
SQLite schema → History list → Re-send → Drawing replay
Drawing tool → Drawing send → SQLite store → History list
Global timeout setting → Per-send override → SQLite store (duration field)
Hue config → Hue client → Auto-on on send
```
