package game

import (
	"fmt"
	"image/color"
	"log"
	"math"
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
	playerCollisionRX = tileSize / 3
	playerCollisionRY = 2.5
)

type Game struct {
	version          string
	screenWidth      int
	screenHeight     int
	camera           *Camera
	world            *World
	fov              *FOVRenderer
	memory           *MemoryLayer
	visibility       *VisibilityRenderer
	worldColor       *ebiten.Image // world-pixel snapshot of the live world for the memory layer
	visibilityWorld  *ebiten.Image // world-pixel line-of-sight mask (visibility pass output)
	memoryScreen     *ebiten.Image // screen-sized projection of memory for the FOV shader
	visibilityScreen *ebiten.Image // screen-sized projection of the line-of-sight mask
	lastDraw         time.Time
}

func NewGame(version string, screenWidth, screenHeight int) *Game {
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
	floorGrid, wallsGrid, itemsGrid, collisionGrid := generator.GenerateV2(time.Now().UnixNano())
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

	worldPixW, worldPixH := w.WorldPixelSize()
	mem, err := NewMemoryLayer(worldPixW, worldPixH)
	if err != nil {
		log.Fatal(err)
	}

	vis, err := NewVisibilityRenderer()
	if err != nil {
		log.Fatal(err)
	}

	return &Game{
		version:         version,
		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		camera:          NewCamera(),
		world:           w,
		fov:             fov,
		memory:          mem,
		visibility:      vis,
		worldColor:      ebiten.NewImage(worldPixW, worldPixH),
		visibilityWorld: ebiten.NewImage(worldPixW, worldPixH),
	}
}

func (g *Game) Update() error {
	p := g.world.Player()

	// Compute screen-space angle from player's eye to mouse cursor.
	eyeWX, eyeWY := p.EyeWorldPos()
	eyeSX, eyeSY := g.camera.WorldToScreen(eyeWX, eyeWY, g.screenWidth, g.screenHeight)
	mx, my := ebiten.CursorPosition()
	mouseAngle := math.Atan2(float64(my)-eyeSY, float64(mx)-eyeSX)
	p.SetMouseAngle(mouseAngle)

	g.world.Update()
	g.camera.Follow(p.X, p.Y)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Compute actual frame delta for accurate memory decay/fade.
	now := time.Now()
	dt := now.Sub(g.lastDraw).Seconds()
	if g.lastDraw.IsZero() || dt > 0.1 {
		dt = 1.0 / 60.0
	}
	g.lastDraw = now

	sw, sh := g.screenWidth, g.screenHeight

	// Recreate the screen-sized proxies if dimensions changed.
	if g.memoryScreen == nil || g.memoryScreen.Bounds().Dx() != sw || g.memoryScreen.Bounds().Dy() != sh {
		g.memoryScreen = ebiten.NewImage(sw, sh)
		g.visibilityScreen = ebiten.NewImage(sw, sh)
	}

	g.world.Draw(screen, g.camera, sw, sh)

	// Refresh the world color snapshot only when tiles actually changed; for a
	// large mostly-static map this avoids redrawing every tile each frame.
	if g.world.ColorDirty() {
		g.world.DrawWorldSpace(g.worldColor)
	}

	p := g.world.Player()
	eyeWX, eyeWY := p.EyeWorldPos()
	eyeSX, eyeSY := g.camera.WorldToScreen(eyeWX, eyeWY, sw, sh)

	// Single ray-march pass: build the world-space line-of-sight mask from the
	// cached collision texture. Memory reads it directly; FOV reads its screen
	// projection. Both apply the cheap directional cone themselves.
	collisionTex := g.world.CollisionImage()
	g.visibility.Draw(g.visibilityWorld, collisionTex, eyeWX, eyeWY, sw, sh)
	projectWorldToScreen(g.visibilityScreen, g.visibilityWorld, g.camera, sw, sh)

	g.memory.Update(g.worldColor, g.visibilityWorld, eyeWX, eyeWY, p.DirAngle(), dt, p.Memory)
	g.memory.DrawToScreen(g.memoryScreen, g.camera, sw, sh)
	g.fov.Draw(screen, g.memoryScreen, g.visibilityScreen, eyeSX, eyeSY, p.DirAngle(), sw, sh)

	// DEBUG (hold B): tint sight blockers (blocksSightAt → CollisionImage) red over
	// the world. Additive blend so the black floor adds nothing and only walls glow.
	if ebiten.IsKeyPressed(ebiten.KeyB) {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(Scale, Scale)
		op.GeoM.Translate(-g.camera.X*Scale+float64(sw)/2, -g.camera.Y*Scale+float64(sh)/2)
		op.Blend = ebiten.BlendLighter
		screen.DrawImage(collisionTex, op)
	}

	//g.drawWorldAxes(screen)

	// Memory forget time as a human-readable string ("infinity" when the player never forgets).
	forget := memoryForgetSeconds(p.Memory)
	forgetStr := "infinity"
	if !math.IsInf(forget, 1) {
		forgetStr = fmt.Sprintf("%.1fs", forget)
	}

	info := fmt.Sprintf(
		"TPS% 3d FPS% 4d\nWorld: player (%.1f, %.1f)\nMouse: angle %.1f°\nMemory [ / ]: %.0f/100  forget: %s",
		int(ebiten.ActualTPS()), int(ebiten.ActualFPS()),
		p.X, p.Y, math.Mod(p.DirAngle()*180/math.Pi+360, 360),
		p.Memory, forgetStr,
	)
	ebitenutil.DebugPrint(screen, info)

	const glyphW = 6
	ebitenutil.DebugPrintAt(screen, g.version, sw-len(g.version)*glyphW-3, 0)
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
