package game

import (
	"fmt"
	"image/color"
	"log"

	"github.com/dqso/after-the-last/assets"
	"github.com/dqso/after-the-last/tileset"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const tileSize = 16

// mapGrid is a 9x7 tile map; all cells use tile index 182.
// TODO it is temporary
var mapGrid = func() [][]int {
	const (
		mapCols = 9
		mapRows = 7
		tile    = 182
	)
	grid := make([][]int, mapRows)
	for r := range grid {
		grid[r] = make([]int, mapCols)
		for c := range grid[r] {
			grid[r][c] = tile
		}
	}
	return grid
}()

type Game struct {
	screenWidth  int
	screenHeight int
	player       *Player
	camera       *Camera
	world        *World
}

func NewGame(screenWidth, screenHeight int) *Game {
	roomTS, err := tileset.New(assets.RoomBuilderSheet, tileSize, tileSize)
	if err != nil {
		log.Fatal(err)
	}
	bobIdleTS, err := tileset.New(assets.BobIdle, tileSize, tileSize*2)
	if err != nil {
		log.Fatal(err)
	}
	bobRunTS, err := tileset.New(assets.BobRun, tileSize, tileSize*2)
	if err != nil {
		log.Fatal(err)
	}

	w := NewWorld(roomTS, mapGrid)

	cx, cy := w.CellCenter(5, 5)
	return &Game{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		player:       NewPlayer(bobIdleTS, bobRunTS, cx, cy),
		camera:       NewCamera(),
		world:        w,
	}
}

func (g *Game) Update() error {
	g.player.Update()
	g.camera.Follow(g.player.X, g.player.Y)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Draw(screen, g.camera.X, g.camera.Y, g.screenWidth, g.screenHeight)
	g.player.Draw(screen, g.camera, g.screenWidth, g.screenHeight)
	g.drawWorldAxes(screen)

	playerSX, playerSY := g.camera.WorldToScreen(g.player.X, g.player.Y, g.screenWidth, g.screenHeight)
	info := fmt.Sprintf(
		"World:  player (%.1f, %.1f)\nCamera: center (%.1f, %.1f)\nScreen: player (%.1f, %.1f)",
		g.player.X, g.player.Y,
		g.camera.X, g.camera.Y,
		playerSX, playerSY,
	)
	ebitenutil.DebugPrint(screen, info)
}

// drawWorldAxes draws X/Y axes intersecting at world origin (0,0).
func (g *Game) drawWorldAxes(screen *ebiten.Image) {
	sw, sh := float64(g.screenWidth), float64(g.screenHeight)

	// origin in screen space
	ox, oy := g.camera.WorldToScreen(0, 0, g.screenWidth, g.screenHeight)

	red := color.RGBA{R: 255, A: 255}
	green := color.RGBA{G: 255, A: 255}

	vector.StrokeLine(screen, 0, float32(oy), float32(sw), float32(oy), Scale, red, false)
	vector.StrokeLine(screen, float32(ox), 0, float32(ox), float32(sh), Scale, green, false)
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	g.screenWidth, g.screenHeight = outsideWidth, outsideHeight
	return outsideWidth, outsideHeight
}
