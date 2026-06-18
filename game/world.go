package game

import (
	"math/rand"

	"github.com/dqso/after-the-last/collision"
	"github.com/hajimehoshi/ebiten/v2"
)

type tilesetListProvider interface {
	Tile(tileID int) *ebiten.Image
	TileW() int
	TileH() int
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
}

func NewWorld(ts tilesetListProvider, floor, walls [][]int, items [][]int, collision [][]collision.Type, player *Player) *World {
	rows := len(floor)
	cols := 0
	if rows > 0 {
		cols = len(floor[0])
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
	}
}

func (w *World) Player() *Player { return w.player }

func (w *World) Update() {
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
	for row := range w.floor {
		w.drawRow(screen, w.floor, row, cam, screenW, screenH)
	}

	th := float64(w.tiles.TileH())
	playerDrawn := false

	for row := range w.walls {
		tilePivotY := float64(row)*th + th/2

		if !playerDrawn && w.player.Y >= tilePivotY-th && w.player.Y < tilePivotY {
			w.player.Draw(screen, cam, screenW, screenH)
			playerDrawn = true
		}

		w.drawRow(screen, w.walls, row, cam, screenW, screenH)
		w.drawRow(screen, w.items, row, cam, screenW, screenH)
	}

	if !playerDrawn {
		w.player.Draw(screen, cam, screenW, screenH)
	}
}

func (w *World) drawRow(screen *ebiten.Image, layer [][]int, row int, cam *Camera, screenW, screenH int) {
	if row >= len(layer) {
		return
	}
	tw := float64(w.tiles.TileW())
	th := float64(w.tiles.TileH())
	hw, hh := float64(screenW)/2, float64(screenH)/2

	for col, tileIdx := range layer[row] {
		if tileIdx == -1 {
			continue
		}
		sx := (float64(col)*tw-cam.X)*Scale + hw
		sy := (float64(row)*th-cam.Y)*Scale + hh

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(Scale, Scale)
		op.GeoM.Translate(sx, sy)
		screen.DrawImage(w.tiles.Tile(tileIdx), op)
	}
}
