package game

import "github.com/dqso/after-the-last/collision"

// blockedSubrect returns the blocked region of a tile as fractions [0,1] of (tileW, tileH).
func blockedSubrect(ct collision.Type) (x0, y0, x1, y1 float64) {
	switch ct {
	case collision.Blocked:
		return 0, 0, 1, 1
	case collision.BlockedTL:
		return 0, 0, 0.5, 0.5
	case collision.BlockedTR:
		return 0.5, 0, 1, 0.5
	case collision.BlockedBL:
		return 0, 0.5, 0.5, 1
	case collision.BlockedBR:
		return 0.5, 0.5, 1, 1
	case collision.BlockedTop:
		return 0, 0, 1, 0.5
	case collision.BlockedBot:
		return 0, 0.5, 1, 1
	case collision.BlockedLeft:
		return 0, 0, 0.5, 1
	case collision.BlockedRight:
		return 0.5, 0, 1, 1
	default:
		return 0, 0, 0, 0
	}
}
