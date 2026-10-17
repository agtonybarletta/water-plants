package main

import (
	"fmt"
)

func main() {
	fmt.Println("Welcome to Water plants!")

	view := &TerminalView{}
	view.Init()
	defer view.Exit()

	model := &Model{}
	model.Init()

	presenter := &Presenter{}
	presenter.Init(view, model, 15)

	presenter.Run()
}
