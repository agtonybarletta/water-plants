# Code Review: Water Plants

## Overview

A solid terminal-based MVC game written in Go. The code is generally well-structured with clear separation of concerns (Model, View, Presenter). However, there are several issues ranging from critical bugs to style concerns that should be addressed before this code is considered production-ready.

---

## Critical Issues

### 1. `wetNeighborCells` has a logic bug — it wets a 4×4 area instead of the 4 cardinal neighbors

**File:** `model.go`, method `wetNeighborCells`

```go
func (m *Model) wetNeighborCells(x, y int) {
    d := []int{-1, 0, 1, 0}
    for i := 0; i < 4; i++ {
        for j := 0; j < 4; j++ {
            nx := x + d[i]
            ny := y + d[j]
            m.wetCell(nx, ny)
        }
    }
}
```

This double loop iterates over **all 16 combinations** of `d[i]` × `d[j]`, wetting a 4×4 block of cells. The intent is clearly to wet only the 4 cardinal neighbors (up, down, left, right). This should be a single loop:

```go
func (m *Model) wetNeighborCells(x, y int) {
    d := []int{-1, 0, 1, 0}
    for i := 0; i < 4; i++ {
        nx := x + d[i]
        ny := y + d[(i+2)%4]  // or use dx/dy pairs
        m.wetCell(nx, ny)
    }
}
```

### 2. `wetCell` wets cells that are `digged` — contrary to the condition

**File:** `model.go`, method `wetCell`

```go
if !game.cells[x][y].digged && !game.cells[x][y].wet {
```

This means a **digged** cell cannot be wet, but a **planted** cell can be wet (and when it is, `plantsLeft` decrements). However, the start cell is already `digged = true`, so it can never be wet. This is likely intentional, but the real issue: you can only water a plant by wetting its cell **before** digging it. Since you dig the cell you step on, you can never water a plant on the same cell. Verify the game logic is correct.

### 3. `digCell` decrements shovels below zero

**File:** `model.go`, method `digCell`

```go
if !game.cells[x][y].digged {
    game.shovels -= 1
    ...
}
```

If `shovels` is already 0, this goes negative. The check `game.shovels == 0 && game.plantsLeft > 0` sets `StateLost`, but that happens **after** the decrement, so `shovels` becomes -1. Either check `shovels > 0` before decrementing, or guard against digging with no shovels.

### 4. `GetState()` is called once per frame, then the state is used in a stale way

**File:** `presenter.go`, method `processInput`

```go
state := c.model.GetState()
for _, key := range c.view.PollKeys() {
    ...
    case KeyEnter:
        if state == StateWon {
            c.model.NextLevel()
        }
```

`state` is captured once at the top of the function. If multiple keys are polled in one frame (e.g., Enter then Left), the stale state check might allow movement when it shouldn't, or block movement after advancing. Re-check state inside the loop, or process one key per frame.

### 5. `Reset()` for predefined levels re-initializes from `LevelDef` — losing the current seed

**File:** `model.go`, method `Reset`

```go
if m.currentLevel < len(levelDefs) {
    m.levels[m.currentLevel].Init(levelDefs[m.currentLevel])
} else {
    m.ResetFromSeed()
}
```

For predefined levels, `Reset` calls `Level.Init` which generates a **new random seed** for random levels — so resetting level 5 gives a different layout. For `StateWon`/`StateLost`, the presenter calls `ResetFromSeed` (preserves layout), but during normal play `Reset` reshuffles. This inconsistency may be intentional, but it's worth noting.

---

## Medium Issues

### 6. `LevelDef` is a struct but `levelDefs` is a global variable

**File:** `model.go`

Global mutable state is generally discouraged. If test code modifies `levelDefs`, behavior changes globally. Consider making it a `const`-like private variable or a function that returns levels.

### 7. Dead code: the commented-out `move` function

**File:** `model.go`, lines 158–177

A large commented-out block with broken syntax. Remove it.

### 8. `CHAR_*` constants are `var` instead of `const`

**File:** `presenter.go`

```go
var CHAR_WAVE = '🌊'
```

These are never reassigned. Go allows `const` for rune literals:

```go
const CHAR_WAVE = '🌊'
```

### 9. `Box drawing` constants are `var` instead of `const`

**File:** `terminal_view.go`

```go
var BOX_HOR = 5
var BOX_VER = 10
var BOX_EMPTY = 0
var BOX_CROSS = 15
```

These should be `const`.

### 10. `GetLogger()` uses a global variable with lazy initialization

**File:** `log.go`

```go
var Logger *slog.Logger
func GetLogger() *slog.Logger {
    if Logger == nil {
        ...
    }
    return Logger
}
```

This is not thread-safe (two goroutines could both see `Logger == nil`). Use `sync.Once` or initialize in `init()`.

### 11. `drawGrid()` has redundant bounds checking in `drawBoxLine`

**File:** `terminal_view.go`

The bounds check in `drawBoxLine` is convoluted:

```go
if x1 < 0 && x2 < 0 || x1 >= t.width && x2 >= t.width || y1 < 0 && y2 < 0 || y1 >= t.height && y2 >= t.height {
    return
}
```

Due to operator precedence (`&&` before `||`), this works but is confusing. Parentheses would help. Also, it only rejects if **both** endpoints are out of bounds. A single point out of bounds will still draw partially.

### 12. `flushScreen()` writes every cell every frame

**File:** `terminal_view.go`

```go
for y := 0; y < t.height; y++ {
    for x := 0; x < t.width; x++ {
        sb.WriteString(fmt.Sprintf("\033[%d;%dH", y+1, x+1))
        sb.WriteRune(t.buffer[y][x])
    }
}
```

This writes the entire screen buffer every frame, even cells that haven't changed. For a 212×50 terminal, that's 10,600 ANSI escape sequences per frame. You could track dirty regions and only write changed cells. This is likely fine for now, but keep an eye on performance.

---

## Minor Issues

### 13. `for true` should be `for`

**File:** `presenter.go`, method `Run`

```go
for true {
```

Go idiom: `for { ... }`.

### 14. `NextLevel()` always returns `true` but callers ignore the return value

**File:** `model.go`, method `NextLevel`, and `presenter.go`

The return value is never used. Consider removing it.

### 15. `Init()` methods don't return errors

**File:** `model.go`, `presenter.go`, `terminal_view.go`

If `Init` fails (e.g., terminal size), the state is silently partially-initialized. Consider returning `error`.

### 16. `Draw()` swallows errors from `drawBoxChar`

**File:** `terminal_view.go`

`drawBoxChar` returns an error for invalid box characters, but `drawBoxLine` (and by extension `drawGrid`) ignores it. Log the error at minimum.

### 17. `SetSize` allocates a buffer one row/column too large

**File:** `terminal_view.go`

```go
buffer := make([][]rune, height+1)
for i := range buffer {
    buffer[i] = make([]rune, width+1)
}
```

The `+1` is unnecessary — the buffer is indexed `[0..height-1]` and `[0..width-1]`. This wastes a small amount of memory.

### 18. `rand.Int()` is used without seeding the global source

**File:** `model.go`

```go
l.seed = rand.Int()
```

`math/rand`'s default source is seeded with a constant (1) in Go 1.20+. If you're on an older Go, you'd get the same sequence every run. Your code explicitly seeds a local `rand.New(rand.NewSource(int64(l.seed)))` for the actual random generation, so this is fine. Just be aware that the **global seed** for `rand.Int()` on line `l.seed = rand.Int()` depends on Go version.

### 19. `readKeys()` leaks the goroutine on error

**File:** `terminal_view.go`

```go
go t.readKeys()
```

If `term.MakeRaw` fails, the goroutine returns immediately. This is fine. But if the program exits without calling `Exit()`, the terminal stays in raw mode. Consider a `defer` in `main()`.

### 20. `BOX_ALPHA` has a confusing comment with commented-out code

**File:** `terminal_view.go`

```go
//'|', //1010
//'|', //1010
'│', //1010
```

Remove the dead comments.

---

## Architecture & Design Suggestions

### 21. `Level` and `LevelDef` are separate but tightly coupled

`LevelDef` is a "blueprint" and `Level` is a "runtime instance" — a good pattern. But `Level` exposes `game` as a public field (uppercase `Game`), and `Model` accesses it directly. Consider making `game` private and exposing methods.

### 22. `Game` has a mixed responsibility

`Game` holds both state (`cells`, `shovels`, `plantsLeft`) and logic (`init`, `Init`). The `init`/`Init` split is clean. But `shovels` calculation (`horTiles + verTiles`) is a magic formula living inside `init`. Consider making it a parameter or a named constant.

### 23. View interface is large

`View` has 9 methods. Consider splitting into smaller interfaces (`Drawer`, `InputHandler`, etc.) for better testability.

### 24. No tests

There are zero test files. The game logic in `model.go` is well-isolated and would be easy to test with standard Go table-driven tests.

### 25. No graceful handling of terminal resize

`updateSize()` is called every frame, but the grid is recalculated from scratch. The `SetSize` method reallocates the entire buffer. On resize, the grid position shifts, but the game state is preserved. This is acceptable for a terminal game.

---

## Summary

**Must fix before approval:**
1. Fix `wetNeighborCells` — wets a 4×4 block instead of 4 neighbors
2. Prevent `shovels` from going negative
3. Remove dead commented-out code
4. Fix stale `state` variable in `processInput`

**Should fix:**
5. Use `const` for `CHAR_*` and box drawing constants
6. Thread-safe logger initialization
7. Add tests for game logic
8. Clarify `wetCell` / `digCell` game logic

**Nice to have:**
9. Dirty-region rendering for performance
10. Flatten `for true` to `for`
11. Remove unused `NextLevel()` return value