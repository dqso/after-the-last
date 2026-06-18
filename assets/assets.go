package assets

import _ "embed"

// Resources from https://limezu.itch.io/moderninteriors

//go:embed "Modern tiles_Free/Characters_free/Bob_idle_16x16.png"
var BobIdle []byte

//go:embed "Modern tiles_Free/Characters_free/Bob_run_16x16.png"
var BobRun []byte

//go:embed "Modern tiles_Free/Interiors_free/16x16/Interiors_free_16x16.png"
var InteriorsSheet []byte

//go:embed "Modern tiles_Free/Interiors_free/16x16/Room_Builder_free_16x16.png"
var RoomBuilderSheet []byte

//go:embed "extra_set_16x16.png"
var ExtraSheet []byte

//go:embed "tiles.xml"
var TilesXML []byte
