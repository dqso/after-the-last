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

const (
	tileSize          = 16
	playerCollisionRX = tileSize / 2
	playerCollisionRY = 2.5
)

const none = -1

// TODO: temporary hardcoded maps.
var floorGrid = [][]int{
	{none, none, none, none, none, none, none, none, none},
	{none, none, none, none, none, none, none, none, none},
	{none, 164, 165, 165, 165, 165, 165, 165, none},
	{none, 181, 182, 182, 182, 182, 182, 182, none},
	{none, 181, 182, 182, 182, 182, 182, 182, none},
	{none, 181, 182, 182, 182, 182, 182, 182, none},
	{none, 181, 182, 182, 182, 182, 182, 182, none},
	{none, none, none, none, none, none, none, none, none},
}

var wallsGrid = [][]int{
	{28, 324, 324, 324, 324, 324, 324, 324, 30},
	{7, 348, 348, 348, 348, 348, 348, 348, 8},
	{7, none, none, none, none, none, none, none, 8},
	{7, none, none, none, none, none, none, none, 8},
	{7, none, none, none, none, none, none, none, 8},
	{7, none, none, none, none, none, none, none, 8},
	{7, none, none, none, none, none, none, none, 8},
	{62, 63, 63, 63, 63, 63, 63, 63, 64},
}

var itemsGrid = [][]int{
	{none, none, none, none, none, none, none, none, none},
	{none, none, none, none, none, none, none, none, none},
	{none, none, none, none, none, none, none, none, none},
	{none, none, 610, 611, none, none, none, none, none},
	{none, none, 626, 627, none, none, none, none, none},
	{none, none, none, none, none, none, none, none, none},
	{none, none, none, none, none, none, none, none, none},
	{none, none, none, none, none, none, none, none, none},
}

// CollFree=0, CollBlocked=1, see collision.go for partial types.
var collisionGrid = [][]CollisionType{
	{1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 5, 4, 0, 0, 0, 0, 1},
	{1, 0, 3, 2, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 7, 7, 7, 7, 7, 7, 7, 1},
}

type Game struct {
	screenWidth  int
	screenHeight int
	camera       *Camera
	world        *World
}

func NewGame(screenWidth, screenHeight int) *Game {
	roomTS, err := tileset.New(assets.RoomBuilderSheet, tileSize, tileSize)
	if err != nil {
		log.Fatal(err)
	}
	interiorsTS, err := tileset.New(assets.InteriorsSheet, tileSize, tileSize)
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

	player := NewPlayer(bobIdleTS, bobRunTS, playerCollisionRX, playerCollisionRY)
	w := NewWorld(roomTS, floorGrid, wallsGrid, interiorsTS, itemsGrid, collisionGrid, player)
	player.SetPosition(w.CellCenter(5, 5))

	return &Game{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		camera:       NewCamera(),
		world:        w,
	}
}

func (g *Game) Update() error {
	g.world.Update()
	g.camera.Follow(g.world.Player().X, g.world.Player().Y)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Draw(screen, g.camera, g.screenWidth, g.screenHeight)
	//g.drawWorldAxes(screen)

	p := g.world.Player()
	sx, sy := g.camera.WorldToScreen(p.X, p.Y, g.screenWidth, g.screenHeight)
	info := fmt.Sprintf(
		"World:  player (%.1f, %.1f)\nCamera: center (%.1f, %.1f)\nScreen: player (%.1f, %.1f)",
		p.X, p.Y,
		g.camera.X, g.camera.Y,
		sx, sy,
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
