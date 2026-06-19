package main

import (
	"log"

	"github.com/dqso/after-the-last/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	const (
		screenWidth  = 800
		screenHeight = 600
	)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("After The Last")
	if err := ebiten.RunGame(game.NewGame(screenWidth, screenHeight)); err != nil {
		log.Fatal(err)
	}
}
