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

	// faceSpeed is the maximum head rotation in radians per tick (~300°/s at 60 TPS).
	faceSpeed = 5 * math.Pi / 180
)

type direction int

const (
	dirRight direction = iota // 0 — matches sprite column order: right, up, left, down
	dirUp                     // 1
	dirLeft                   // 2
	dirDown                   // 3
)

// eyeOffsetX, eyeOffsetY — fixed eye position in unscaled sprite pixels from the sprite's top-left corner.
const eyeOffsetX, eyeOffsetY = 8.0, 21.0

type Player struct {
	X, Y        float64 // world position = bottom-center of sprite (pivot)
	CollisionRX float64
	CollisionRY float64
	dir         direction
	mouseAngle  float64 // raw screen-space angle from eye to cursor (right=0, down=π/2)
	faceAngle   float64 // current smoothed facing angle, chases mouseAngle at faceSpeed
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

// SetMouseAngle stores the screen-space angle from the player's eye to the cursor.
// Must be called before Update() each frame.
func (p *Player) SetMouseAngle(angle float64) {
	p.mouseAngle = angle
}

// dirFromAngle snaps a screen-space angle to one of 4 sprite directions.
func dirFromAngle(angle float64) direction {
	switch {
	case angle >= -math.Pi/4 && angle < math.Pi/4:
		return dirRight
	case angle >= math.Pi/4 && angle < 3*math.Pi/4:
		return dirDown
	case angle >= -3*math.Pi/4 && angle < -math.Pi/4:
		return dirUp
	default:
		return dirLeft
	}
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

	// Rotate faceAngle toward mouseAngle along the shortest arc.
	diff := math.Atan2(math.Sin(p.mouseAngle-p.faceAngle), math.Cos(p.mouseAngle-p.faceAngle))
	if math.Abs(diff) <= faceSpeed {
		p.faceAngle = p.mouseAngle
	} else {
		p.faceAngle += math.Copysign(faceSpeed, diff)
		p.faceAngle = math.Atan2(math.Sin(p.faceAngle), math.Cos(p.faceAngle))
	}

	// Sprite direction always follows mouse, even during movement (can run backwards).
	p.dir = dirFromAngle(p.faceAngle)

	if p.moving {
		if dx != 0 && dy != 0 {
			dx *= math.Sqrt2 / 2
			dy *= math.Sqrt2 / 2
		}
		p.X += dx
		p.Y += dy

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

func (p *Player) Moving() bool { return p.moving }

func (p *Player) DirName() string {
	switch p.dir {
	case dirRight:
		return "right"
	case dirLeft:
		return "left"
	case dirUp:
		return "up"
	case dirDown:
		return "down"
	}
	return "?"
}

// EyeWorldPos returns the eye position in world coordinates.
func (p *Player) EyeWorldPos() (float64, float64) {
	tw := float64(p.tiles.TileW())
	th := float64(p.tiles.TileH())
	// Pivot is bottom-center: sprite top-left = (p.X - tw/2, p.Y - th).
	return p.X - tw/2 + eyeOffsetX, p.Y - th + eyeOffsetY
}

// DirAngle returns the current smoothed facing angle in radians for use by the FOV shader.
func (p *Player) DirAngle() float64 {
	return p.faceAngle
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
