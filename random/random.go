package random

import (
	"encoding/xml"
	"strconv"
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

	// Big rounded corners that join a room wall to a tunnel at its mouth.
	bigCornerLeftTop   tileID // left mouth (room opens to the right)
	bigCornerLeftRight tileID // right mouth (room opens to the left)
}

// wallVariantType holds fill tile IDs for a named wall style.
type wallVariantType struct {
	name      string
	topBorder tileID // row 0 fill: top decoration stripe
	face      tileID // row 1 fill: inner-top face (wall depth visible from above)

	// Variant-specific corner and edge tiles (optional decoration).
	leftTop     tileID
	top         tileID
	rightTop    tileID
	leftBottom  tileID
	rightBottom tileID
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
				case "big_corner_left_top":
					wc.bigCornerLeftTop = newTileID(ws, t.ID)
				case "big_corner_left_right":
					wc.bigCornerLeftRight = newTileID(ws, t.ID)
				}
			}
			continue
		}
		wv := wallVariantType{name: w.Name}
		for _, t := range w.Tiles {
			switch t.Type {
			case "top_border":
				wv.topBorder = newTileID(ws, t.ID)
			case "bottom":
				wv.face = newTileID(ws, t.ID)
			case "left_top":
				wv.leftTop = newTileID(ws, t.ID)
			case "top":
				wv.top = newTileID(ws, t.ID)
			case "right_top":
				wv.rightTop = newTileID(ws, t.ID)
			case "left_bottom":
				wv.leftBottom = newTileID(ws, t.ID)
			case "right_bottom":
				wv.rightBottom = newTileID(ws, t.ID)
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
