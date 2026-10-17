package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	"golang.org/x/term"
)

// Byte decoding |left, up, right, down| 0000
var BOX_ALPHA []rune = []rune{
	' ', //0000
	'╴', //0001
	'╵', //0010
	'┘', //0011
	'╶', //0100
	'─', //0101
	'└', //0110
	'┴', //0111
	'╷', //1000
	'┐', //1001
	//'|', //1010
	//'|', //1010
	'│', //1010
	'┤', //1011
	'┌', //1100
	'┬', //1101
	'├', //1110
	'┼', //1111
}
var BOX_HOR = 5
var BOX_VER = 10
var BOX_EMPTY = 0
var BOX_CROSS = 15

type TerminalView struct {
	width     int
	height    int
	tileChars [][][]rune
	horTiles  int
	verTiles  int
	title     string
	subtitle  string
	buffer    [][]rune
	style     struct {
		tile_width  int
		tile_height int
	}
	keysMu sync.Mutex
	keys   []Key
}

func (t *TerminalView) SetSize(width, height int) {
	t.width = width
	t.height = height
	buffer := make([][]rune, height+1)
	for i := range buffer {
		buffer[i] = make([]rune, width+1)
	}
	t.buffer = buffer
}

func (t *TerminalView) SetGridTiles(horTiles, verTiles int) {
	t.horTiles = horTiles
	t.verTiles = verTiles
	t.tileChars = make([][][]rune, verTiles)
	for i := range t.tileChars {
		t.tileChars[i] = make([][]rune, horTiles)
	}
}

func (t *TerminalView) SetTileChars(x, y int, chars []rune) {
	if x < 0 || x >= t.horTiles || y < 0 || y >= t.verTiles {
		return
	}
	t.tileChars[y][x] = chars
}

func (t *TerminalView) SetTitle(title string) {
	t.title = title
}

func (t *TerminalView) SetSubtitle(subtitle string) {
	t.subtitle = subtitle
}

func (t *TerminalView) Draw() {
	t.updateSize()
	t.clearScreen()
	if t.canDraw() {
		t.drawTitle()
		t.drawSubtitle()
		t.drawGrid()
		t.drawTileChars()
		t.flushScreen()
	} else {
		t.flushScreen()
		t.printError()
	}
}

func (t *TerminalView) drawTitle() {
	if t.title == "" {
		return
	}
	runes := []rune(t.title)
	x0 := (t.width - len(runes)) / 2
	if x0 < 0 {
		x0 = 0
	}
	for i, ch := range runes {
		px := x0 + i
		if px < t.width {
			t.buffer[0][px] = ch
		}
	}
}

func (t *TerminalView) drawSubtitle() {
	if t.subtitle == "" {
		return
	}
	runes := []rune(t.subtitle)
	x0 := (t.width - len(runes)) / 2
	if x0 < 0 {
		x0 = 0
	}
	for i, ch := range runes {
		px := x0 + i
		if px < t.width {
			t.buffer[1][px] = ch
		}
	}
}

func (t *TerminalView) Init() {

	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		GetLogger().Warn("Could not get terminal size, using defaults", "error", err)
		w = 12
		h = 12
	}

	t.SetSize(w, h)
	GetLogger().Debug("Screen size: ", "w", t.width, "h", t.height)

	t.style.tile_height = 2
	t.style.tile_width = 8

	// Switch to alternate screen buffer and hide cursor
	fmt.Print("\033[?1049h\033[?25l")

	go t.readKeys()
}

func (t *TerminalView) Exit() {
	// Show cursor and switch back to main screen buffer
	fmt.Print("\033[?25h\033[?1049l")
}

func (t *TerminalView) PollKeys() []Key {
	t.keysMu.Lock()
	defer t.keysMu.Unlock()
	keys := t.keys
	t.keys = nil
	return keys
}

func (t *TerminalView) readKeys() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		GetLogger().Warn("Could not set terminal to raw mode", "error", err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}
		t.keysMu.Lock()
		if n == 1 {
			switch buf[0] {
			case 'r', 'R':
				t.keys = append(t.keys, KeyReset)
			case 0x0d: // Enter (carriage return)
				t.keys = append(t.keys, KeyEnter)
			case 0x1b: // ESC
				t.keys = append(t.keys, KeyQuit)
				t.keysMu.Unlock()
				return
			case 0x03: // Ctrl+C
				t.keys = append(t.keys, KeyQuit)
				t.keysMu.Unlock()
				return
			case 0x04: // Ctrl+D
				t.keys = append(t.keys, KeyQuit)
				t.keysMu.Unlock()
				return
			}
		} else if n == 3 && buf[0] == 0x1b && buf[1] == '[' {
			switch buf[2] {
			case 'A':
				t.keys = append(t.keys, KeyUp)
			case 'B':
				t.keys = append(t.keys, KeyDown)
			case 'C':
				t.keys = append(t.keys, KeyRight)
			case 'D':
				t.keys = append(t.keys, KeyLeft)
			}
		}
		t.keysMu.Unlock()
	}
}

func (t *TerminalView) updateSize() {
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		GetLogger().Warn("Could not get terminal size, using defaults", "error", err)
	}

	t.SetSize(w, h)
}

/* check if the grid fits within the screen
 */
func (t *TerminalView) canDraw() bool {
	fitHor := t.gridWidth() < t.width
	fitVer := t.gridHeight() < t.height
	return fitHor && fitVer
}

func (t *TerminalView) printError() {
	fmt.Println("Can't draw, screen too small.")
}

func (t *TerminalView) flushScreen() {
	var sb strings.Builder
	// Pre-allocate enough space for the whole screen
	sb.Grow(t.width * t.height * 4)
	// Clear screen
	sb.WriteString("\033[2J")
	for y := 0; y < t.height; y++ {
		for x := 0; x < t.width; x++ {
			// Position cursor to the exact cell before writing, so double-width
			// emoji never misalign subsequent single-width characters.
			sb.WriteString(fmt.Sprintf("\033[%d;%dH", y+1, x+1))
			sb.WriteRune(t.buffer[y][x])
		}
	}
	fmt.Print(sb.String())
}

func (t *TerminalView) clearScreen() {
	for i := 0; i < t.height; i++ {
		for j := 0; j < t.width; j++ {
			t.buffer[i][j] = BOX_ALPHA[BOX_EMPTY]
		}
	}
}

func (t *TerminalView) gridWidth() int {
	return t.horTiles*t.style.tile_width + (t.horTiles - 1)
}

func (t *TerminalView) gridHeight() int {
	return t.verTiles*t.style.tile_height + (t.verTiles - 1)
}

/*
 * Update box char in position x,y based on neighbors
 */
func (t *TerminalView) updateBoxChar(x int, y int) error {

	if x < 0 || x >= t.width || y < 0 || y >= t.height {
		return nil
	}
	if t.buffer[y][x] == BOX_ALPHA[BOX_EMPTY] {
		return nil
	}

	/* Next char, index of BOX_ALPHA */
	newCharIndex := 0

	var dx = []int{-1, 0, 1, 0}
	var dy = []int{0, -1, 0, 1}

	for i := range dx {
		nx := x + dx[i]
		ny := y + dy[i]

		if nx >= 0 && nx < t.width && ny >= 0 && ny < t.height && t.buffer[ny][nx] != BOX_ALPHA[BOX_EMPTY] {
			// add component to newCharIndex for the current neighbor (nx, ny)
			newCharIndex |= 1 << i
		}
	}

	t.buffer[y][x] = BOX_ALPHA[newCharIndex]
	return nil
}

/*
 * Draw box char c in position x,y
 */
func (t *TerminalView) drawBoxChar(x int, y int, c rune) error {

	/* current index in boxChars of the char in position x,y */
	alphIdx := slices.Index(BOX_ALPHA, t.buffer[y][x])

	if alphIdx == -1 {
		return fmt.Errorf("Invalid box char at position (%d,%d): %c", x, y, t.buffer[y][x])
	}

	// if c is empty, erase it
	// otherwise, draw the temp char (cross).
	// This will be updated later to the correct char based on neighbors.
	if c == BOX_ALPHA[BOX_EMPTY] {
		t.buffer[y][x] = BOX_ALPHA[BOX_EMPTY]
	} else {
		t.buffer[y][x] = BOX_ALPHA[BOX_CROSS]
	}

	// update neighbors to reflect the new character
	var dx = []int{-1, 0, 1, 0}
	var dy = []int{0, -1, 0, 1}
	var newCharIndex int = 0
	for i := range dx {
		nx := x + dx[i]
		ny := y + dy[i]
		if nx >= 0 && nx < t.width && ny >= 0 && ny < t.height {
			if t.buffer[ny][nx] != BOX_ALPHA[BOX_EMPTY] {
				newCharIndex |= 1 << i
				t.updateBoxChar(nx, ny)

			}
		}
	}

	// if new char is empty but c is not, then all neighbors are empty, so we can just draw c
	if BOX_ALPHA[newCharIndex] == BOX_ALPHA[BOX_EMPTY] && c != BOX_ALPHA[BOX_EMPTY] {
		t.buffer[y][x] = c
	} else {
		t.buffer[y][x] = BOX_ALPHA[newCharIndex]
	}
	return nil
}

func (t *TerminalView) drawBoxLine(x1, y1, x2, y2 int) {

	// diagonal line not supported
	if x1 != x2 && y1 != y2 {
		return
	}

	// check for out of bounds
	if x1 < 0 && x2 < 0 || x1 >= t.width && x2 >= t.width || y1 < 0 && y2 < 0 || y1 >= t.height && y2 >= t.height {
		return
	}

	if x1 == x2 {
		for y := y1; y <= y2; y++ {
			t.drawBoxChar(x1, y, BOX_ALPHA[BOX_VER])
		}
	} else if y1 == y2 {
		for x := x1; x <= x2; x++ {
			t.drawBoxChar(x, y1, BOX_ALPHA[BOX_HOR])
		}
	}
}

func (t *TerminalView) drawGrid() {

	var gridY0 = (t.height - t.gridHeight()) / 2
	var gridX0 = (t.width - t.gridWidth()) / 2

	for i := 0; i < t.horTiles; i++ {
		for j := 0; j < t.verTiles; j++ {
			boxX0 := gridX0 + i*(t.style.tile_width+1)
			boxY0 := gridY0 + j*(t.style.tile_height+1)

			t.drawBoxLine(boxX0, boxY0, boxX0+t.style.tile_width+1, boxY0)
			t.drawBoxLine(boxX0, boxY0, boxX0, boxY0+t.style.tile_height+1)
			t.drawBoxLine(boxX0+t.style.tile_width+1, boxY0, boxX0+t.style.tile_width+1, boxY0+t.style.tile_height+1)
			t.drawBoxLine(boxX0, boxY0+t.style.tile_height+1, boxX0+t.style.tile_width+1, boxY0+t.style.tile_height+1)

		}
	}
}

func (t *TerminalView) drawTileChars() {
	var gridY0 = (t.height - t.gridHeight()) / 2
	var gridX0 = (t.width - t.gridWidth()) / 2

	for i := 0; i < t.horTiles; i++ {
		for j := 0; j < t.verTiles; j++ {
			chars := t.tileChars[j][i]
			if len(chars) == 0 {
				continue
			}

			boxX0 := gridX0 + i*(t.style.tile_width+1) + 1
			boxY0 := gridY0 + j*(t.style.tile_height+1) + 1

			// Draw characters inside the tile, two per row
			for k, ch := range chars {
				if k >= 4 {
					break
				}
				charX := boxX0 + (k%2)*2
				charY := boxY0 + (k / 2)
				t.buffer[charY][charX] = ch
			}
		}
	}
}
