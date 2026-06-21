package game

import (
	"math/rand"

	"github.com/dqso/after-the-last/collision"
	"github.com/dqso/after-the-last/random"
	"github.com/hajimehoshi/ebiten/v2"
)

type tilesetListProvider interface {
	Tile(tileID int) *ebiten.Image
	TileW() int
	TileH() int
}

// buttonToggleTicks is how many Update ticks between button state changes (60 TPS = 1 s).
const buttonToggleTicks = 60

// button tracks an animated wall button at a fixed grid position.
type button struct {
	row, col int
	tick     int
	active   bool
}

// World holds the tile map, entities, and renders them.
type World struct {
	tiles     tilesetListProvider
	floor     [][]int            // floor[row][col] = tile index, -1 = empty
	walls     [][]int            // walls[row][col] = tile index, -1 = empty
	items     [][]int            // items[row][col] = tile index, -1 = empty
	collision [][]collision.Type // collision[row][col]
	cols      int
	rows      int
	player    *Player
	buttons   []button

	// colorDirty marks that the world color snapshot must be re-rendered for
	// the memory layer (set on tile changes). Avoids redrawing a large, mostly
	// static map every frame.
	colorDirty bool

	// collisionTex is a cached world-pixel texture of sight blockers, built once
	// from the (static) collision data. See CollisionImage.
	collisionTex *ebiten.Image
}

func NewWorld(ts tilesetListProvider, floor, walls [][]int, items [][]int, collision [][]collision.Type, player *Player) *World {
	rows := len(floor)
	cols := 0
	if rows > 0 {
		cols = len(floor[0])
	}

	// Collect buttons from the items grid.
	var buttons []button
	for r, row := range items {
		for c, id := range row {
			if id == random.ButtonInactiveTileID || id == random.ButtonActiveTileID {
				buttons = append(buttons, button{row: r, col: c})
			}
		}
	}

	return &World{
		tiles:     ts,
		floor:     floor,
		walls:     walls,
		items:     items,
		collision: collision,
		cols:      cols,
		rows:      rows,
		player:    player,
		buttons:   buttons,

		colorDirty: true, // force initial snapshot render
	}
}

func (w *World) Player() *Player { return w.player }

// WorldPixelSize returns the world dimensions in unscaled tile pixels.
func (w *World) WorldPixelSize() (int, int) {
	return w.cols * w.tiles.TileW(), w.rows * w.tiles.TileH()
}

// CollisionImage returns a world-pixel-sized texture holding two sight layers,
// sampled by the visibility pass:
//
//	R = collision layer  — 1 where the pixel is impassable by collision.
//	G = wall layer       — 1 where the pixel's tile holds a wall tile.
//
// The ray stops at the collision layer, but a wall pixel stays visible (its face
// is drawn): the visibility pass tests the collision layer only up to just before
// a wall pixel, so the player sees wall surfaces without peeking through them into
// the room beyond. The resolution is 1:1 with world pixels so obstacles of any
// shape occlude exactly per pixel.
//
// Built once and cached: the layers are static after world generation. If dynamic
// blockers are added later, invalidate w.collisionTex when they move.
func (w *World) CollisionImage() *ebiten.Image {
	if w.collisionTex != nil {
		return w.collisionTex
	}
	tw, th := w.tiles.TileW(), w.tiles.TileH()
	pw, ph := w.WorldPixelSize()
	img := ebiten.NewImage(pw, ph)
	buf := make([]byte, pw*ph*4)
	for py := 0; py < ph; py++ {
		row := py / th
		for px := 0; px < pw; px++ {
			i := (py*pw + px) * 4
			buf[i+3] = 0xff // opaque, so the channels survive alpha premultiplication
			if w.CollidesAt(float64(px)+0.5, float64(py)+0.5) {
				buf[i] = 0xff // R: collision layer
			}
			col := px / tw
			if row >= 0 && row < len(w.walls) && col >= 0 && col < len(w.walls[row]) &&
				w.walls[row][col] != emptyTile {
				buf[i+1] = 0xff // G: wall layer
			}
		}
	}
	img.WritePixels(buf)
	w.collisionTex = img
	return w.collisionTex
}

// emptyTile marks an empty cell in the floor/walls/items grids.
const emptyTile = -1

// updateButtons advances each button's tick and toggles its tile ID every second.
func (w *World) updateButtons() {
	for i := range w.buttons {
		b := &w.buttons[i]
		b.tick++
		if b.tick < buttonToggleTicks {
			continue
		}
		b.tick = 0
		b.active = !b.active
		if b.active {
			w.items[b.row][b.col] = random.ButtonActiveTileID
		} else {
			w.items[b.row][b.col] = random.ButtonInactiveTileID
		}
		w.colorDirty = true // tile changed: snapshot needs a refresh
	}
}

func (w *World) Update() {
	w.updateButtons()

	oldX, oldY := w.player.X, w.player.Y
	w.player.Update()

	rx, ry := w.player.CollisionRX, w.player.CollisionRY
	if w.EllipseCollidesAt(w.player.X, w.player.Y, rx, ry) {
		// Try sliding along X axis.
		switch {
		case !w.EllipseCollidesAt(w.player.X, oldY, rx, ry):
			w.player.Y = oldY
		case !w.EllipseCollidesAt(oldX, w.player.Y, rx, ry):
			// Try sliding along Y axis.
			w.player.X = oldX
		default:
			w.player.X, w.player.Y = oldX, oldY
		}
	}
}

// CollidesAt reports whether point (x, y) is blocked.
func (w *World) CollidesAt(x, y float64) bool {
	tw := float64(w.tiles.TileW())
	th := float64(w.tiles.TileH())
	col, row := int(x/tw), int(y/th)
	ct := w.collisionAt(col, row)
	if ct == collision.Free {
		return false
	}
	if ct == collision.Blocked {
		return true
	}
	tileL, tileT := float64(col)*tw, float64(row)*th
	fx0, fy0, fx1, fy1 := blockedSubrect(ct)
	return x >= tileL+fx0*tw && x < tileL+fx1*tw &&
		y >= tileT+fy0*th && y < tileT+fy1*th
}

// EllipseCollidesAt reports whether an ellipse at (cx, cy) with radii (rx, ry)
// overlaps any blocked region.
func (w *World) EllipseCollidesAt(cx, cy, rx, ry float64) bool {
	tw := float64(w.tiles.TileW())
	th := float64(w.tiles.TileH())

	colMin := int((cx - rx) / tw)
	colMax := int((cx + rx) / tw)
	rowMin := int((cy - ry) / th)
	rowMax := int((cy + ry) / th)

	for row := rowMin; row <= rowMax; row++ {
		for col := colMin; col <= colMax; col++ {
			ct := w.collisionAt(col, row)
			if ct != collision.Free && w.ellipseOverlapsSubrect(cx, cy, rx, ry, col, row, ct, tw, th) {
				return true
			}
		}
	}
	return false
}

func (w *World) collisionAt(col, row int) collision.Type {
	if row < 0 || row >= len(w.collision) || col < 0 || col >= len(w.collision[row]) {
		return collision.Blocked
	}
	return w.collision[row][col]
}

// ellipseOverlapsSubrect checks if ellipse at (cx,cy) with radii (rx,ry) overlaps
// the blocked sub-rect of the tile defined by ct.
func (w *World) ellipseOverlapsSubrect(cx, cy, rx, ry float64, col, row int, ct collision.Type, tw, th float64) bool {
	tileL, tileT := float64(col)*tw, float64(row)*th
	fx0, fy0, fx1, fy1 := blockedSubrect(ct)
	l := tileL + fx0*tw
	t := tileT + fy0*th
	r := tileL + fx1*tw
	b := tileT + fy1*th
	nx := max(l, min(cx, r))
	ny := max(t, min(cy, b))
	dx := (nx - cx) / rx
	dy := (ny - cy) / ry
	return dx*dx+dy*dy <= 1
}

// FindFreeCell returns (col, row, true) for a cell with collision.Free.
// If rng is nil, the first such cell (top-left scan) is returned; otherwise
// a random one is picked. Returns (0, 0, false) if none exists.
func (w *World) FindFreeCell(rng *rand.Rand) (col, row int, ok bool) {
	var candidates [][2]int
	for r := 0; r < w.rows; r++ {
		for c := 0; c < w.cols; c++ {
			if w.collisionAt(c, r) != collision.Free {
				continue
			}
			if rng == nil {
				return c, r, true
			}
			candidates = append(candidates, [2]int{c, r})
		}
	}
	if len(candidates) == 0 {
		return 0, 0, false
	}
	pick := candidates[rng.Intn(len(candidates))]
	return pick[0], pick[1], true
}

// CellCenter returns the world-space center of cell (col, row).
func (w *World) CellCenter(col, row int) (float64, float64) {
	tw := float64(w.tiles.TileW())
	th := float64(w.tiles.TileH())
	return float64(col)*tw + tw/2, float64(row)*th + th/2
}

// Draw renders rows top-to-bottom. The player is inserted when their pivot Y
// falls within the window [tilePivotY-th, tilePivotY) of the current row.
func (w *World) Draw(screen *ebiten.Image, cam *Camera, screenW, screenH int) {
	view := screenView{cam: cam, screenW: screenW, screenH: screenH}

	for row := range w.floor {
		w.drawRow(screen, w.floor, row, view)
	}

	th := float64(w.tiles.TileH())
	playerDrawn := false

	for row := range w.walls {
		tilePivotY := float64(row)*th + th/2

		if !playerDrawn && w.player.Y >= tilePivotY-th && w.player.Y < tilePivotY {
			w.player.Draw(screen, cam, screenW, screenH)
			playerDrawn = true
		}

		w.drawRow(screen, w.walls, row, view)
		w.drawRow(screen, w.items, row, view)
	}

	if !playerDrawn {
		w.player.Draw(screen, cam, screenW, screenH)
	}
}

// DrawWorldSpace renders all tile layers into dst row by row (top-to-bottom),
// at world-pixel scale (1:1) without camera and without the player. It reuses
// the same row-drawing flow as Draw to capture a flat color snapshot of the
// world for the memory layer. dst must be world-pixel sized. It clears the
// dirty flag so the (large, mostly static) map is re-rendered only on changes.
func (w *World) DrawWorldSpace(dst *ebiten.Image) {
	dst.Clear()
	view := worldView{}

	for row := range w.floor {
		w.drawRow(dst, w.floor, row, view)
	}
	for row := range w.walls {
		w.drawRow(dst, w.walls, row, view)
		w.drawRow(dst, w.items, row, view)
	}

	w.colorDirty = false
}

// ColorDirty reports whether the world color snapshot needs re-rendering.
func (w *World) ColorDirty() bool { return w.colorDirty }

type ebitenImage interface {
	DrawImage(img *ebiten.Image, options *ebiten.DrawImageOptions)
}

// tileView computes the draw transform for a tile cell, letting the same row
// flow target either the screen (camera + Scale) or a flat world-space buffer.
type tileView interface {
	tileGeoM(col, row int, tw, th float64) ebiten.GeoM
}

// screenView projects world tiles to the screen via the camera and Scale.
type screenView struct {
	cam              *Camera
	screenW, screenH int
}

func (v screenView) tileGeoM(col, row int, tw, th float64) ebiten.GeoM {
	var g ebiten.GeoM
	g.Scale(Scale, Scale)
	g.Translate(
		(float64(col)*tw-v.cam.X)*Scale+float64(v.screenW)/2,
		(float64(row)*th-v.cam.Y)*Scale+float64(v.screenH)/2,
	)
	return g
}

// worldView draws tiles 1:1 into a world-pixel-sized buffer (no camera).
type worldView struct{}

func (worldView) tileGeoM(col, row int, tw, th float64) ebiten.GeoM {
	var g ebiten.GeoM
	g.Translate(float64(col)*tw, float64(row)*th)
	return g
}

func (w *World) drawRow(dst ebitenImage, layer [][]int, row int, view tileView) {
	if row >= len(layer) {
		return
	}
	tw := float64(w.tiles.TileW())
	th := float64(w.tiles.TileH())

	for col, tileIdx := range layer[row] {
		if tileIdx == -1 {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM = view.tileGeoM(col, row, tw, th)
		dst.DrawImage(w.tiles.Tile(tileIdx), op)
	}
}
