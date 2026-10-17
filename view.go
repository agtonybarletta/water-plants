package main

type Key int

const (
	KeyNone Key = iota
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
	KeyReset
	KeyEnter
	KeyQuit
)

type View interface {
	SetSize(width, height int)
	SetGridTiles(horTiles, verTiles int)
	SetTileChars(x, y int, chars []rune)
	SetTitle(title string)
	SetSubtitle(subtitle string)
	Init()
	Draw()
	PollKeys() []Key
}
