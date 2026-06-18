package random

import (
	"encoding/xml"
	"math/rand"
	"strconv"

	"github.com/dqso/after-the-last/collision"
)

// tileID encodes both sheet and tile index in a single value: 0xXXXXYYYY.
// XXXX - sheet ID, YYYY - tile index within the sheet.
type tileID = int

func newTileID(sheet, idx int) tileID {
	return sheet<<16 | idx
}

// floorType holds tile IDs for each zone of a floor pattern.
type floorType struct {
	name    string
	leftTop tileID // left|top corner
	top     tileID // top edge
	left    tileID // left edge
	clear   tileID // interior fill
}

// wallCommonType holds structural tile IDs shared by all wall styles.
type wallCommonType struct {
	cornerLeftTop  tileID
	cornerRightTop tileID
	left           tileID
	right          tileID
	cornerLeftBot  tileID
	bottom         tileID
	cornerRightBot tileID
}

// wallVariantType holds fill tile IDs for a named wall style.
type wallVariantType struct {
	name string
	top  tileID // row 0 fill: top decoration stripe
	face tileID // row 1 fill: inner-top face (wall depth visible from above)
}

type Generator struct {
	floors     []floorType
	wallCommon wallCommonType
	walls      []wallVariantType
}

func NewGenerator(tileInfoXML []byte) (Generator, error) {
	var raw struct {
		Floors struct {
			Tileset string `xml:"tileset,attr"`
			List    []struct {
				Name  string `xml:"name,attr"`
				Tiles []struct {
					ID   int    `xml:"id,attr"`
					Type string `xml:"type,attr"`
				} `xml:"tile"`
			} `xml:"floor"`
		} `xml:"floors"`
		Walls struct {
			Tileset string `xml:"tileset,attr"`
			List    []struct {
				Name  string `xml:"name,attr"`
				Tiles []struct {
					ID   int    `xml:"id,attr"`
					Type string `xml:"type,attr"`
				} `xml:"tile"`
			} `xml:"wall"`
		} `xml:"walls"`
	}

	if err := xml.Unmarshal(tileInfoXML, &raw); err != nil {
		return Generator{}, err
	}

	floorsSheet, err := strconv.ParseInt(raw.Floors.Tileset, 0, 64)
	if err != nil {
		return Generator{}, err
	}
	wallsSheet, err := strconv.ParseInt(raw.Walls.Tileset, 0, 64)
	if err != nil {
		return Generator{}, err
	}
	fs, ws := int(floorsSheet), int(wallsSheet)

	floors := make([]floorType, 0, len(raw.Floors.List))
	for _, f := range raw.Floors.List {
		ft := floorType{name: f.Name}
		for _, t := range f.Tiles {
			switch t.Type {
			case "left|top":
				ft.leftTop = newTileID(fs, t.ID)
			case "top":
				ft.top = newTileID(fs, t.ID)
			case "left":
				ft.left = newTileID(fs, t.ID)
			case "clear":
				ft.clear = newTileID(fs, t.ID)
			}
		}
		floors = append(floors, ft)
	}

	var wc wallCommonType
	walls := make([]wallVariantType, 0)
	for _, w := range raw.Walls.List {
		if w.Name == "common" {
			for _, t := range w.Tiles {
				switch t.Type {
				case "corner_left_top":
					wc.cornerLeftTop = newTileID(ws, t.ID)
				case "corner_right_top":
					wc.cornerRightTop = newTileID(ws, t.ID)
				case "left":
					wc.left = newTileID(ws, t.ID)
				case "right":
					wc.right = newTileID(ws, t.ID)
				case "corner_left_bottom":
					wc.cornerLeftBot = newTileID(ws, t.ID)
				case "bottom":
					wc.bottom = newTileID(ws, t.ID)
				case "corner_right_bottom":
					wc.cornerRightBot = newTileID(ws, t.ID)
				}
			}
			continue
		}
		wv := wallVariantType{name: w.Name}
		for _, t := range w.Tiles {
			switch t.Type {
			case "top":
				wv.top = newTileID(ws, t.ID)
			case "bottom":
				wv.face = newTileID(ws, t.ID)
			}
		}
		walls = append(walls, wv)
	}

	return Generator{floors: floors, wallCommon: wc, walls: walls}, nil
}

const none = -1

const (
	// ButtonInactiveTileID is ExtraSheet tile 4 — button off state.
	ButtonInactiveTileID = 0x00030004
	// ButtonActiveTileID is ExtraSheet tile 2 — button on state.
	ButtonActiveTileID = 0x00030002
)

func (g Generator) GenerateV0() ([][]int, [][]int, [][]int, [][]int) {
	var floorGrid = [][]int{
		{none, none, none, none, none, none, none, none, none},
		{none, none, none, none, none, none, none, none, none},
		{none, 0x100A4, 0x100A5, 0x100A5, 0x100A5, 0x100A5, 0x100A5, 0x100A5, none},
		{none, 0x100B5, 0x100B6, 0x100B6, 0x100B6, 0x100B6, 0x100B6, 0x100B6, none},
		{none, 0x100B5, 0x100B6, 0x100B6, 0x100B6, 0x100B6, 0x100B6, 0x100B6, none},
		{none, 0x100B5, 0x100B6, 0x100B6, 0x100B6, 0x100B6, 0x100B6, 0x100B6, none},
		{none, 0x100B5, 0x100B6, 0x100B6, 0x100B6, 0x100B6, 0x100B6, 0x100B6, none},
		{none, none, none, none, none, none, none, none, none},
	}

	var wallsGrid = [][]int{
		{0x1001C, 0x10144, 0x10144, 0x10144, 0x10144, 0x10144, 0x10144, 0x10144, 0x1001E},
		{0x10007, 0x1015C, 0x1015C, 0x1015C, 0x1015C, 0x1015C, 0x1015C, 0x1015C, 0x10008},
		{0x10007, none, none, none, none, none, none, none, 0x10008},
		{0x10007, none, none, none, none, none, none, none, 0x10008},
		{0x10007, none, none, none, none, none, none, none, 0x10008},
		{0x10007, none, none, none, none, none, none, none, 0x10008},
		{0x10007, none, none, none, none, none, none, none, 0x10008},
		{0x1003E, 0x1003F, 0x1003F, 0x1003F, 0x1003F, 0x1003F, 0x1003F, 0x1003F, 0x10040},
	}

	var itemsGrid = [][]int{
		{none, none, none, none, none, none, none, none, none},
		{none, none, 0x30002, none, none, none, none, none, none},
		{none, none, none, none, none, none, none, none, none},
		{none, none, 0x20262, 0x20263, none, none, none, none, none},
		{none, none, 0x20272, 0x20273, none, none, none, none, none},
		{none, none, none, none, none, none, none, none, none},
		{none, none, none, none, none, none, none, none, none},
		{none, none, none, none, none, none, none, none, none},
	}

	// CollFree=0, CollBlocked=1, see collision/collision.go for partial types.
	var collisionGrid = [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 5, 4, 0, 0, 0, 0, 1},
		{1, 0, 3, 2, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 7, 7, 7, 7, 7, 7, 7, 1},
	}

	return floorGrid, wallsGrid, itemsGrid, collisionGrid
}

func (g Generator) GenerateV1(seed int64) ([][]int, [][]int, [][]int, [][]int) {
	const (
		minWidth  = 10
		maxWidth  = 20
		minHeight = 5
		maxHeight = 15
	)

	rnd := rand.New(rand.NewSource(seed))
	width, height := rnd.Intn(maxWidth-minWidth)+minWidth, rnd.Intn(maxHeight-minHeight)+minHeight

	floorGrid := make([][]int, height)
	for row := range floorGrid {
		floorGrid[row] = make([]int, width)
		for col := range floorGrid[row] {
			floorGrid[row][col] = none
		}
	}

	wallsGrid := make([][]int, height)
	for row := range wallsGrid {
		wallsGrid[row] = make([]int, width)
		for col := range wallsGrid[row] {
			wallsGrid[row][col] = none
		}
	}

	itemsGrid := make([][]int, height)
	for row := range itemsGrid {
		itemsGrid[row] = make([]int, width)
		for col := range itemsGrid[row] {
			itemsGrid[row][col] = none
		}
	}

	collisionGrid := make([][]int, height)
	for row := range collisionGrid {
		collisionGrid[row] = make([]int, width)
	}

	// Pick a random floor pattern and fill the interior (rows 2..h-2, cols 1..w-2).
	floor := g.floors[rnd.Intn(len(g.floors))]
	for row := 2; row < height-1; row++ {
		for col := 1; col < width-1; col++ {
			var id tileID
			switch {
			case row == 2 && col == 1:
				id = floor.leftTop
			case row == 2:
				id = floor.top
			case col == 1:
				id = floor.left
			default:
				id = floor.clear
			}
			floorGrid[row][col] = id
		}
	}

	// Walls: pick a random style, fill border tiles and collision.
	wall := g.walls[rnd.Intn(len(g.walls))]
	wc := g.wallCommon

	// Row 0: top decoration stripe.
	wallsGrid[0][0] = wc.cornerLeftTop
	for col := 1; col < width-1; col++ {
		wallsGrid[0][col] = wall.top
	}
	wallsGrid[0][width-1] = wc.cornerRightTop

	// Row 1: inner-top face (wall depth seen from above).
	wallsGrid[1][0] = wc.left
	for col := 1; col < width-1; col++ {
		wallsGrid[1][col] = wall.face
	}
	wallsGrid[1][width-1] = wc.right

	// Rows 2..h-2: left and right sides only.
	for row := 2; row < height-1; row++ {
		wallsGrid[row][0] = wc.left
		wallsGrid[row][width-1] = wc.right
	}

	// Row h-1: bottom border.
	wallsGrid[height-1][0] = wc.cornerLeftBot
	for col := 1; col < width-1; col++ {
		wallsGrid[height-1][col] = wc.bottom
	}
	wallsGrid[height-1][width-1] = wc.cornerRightBot

	// Collision: rows 0-1 fully blocked, sides blocked, bottom half-blocked.
	for col := 0; col < width; col++ {
		collisionGrid[0][col] = collision.Blocked
		collisionGrid[1][col] = collision.Blocked
	}
	for row := 2; row < height-1; row++ {
		collisionGrid[row][0] = collision.Blocked
		collisionGrid[row][width-1] = collision.Blocked
	}
	collisionGrid[height-1][0] = collision.Blocked
	for col := 1; col < width-1; col++ {
		collisionGrid[height-1][col] = collision.BlockedBot
	}
	collisionGrid[height-1][width-1] = collision.Blocked

	// Place one button on a random tile of the bottom inner wall face.
	buttonCol := 1 + rnd.Intn(width-2)
	itemsGrid[1][buttonCol] = ButtonInactiveTileID

	return floorGrid, wallsGrid, itemsGrid, collisionGrid
}
