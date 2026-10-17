package main

import (
	"fmt"
	"strings"
	"time"
)

var CHAR_WAVE = '🌊'
var CHAR_PLANT = '🌱'
var CHAR_DROP = '💦'
var CHAR_SHAVEL = '🪏'

type Presenter struct {
	fps      int
	view     View
	model    *Model
	subtitle string
}

func (c *Presenter) Init(view View, model *Model, fps int) {

	c.fps = fps
	c.view = view
	c.model = model

}

func (c *Presenter) SetSubtitle(subtitle string) {
	c.subtitle = subtitle
}

func (c *Presenter) Run() {
	for true {
		if c.processInput() {
			return
		}
		c.updateViewFromModel()
		c.view.Draw()
		time.Sleep(time.Second / time.Duration(c.fps))
	}

}

// processInput returns true if the presenter should quit.
func (c *Presenter) processInput() bool {
	state := c.model.GetState()

	for _, key := range c.view.PollKeys() {
		switch key {
		case KeyLeft:
			if state == StatePlaying {
				c.model.MoveLeft()
			}
		case KeyRight:
			if state == StatePlaying {
				c.model.MoveRight()
			}
		case KeyUp:
			if state == StatePlaying {
				c.model.MoveUp()
			}
		case KeyDown:
			if state == StatePlaying {
				c.model.MoveDown()
			}
		case KeyReset:
			if state == StateWon || state == StateLost {
				c.model.ResetFromSeed()
			} else {
				c.model.Reset()
			}
		case KeyEnter:
			if state == StateWon {
				c.model.NextLevel()
			}
		case KeyQuit:
			return true
		}
	}
	return false
}

func (c *Presenter) updateViewFromModel() {
	model := c.model
	view := c.view
	horTiles, verTiles := model.GetCurrentTilesSize()
	// model: horTiles=rows, verTiles=cols; view: horTiles=cols, verTiles=rows
	view.SetGridTiles(verTiles, horTiles)
	view.SetTitle(c.getTitle())
	state := model.GetState()
	switch state {
	case StateWon:
		c.SetSubtitle("You win! Press Enter for next level")
	case StateLost:
		c.SetSubtitle("Game over! Press R to try again")
	case StatePlaying:
		c.SetSubtitle("")
		if model.GetCurrentLevelNumber() == 1 {
			c.SetSubtitle("Dig near the plants to water them.")
		}
		if model.GetCurrentLevelNumber() == 2 {
			c.SetSubtitle("Water all the plants to move to the next level.")
		}
		if model.GetCurrentLevelNumber() == 3 {
			c.SetSubtitle("Each time you dig, you consume shovels")
		}
	}
	view.SetSubtitle(c.subtitle)
	cells := model.GetCurrentLevel()
	for i := 0; i < horTiles; i++ {
		for j := 0; j < verTiles; j++ {
			cell := cells[i][j]
			chars := make([]rune, 0, 3)
			if cell.digged {
				chars = append(chars, CHAR_WAVE)
			}

			if cell.planted {
				chars = append(chars, CHAR_PLANT)
			}

			if cell.wet {
				chars = append(chars, CHAR_DROP)
			}
			if currRow, currCol := model.GetCurrentPosition(); currRow == i && currCol == j {
				chars = append(chars, CHAR_SHAVEL)
			}
			view.SetTileChars(j, i, chars)
		}
	}
}

func (p *Presenter) getTitle() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Level %d", p.model.GetCurrentLevelNumber()))
	sb.WriteString(" | ")
	sb.WriteString(fmt.Sprintf("%s x %d", string(CHAR_SHAVEL), p.model.GetCurrentShovels()))
	sb.WriteString(" | ")
	sb.WriteString(fmt.Sprintf("%s x %d", string(CHAR_PLANT), p.model.GetCurrentPlantsLeft()))
	sb.WriteString(" | ")
	return sb.String()
}
