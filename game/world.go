package game

import "github.com/hajimehoshi/ebiten/v2"

type tileProvider interface {
	Tile(index int) *ebiten.Image
	TileW() int
	TileH() int
}

// World holds the tile map and renders it.
type World struct {
	tiles     tileProvider
	floor     [][]int // floor[row][col] = tile index, -1 = empty
	walls     [][]int // walls[row][col] = tile index, -1 = empty
	itemTiles tileProvider
	items     [][]int           // items[row][col] = tile index, -1 = empty
	collision [][]CollisionType // collision[row][col]
	cols      int
	rows      int
}

func NewWorld(ts tileProvider, floor, walls [][]int, itemTiles tileProvider, items [][]int, collision [][]CollisionType) *World {
	rows := len(floor)
	cols := 0
	if rows > 0 {
		cols = len(floor[0])
	}
	return &World{
		tiles:     ts,
		floor:     floor,
		walls:     walls,
		itemTiles: itemTiles,
		items:     items,
		collision: collision,
		cols:      cols,
		rows:      rows,
	}
}

// CollidesAt reports whether point (x, y) is blocked.
func (w *World) CollidesAt(x, y float64) bool {
	tw := float64(w.tiles.TileW())
	th := float64(w.tiles.TileH())
	col, row := int(x/tw), int(y/th)
	ct := w.collisionAt(col, row)
	if ct == CollFree {
		return false
	}
	if ct == CollBlocked {
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
			if ct != CollFree && w.ellipseOverlapsSubrect(cx, cy, rx, ry, col, row, ct, tw, th) {
				return true
			}
		}
	}
	return false
}

func (w *World) collisionAt(col, row int) CollisionType {
	if row < 0 || row >= len(w.collision) || col < 0 || col >= len(w.collision[row]) {
		return CollBlocked
	}
	return w.collision[row][col]
}

// ellipseOverlapsSubrect checks if ellipse at (cx,cy) with radii (rx,ry) overlaps
// the blocked sub-rect of the tile defined by ct.
func (w *World) ellipseOverlapsSubrect(cx, cy, rx, ry float64, col, row int, ct CollisionType, tw, th float64) bool {
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

// CellCenter returns the world-space center of cell (col, row).
func (w *World) CellCenter(col, row int) (float64, float64) {
	tw := float64(w.tiles.TileW())
	th := float64(w.tiles.TileH())
	return float64(col)*tw + tw/2, float64(row)*th + th/2
}

// Draw renders layers in order: floor, walls, items.
func (w *World) Draw(screen *ebiten.Image, cameraX, cameraY float64, screenW, screenH int) {
	w.drawLayer(screen, w.tiles, w.floor, cameraX, cameraY, screenW, screenH)
	w.drawLayer(screen, w.tiles, w.walls, cameraX, cameraY, screenW, screenH)
	w.drawLayer(screen, w.itemTiles, w.items, cameraX, cameraY, screenW, screenH)
}

func (w *World) drawLayer(screen *ebiten.Image, tiles tileProvider, layer [][]int, cameraX, cameraY float64, screenW, screenH int) {
	tw := float64(tiles.TileW())
	th := float64(tiles.TileH())
	hw, hh := float64(screenW)/2, float64(screenH)/2

	for row, rowTiles := range layer {
		for col, tileIdx := range rowTiles {
			if tileIdx == -1 {
				continue
			}
			sx := (float64(col)*tw-cameraX)*Scale + hw
			sy := (float64(row)*th-cameraY)*Scale + hh

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(Scale, Scale)
			op.GeoM.Translate(sx, sy)
			screen.DrawImage(tiles.Tile(tileIdx), op)
		}
	}
}
