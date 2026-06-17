package game

import "github.com/hajimehoshi/ebiten/v2"

type tileProvider interface {
	Tile(index int) *ebiten.Image
	TileW() int
	TileH() int
}

// World holds the tile map and renders it.
type World struct {
	tiles tileProvider
	grid  [][]int // grid[row][col] = tile index
	cols  int
	rows  int
}

func NewWorld(ts tileProvider, grid [][]int) *World {
	rows := len(grid)
	cols := 0
	if rows > 0 {
		cols = len(grid[0])
	}
	return &World{tiles: ts, grid: grid, cols: cols, rows: rows}
}

// CellCenter returns the world-space center of cell (col, row).
func (w *World) CellCenter(col, row int) (float64, float64) {
	tw := float64(w.tiles.TileW())
	th := float64(w.tiles.TileH())
	return float64(col)*tw + tw/2, float64(row)*th + th/2
}

// Draw renders the world onto screen, offset by cameraX/cameraY (world-space screen center).
func (w *World) Draw(screen *ebiten.Image, cameraX, cameraY float64, screenW, screenH int) {
	tw := float64(w.tiles.TileW())
	th := float64(w.tiles.TileH())
	hw, hh := float64(screenW)/2, float64(screenH)/2

	for row, rowTiles := range w.grid {
		for col, tileIdx := range rowTiles {
			wx := float64(col) * tw
			wy := float64(row) * th
			sx := (wx-cameraX)*Scale + hw
			sy := (wy-cameraY)*Scale + hh

			tile := w.tiles.Tile(tileIdx)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(Scale, Scale)
			op.GeoM.Translate(sx, sy)
			screen.DrawImage(tile, op)
		}
	}
}
