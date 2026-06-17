package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	playerSpeed       = 1.0
	animTicksPerFrame = 8
	runFrames         = 6
)

type direction int

const (
	dirRight direction = iota // 0 — matches sprite column order: right, up, left, down
	dirUp                     // 1
	dirLeft                   // 2
	dirDown                   // 3
)

type Player struct {
	X, Y   float64 // world position = bottom-center of sprite (pivot)
	dir    direction
	moving bool
	frame  int
	tick   int
	idle   tileProvider // 4 frames: one per direction, each 16x32
	run    tileProvider // 24 frames: 6 per direction, each 16x32
}

func NewPlayer(idle, run tileProvider, x, y float64) *Player {
	return &Player{idle: idle, run: run, X: x, Y: y, dir: dirDown}
}

func (p *Player) Update() {
	dx, dy := 0.0, 0.0

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		dy = -playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		dy = playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx = -playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		dx = playerSpeed
	}

	p.moving = dx != 0 || dy != 0

	if p.moving {
		if dx != 0 && dy != 0 {
			dx *= math.Sqrt2 / 2
			dy *= math.Sqrt2 / 2
		}
		p.X += dx
		p.Y += dy

		// Horizontal takes priority for direction.
		if dx > 0 {
			p.dir = dirRight
		} else if dx < 0 {
			p.dir = dirLeft
		} else if dy < 0 {
			p.dir = dirUp
		} else {
			p.dir = dirDown
		}

		p.tick++
		if p.tick >= animTicksPerFrame {
			p.tick = 0
			p.frame = (p.frame + 1) % runFrames
		}
	} else {
		p.tick = 0
		p.frame = 0
	}
}

func (p *Player) Draw(screen *ebiten.Image, cam *Camera, screenW, screenH int) {
	var ts tileProvider
	var tileIndex int
	if p.moving {
		ts = p.run
		tileIndex = int(p.dir)*runFrames + p.frame
	} else {
		ts = p.idle
		tileIndex = int(p.dir)
	}

	tile := ts.Tile(tileIndex)
	sx, sy := cam.WorldToScreen(p.X, p.Y, screenW, screenH)

	// Pivot is bottom-center: top-left offset = (-tileW/2, -tileH) in screen pixels.
	tw := float64(ts.TileW())
	th := float64(ts.TileH())
	drawX := sx - tw*Scale/2
	drawY := sy - th*Scale

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(Scale, Scale)
	op.GeoM.Translate(drawX, drawY)
	screen.DrawImage(tile, op)
}
