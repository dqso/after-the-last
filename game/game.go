package game

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/dqso/after-the-last/assets"
	"github.com/dqso/after-the-last/random"
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

type Game struct {
	screenWidth  int
	screenHeight int
	camera       *Camera
	world        *World
	fov          *FOVRenderer
}

func NewGame(screenWidth, screenHeight int) *Game {
	tilesetList := tileset.NewTileSetList()
	roomTS, err := tileset.NewTileSet(assets.RoomBuilderSheet, tileSize, tileSize)
	if err != nil {
		log.Fatal(err)
	}
	if err := tilesetList.Register(tileset.RoomBuilderSheet, roomTS); err != nil {
		log.Fatal(err)
	}
	interiorsTS, err := tileset.NewTileSet(assets.InteriorsSheet, tileSize, tileSize)
	if err != nil {
		log.Fatal(err)
	}
	if err := tilesetList.Register(tileset.InteriorsSheet, interiorsTS); err != nil {
		log.Fatal(err)
	}
	extraTS, err := tileset.NewTileSet(assets.ExtraSheet, tileSize, tileSize)
	if err != nil {
		log.Fatal(err)
	}
	if err := tilesetList.Register(tileset.ExtraSheet, extraTS); err != nil {
		log.Fatal(err)
	}

	tilesetCharacterList := tileset.NewTileSetList()
	bobIdleTS, err := tileset.NewTileSet(assets.BobIdle, tileSize, tileSize*2)
	if err != nil {
		log.Fatal(err)
	}
	if err := tilesetCharacterList.Register(tileset.BobIdleSheet, bobIdleTS); err != nil {
		log.Fatal(err)
	}
	bobRunTS, err := tileset.NewTileSet(assets.BobRun, tileSize, tileSize*2)
	if err != nil {
		log.Fatal(err)
	}
	if err := tilesetCharacterList.Register(tileset.BobRunSheet, bobRunTS); err != nil {
		log.Fatal(err)
	}

	generator, err := random.NewGenerator(assets.TilesXML)
	if err != nil {
		log.Fatal(err)
	}

	player := NewPlayer(tilesetCharacterList, playerCollisionRX, playerCollisionRY)

	//floorGrid, wallsGrid, itemsGrid, collisionGrid := generator.GenerateV0()
	floorGrid, wallsGrid, itemsGrid, collisionGrid := generator.GenerateV1(time.Now().UnixNano())
	w := NewWorld(tilesetList, floorGrid, wallsGrid, itemsGrid, collisionGrid, player)

	col, row, ok := w.FindFreeCell(rand.New(rand.NewSource(0)))
	if !ok {
		log.Fatal("no free cell found in world")
	}
	player.SetPosition(w.CellCenter(col, row))

	fov, err := NewFOVRenderer()
	if err != nil {
		log.Fatal(err)
	}

	return &Game{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		camera:       NewCamera(),
		world:        w,
		fov:          fov,
	}
}

func (g *Game) Update() error {
	g.world.Update()
	g.camera.Follow(g.world.Player().X, g.world.Player().Y)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Draw(screen, g.camera, g.screenWidth, g.screenHeight)

	p := g.world.Player()
	sx, sy := g.camera.WorldToScreen(p.X, p.Y, g.screenWidth, g.screenHeight)

	eyeWX, eyeWY := p.EyeWorldPos()
	eyeSX, eyeSY := g.camera.WorldToScreen(eyeWX, eyeWY, g.screenWidth, g.screenHeight)
	g.fov.Draw(screen, eyeSX, eyeSY, p.DirAngle())

	//g.drawWorldAxes(screen)

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
