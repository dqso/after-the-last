package collision

// Type classifies how much of a tile is impassable.
type Type = int

const (
	Free         Type = 0 // passable
	Blocked      Type = 1 // fully blocked
	BlockedTL    Type = 2 // top-left quarter blocked
	BlockedTR    Type = 3 // top-right quarter blocked
	BlockedBL    Type = 4 // bottom-left quarter blocked
	BlockedBR    Type = 5 // bottom-right quarter blocked
	BlockedTop   Type = 6 // upper half blocked
	BlockedBot   Type = 7 // lower half blocked
	BlockedLeft  Type = 8 // left half blocked
	BlockedRight Type = 9 // right half blocked
)
