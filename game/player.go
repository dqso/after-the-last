package game

import (
	"math"

	"github.com/dqso/after-the-last/tileset"
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

// EyeOffset is an offset in unscaled sprite pixels from the top-left corner of the sprite tile.
type EyeOffset struct {
	X, Y float64
}

// EyeOffsetsIdle holds eye offsets for the idle animation, indexed by direction.
// Directions: dirRight=0, dirUp=1, dirLeft=2, dirDown=3.
var EyeOffsetsIdle = [4]EyeOffset{
	{12, 20.5}, {8, 21}, {4, 20.5}, {8, 21},
}

// EyeOffsetsRun holds eye offsets for the run animation, indexed by [direction][frame].
// 4 directions × runFrames(6) frames.
var EyeOffsetsRun = [4][runFrames]EyeOffset{
	{{12, 20.5}, {12, 20.5}, {12, 20.5}, {12, 20.5}, {12, 20.5}, {12, 20.5}},
	{{8, 21}, {8, 21}, {8, 21}, {8, 21}, {8, 21}, {8, 21}},
	{{4, 20.5}, {4, 20.5}, {4, 20.5}, {4, 20.5}, {4, 20.5}, {4, 20.5}},
	{{8, 21}, {8, 21}, {8, 21}, {8, 21}, {8, 21}, {8, 21}},
}

type Player struct {
	X, Y        float64 // world position = bottom-center of sprite (pivot)
	CollisionRX float64
	CollisionRY float64
	dir         direction
	moving      bool
	frame       int
	tick        int
	tiles       tilesetListProvider
}

func NewPlayer(tiles tilesetListProvider, collisionRX, collisionRY float64) *Player {
	return &Player{
		tiles:       tiles,
		CollisionRX: collisionRX,
		CollisionRY: collisionRY,
		dir:         dirDown,
	}
}

func (p *Player) SetPosition(x, y float64) {
	p.X = x
	p.Y = y
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

// EyeOffset returns the tile-local offset (from sprite top-left, unscaled pixels) for the current frame.
func (p *Player) EyeOffset() EyeOffset {
	if p.moving {
		return EyeOffsetsRun[p.dir][p.frame]
	}
	return EyeOffsetsIdle[p.dir]
}

// EyeWorldPos converts the current frame's eye offset to world coordinates.
func (p *Player) EyeWorldPos() (float64, float64) {
	off := p.EyeOffset()
	tw := float64(p.tiles.TileW())
	th := float64(p.tiles.TileH())
	// Pivot is bottom-center: sprite top-left = (p.X - tw/2, p.Y - th).
	return p.X - tw/2 + off.X, p.Y - th + off.Y
}

// DirAngle returns the facing direction in radians (right=0, down=π/2, left=π, up=-π/2).
func (p *Player) DirAngle() float64 {
	switch p.dir {
	case dirRight:
		return 0
	case dirDown:
		return math.Pi / 2
	case dirLeft:
		return math.Pi
	case dirUp:
		return -math.Pi / 2
	}
	return 0
}

func (p *Player) Draw(screen *ebiten.Image, cam *Camera, screenW, screenH int) {
	var tileIndex int
	if p.moving {
		tileIndex = int(p.dir)*runFrames + p.frame
		tileIndex += tileset.BobRunSheet << tileset.SheetShift
	} else {
		tileIndex = int(p.dir)
		tileIndex += tileset.BobIdleSheet << tileset.SheetShift
	}

	tile := p.tiles.Tile(tileIndex)
	sx, sy := cam.WorldToScreen(p.X, p.Y, screenW, screenH)

	// Pivot is bottom-center: top-left offset = (-tileW/2, -tileH) in screen pixels.
	tw := float64(p.tiles.TileW())
	th := float64(p.tiles.TileH())
	drawX := sx - tw*Scale/2
	drawY := sy - th*Scale

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(Scale, Scale)
	op.GeoM.Translate(drawX, drawY)
	screen.DrawImage(tile, op)
}
