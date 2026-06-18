package tileset

import (
	"bytes"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

// Tileset slices a spritesheet into tiles of fixed size.
// Tiles are indexed in row-major order starting from 0.
type Tileset struct {
	img   *ebiten.Image
	tileW int
	tileH int
	cols  int
}

func NewTileSet(data []byte, tileW, tileH int) (*Tileset, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	eimg := ebiten.NewImageFromImage(img)
	return &Tileset{
		img:   eimg,
		tileW: tileW,
		tileH: tileH,
		cols:  eimg.Bounds().Dx() / tileW,
	}, nil
}

// Tile returns the sub-image for the given tile index (row-major, 0-based).
func (t *Tileset) Tile(index int) *ebiten.Image {
	col := index % t.cols
	row := index / t.cols
	x := col * t.tileW
	y := row * t.tileH
	rect := image.Rect(x, y, x+t.tileW, y+t.tileH)
	return t.img.SubImage(rect).(*ebiten.Image)
}

func (t *Tileset) TileW() int { return t.tileW }
func (t *Tileset) TileH() int { return t.tileH }
