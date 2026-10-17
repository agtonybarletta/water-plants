# Water Plants 🌱

A terminal-based puzzle game written in Go. Navigate a grid, dig cells, and water plants before you run out of shovels!

## Gameplay

You control a shovel (`🪏`) on a grid. Each move:

1. **Digs** the cell you step on (uses one shovel)
2. **Wets** neighboring cells — if a plant is on a wet cell, it's watered

**Goal:** Water all plants (`🌱`) on the grid.

**Lose:** Run out of shovels while plants remain.

### Controls

| Key | Action |
|-----|--------|
| `← ↑ → ↓` | Move the shovel |
| `R` | Reset the current level |
| `Enter` | Advance to the next level (when won) |
| `Esc` / `Ctrl+C` / `Ctrl+D` | Quit |

## Levels

| Level | Grid | Plants |
|-------|------|--------|
| 1 | 5×5 | 1 (fixed) |
| 2 | 5×5 | 2 (fixed) |
| 3 | 5×5 | 3 (fixed) |
| 4 | 6×6 | 4 (fixed) |
| 5 | 7×7 | Random |
| 6+ | 7×7 | Random (endless) |

After completing level 5, new random levels are generated infinitely.

## Architecture

The project follows the **Model-View-Controller (MVC)** pattern:

```
main.go          — Entry point, wires everything together
model.go         — Game state, logic, levels, cells
presenter.go     — Controller: processes input, updates view
terminal_view.go — Terminal rendering with box-drawing characters
view.go          — View interface definition
log.go           — Logging utility
```

### Key types

- **`Game`** — Holds the grid, player position, shovels, plants, and win/loss state
- **`Level`** — A game instance with its plant positions and seed
- **`LevelDef`** — Blueprint for a level (grid size, fixed or random plants)
- **`Model`** — Manages the list of levels and current level index
- **`Presenter`** — Game loop: polls keys, updates model, redraws view
- **`TerminalView`** — Renders the grid using Unicode box-drawing characters and emoji

## Build & Run

```bash
# Build
go build -o water-plants .

# Run
./water-plants
```

Requires Go 1.26.5+ and a terminal with Unicode/emoji support.

## Dependencies

- [`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term) — Terminal raw mode and size detection

## Project Status

This is a work-in-progress game project. See [REVIEW.md](./REVIEW.md) for a detailed code review with suggested improvements.