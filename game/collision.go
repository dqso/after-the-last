package game

type CollisionType int

const (
	CollFree         CollisionType = 0 // passable
	CollBlocked      CollisionType = 1 // fully blocked
	CollBlockedTL    CollisionType = 2 // top-left quarter blocked
	CollBlockedTR    CollisionType = 3 // top-right quarter blocked
	CollBlockedBL    CollisionType = 4 // bottom-left quarter blocked
	CollBlockedBR    CollisionType = 5 // bottom-right quarter blocked
	CollBlockedTop   CollisionType = 6 // upper half blocked
	CollBlockedBot   CollisionType = 7 // lower half blocked
	CollBlockedLeft  CollisionType = 8 // left half blocked
	CollBlockedRight CollisionType = 9 // right half blocked
)

// blockedSubrect returns the blocked region of a tile as fractions [0,1] of (tileW, tileH).
func blockedSubrect(ct CollisionType) (x0, y0, x1, y1 float64) {
	switch ct {
	case CollBlocked:
		return 0, 0, 1, 1
	case CollBlockedTL:
		return 0, 0, 0.5, 0.5
	case CollBlockedTR:
		return 0.5, 0, 1, 0.5
	case CollBlockedBL:
		return 0, 0.5, 0.5, 1
	case CollBlockedBR:
		return 0.5, 0.5, 1, 1
	case CollBlockedTop:
		return 0, 0, 1, 0.5
	case CollBlockedBot:
		return 0, 0.5, 1, 1
	case CollBlockedLeft:
		return 0, 0, 0.5, 1
	case CollBlockedRight:
		return 0.5, 0, 1, 1
	default:
		return 0, 0, 0, 0
	}
}
