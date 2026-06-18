package tileset

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// SheetID identifies a tileset sheet in a TileSetList.
// Matches the XXXX part of the 0xXXXXYYYY tile ID format.
type SheetID = int

const (
	RoomBuilderSheet SheetID = 0x0001
	InteriorsSheet   SheetID = 0x0002
	ExtraSheet       SheetID = 0x0003

	BobIdleSheet SheetID = 0x0008
	BobRunSheet  SheetID = 0x0009
)

const (
	sheetMask  = 0xFFFF0000
	tileMask   = 0x0000FFFF
	SheetShift = 16
)

// TileSetList holds multiple named tilesets and resolves tile IDs of the form 0xXXXXYYYY.
// All registered sheets must share the same tile dimensions.
type TileSetList struct {
	sheets map[SheetID]*Tileset
	tileW  int
	tileH  int
}

// NewTileSetList creates an empty TileSetList.
func NewTileSetList() *TileSetList {
	return &TileSetList{sheets: make(map[SheetID]*Tileset)}
}

// Register adds a tileset under the given SheetID.
// All sheets must share the same tile dimensions; the first registered sheet sets the expected size.
func (l *TileSetList) Register(id SheetID, ts *Tileset) error {
	if l.tileW == 0 {
		l.tileW, l.tileH = ts.TileW(), ts.TileH()
	}
	if ts.TileW() != l.tileW {
		return fmt.Errorf("tileset: sheet 0x%04X tile width %d != expected %d", id, ts.TileW(), l.tileW)
	}
	if ts.TileH() != l.tileH {
		return fmt.Errorf("tileset: sheet 0x%04X tile height %d != expected %d", id, ts.TileH(), l.tileH)
	}
	l.sheets[id] = ts
	return nil
}

// Tile decodes a 0xXXXXYYYY tile ID and returns the corresponding sub-image.
func (l *TileSetList) Tile(tileID int) *ebiten.Image {
	sheetID := (tileID & sheetMask) >> SheetShift
	tileIdx := tileID & tileMask

	ts, ok := l.sheets[sheetID]
	if !ok {
		// Solid magenta tile signals a missing sheet during development.
		img := ebiten.NewImage(l.tileW, l.tileH)
		img.Fill(color.RGBA{R: 0xFF, G: 0x00, B: 0xFF, A: 0xFF})
		return img
	}
	return ts.Tile(tileIdx)
}

func (l *TileSetList) TileW() int { return l.tileW }
func (l *TileSetList) TileH() int { return l.tileH }
