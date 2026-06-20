package random

import (
	"math/rand"

	"github.com/dqso/after-the-last/collision"
)

// room is a rectangular area described by its inner floor bounds (inclusive).
// The wall ring is derived later and lives just outside this rectangle:
// two rows above (face + top stripe), one row below, one column on each side.
type room struct {
	x0, y0, x1, y1 int
}

func (r room) centerX() int { return (r.x0 + r.x1) / 2 }
func (r room) centerY() int { return (r.y0 + r.y1) / 2 }
func (r room) width() int   { return r.x1 - r.x0 + 1 }
func (r room) height() int  { return r.y1 - r.y0 + 1 }

// footprint returns the room rectangle expanded to include its wall ring.
func (r room) footprint() (x0, y0, x1, y1 int) {
	return r.x0 - 1, r.y0 - 2, r.x1 + 1, r.y1 + 1
}

// GenerateV2 builds a small dungeon of several rooms with random sizes and
// designs, connected by straight tunnels. Each room reuses the V1 wall layout.
func (g Generator) GenerateV2(seed int64) ([][]int, [][]int, [][]int, [][]int) {
	const (
		minRooms = 3
		maxRooms = 10

		minRoomW = 6
		maxRoomW = 16
		minRoomH = 5
		maxRoomH = 12

		minTunnel = 5
		maxTunnel = 10

		margin  = 2 // empty gap kept between room footprints
		padding = 3 // empty border around the whole map
	)

	rnd := rand.New(rand.NewSource(seed))
	roomCount := minRooms + rnd.Intn(maxRooms-minRooms+1)

	// 1. Place rooms in an unbounded coordinate space.
	// Each new room buds off an existing one in a cardinal direction, leaving a
	// gap of tunnel length between them. The first room sits at the origin.
	type tunnel struct {
		a, b       int  // indices of the two linked rooms
		horizontal bool // true if the rooms sit side by side
	}
	rooms := make([]room, 0, roomCount)
	tunnels := make([]tunnel, 0, roomCount-1)

	makeRoom := func(cx, cy, w, h int) room {
		x0, y0 := cx-w/2, cy-h/2
		return room{x0: x0, y0: y0, x1: x0 + w - 1, y1: y0 + h - 1}
	}

	rooms = append(rooms, makeRoom(0, 0, randRange(rnd, minRoomW, maxRoomW), randRange(rnd, minRoomH, maxRoomH)))

	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for attempts := 0; len(rooms) < roomCount && attempts < roomCount*50; attempts++ {
		parent := rooms[rnd.Intn(len(rooms))]
		d := dirs[rnd.Intn(len(dirs))]
		gap := randRange(rnd, minTunnel, maxTunnel)
		w, h := randRange(rnd, minRoomW, maxRoomW), randRange(rnd, minRoomH, maxRoomH)

		// Offset the new center away from the parent so the gap between the two
		// footprints is about one tunnel long. Keep the perpendicular axis
		// aligned with the parent so the tunnel stays straight.
		var cx, cy int
		if d[0] != 0 {
			cx = parent.centerX() + d[0]*(parent.width()/2+gap+w/2)
			cy = parent.centerY()
		} else {
			cx = parent.centerX()
			cy = parent.centerY() + d[1]*(parent.height()/2+gap+h/2)
		}
		cand := makeRoom(cx, cy, w, h)

		if overlapsAny(cand, rooms, margin) {
			continue
		}
		// Link to the parent by its index (parent is the element we copied).
		for i := range rooms {
			if rooms[i] == parent {
				tunnels = append(tunnels, tunnel{a: i, b: len(rooms), horizontal: d[0] != 0})
				break
			}
		}
		rooms = append(rooms, cand)
	}

	// 2. Normalize coordinates so the map starts at (padding, padding).
	minX, minY, maxX, maxY := boundingBox(rooms)
	offX, offY := padding-minX, padding-minY
	for i := range rooms {
		rooms[i].x0 += offX
		rooms[i].x1 += offX
		rooms[i].y0 += offY
		rooms[i].y1 += offY
	}
	gridW := (maxX - minX) + 1 + 2*padding
	gridH := (maxY - minY) + 1 + 2*padding

	// Allocate output grids. Floor/walls/items default to none, collision to
	// Blocked (the void between rooms is impassable).
	floorGrid := newGrid(gridW, gridH, none)
	wallsGrid := newGrid(gridW, gridH, none)
	itemsGrid := newGrid(gridW, gridH, none)
	collisionGrid := newGrid(gridW, gridH, collision.Blocked)

	// owner[y][x] is the index of the room a walkable cell belongs to, or -1.
	// Tunnel cells inherit the design of the room they were dug from.
	walk := make([][]bool, gridH)
	owner := make([][]int, gridH)
	tunnelFloor := make([][]bool, gridH) // true for cells carved as a tunnel
	for y := range walk {
		walk[y] = make([]bool, gridW)
		owner[y] = make([]int, gridW)
		tunnelFloor[y] = make([]bool, gridW)
		for x := range owner[y] {
			owner[y][x] = -1
		}
	}

	// 3a. Carve room interiors.
	for i, r := range rooms {
		for y := r.y0; y <= r.y1; y++ {
			for x := r.x0; x <= r.x1; x++ {
				walk[y][x] = true
				owner[y][x] = i
			}
		}
	}

	// 3b. Carve tunnels between rooms. Each tunnel is a single straight line at a
	// shared row / column picked from the overlap of both rooms' interiors, kept
	// at least two tiles from the corners so a doorway's two-tall wall never
	// collides with the room's own corner. Cutting through wall rings makes the
	// doorways for free.
	dig := func(x, y, parent int) {
		if y < 0 || y >= gridH || x < 0 || x >= gridW {
			return
		}
		walk[y][x] = true
		if owner[y][x] == -1 {
			owner[y][x] = parent
			tunnelFloor[y][x] = true
		}
	}
	carveH := func(x0, x1, y, parent int) {
		for x := min(x0, x1); x <= max(x0, x1); x++ {
			dig(x, y, parent)
		}
	}
	carveV := func(x, y0, y1, parent int) {
		for y := min(y0, y1); y <= max(y0, y1); y++ {
			dig(x, y, parent)
		}
	}
	for _, t := range tunnels {
		a, b := rooms[t.a], rooms[t.b]
		if t.horizontal {
			lo, hi := max(a.y0, b.y0)+2, min(a.y1, b.y1)-2
			y := (a.centerY() + b.centerY()) / 2
			if lo <= hi {
				y = randRange(rnd, lo, hi)
			}
			carveH(a.centerX(), b.centerX(), y, t.a)
		} else {
			lo, hi := max(a.x0, b.x0)+2, min(a.x1, b.x1)-2
			x := (a.centerX() + b.centerX()) / 2
			if lo <= hi {
				x = randRange(rnd, lo, hi)
			}
			carveV(x, a.centerY(), b.centerY(), t.a)
		}
	}

	// Pick a random floor and wall design for every room.
	roomFloor := make([]int, len(rooms))
	roomWall := make([]int, len(rooms))
	for i := range rooms {
		roomFloor[i] = rnd.Intn(len(g.floors))
		roomWall[i] = rnd.Intn(len(g.walls))
	}

	walkAt := func(x, y int) bool {
		return y >= 0 && y < gridH && x >= 0 && x < gridW && walk[y][x]
	}
	// setWall writes a wall tile only into a non-walkable cell, with collision.
	setWall := func(x, y, tile, coll int) {
		if y < 0 || y >= gridH || x < 0 || x >= gridW || walk[y][x] {
			return
		}
		wallsGrid[y][x] = tile
		collisionGrid[y][x] = coll
	}

	// 4a. Floor tiles. The edge tiles depend on whether the north / west
	// neighbour is walkable, generalizing the V1 corner/edge selection.
	for y := 0; y < gridH; y++ {
		for x := 0; x < gridW; x++ {
			if !walk[y][x] {
				continue
			}
			f := g.floors[roomFloor[ownerOr(owner, x, y, 0)]]
			n, w := walkAt(x, y-1), walkAt(x-1, y)
			switch {
			case !n && !w:
				floorGrid[y][x] = f.leftTop
			case !n:
				floorGrid[y][x] = f.top
			case !w:
				floorGrid[y][x] = f.left
			default:
				floorGrid[y][x] = f.clear
			}
			collisionGrid[y][x] = collision.Free
		}
	}

	// 4b. Wall edges. Driven from each walkable cell into its blocked
	// neighbours; the north wall is two tiles tall (face + top stripe).
	wc := g.wallCommon
	for y := 0; y < gridH; y++ {
		for x := 0; x < gridW; x++ {
			if !walk[y][x] {
				continue
			}
			wv := g.walls[roomWall[ownerOr(owner, x, y, 0)]]

			// A horizontal tunnel mouth touches room floor on one side while its
			// north wall is present. Left mouth = room opens to the right,
			// right mouth = room opens to the left.
			northWall := !walkAt(x, y-1)
			leftMouth := northWall && wv.leftTop != 0 && walkAt(x-1, y) && walkAt(x-1, y-1)
			rightMouth := northWall && wv.rightTop != 0 && walkAt(x+1, y) && walkAt(x+1, y-1)

			if northWall {
				// The north wall is two tiles tall (upper border row + face).
				// Its tiles depend on what the wall borders:
				//   - a tunnel mouth gets the variant's decorated corner;
				//   - a straight tunnel top uses the plain `top` tile;
				//   - a room top uses the decorative `top_border`.
				switch {
				case leftMouth:
					setWall(x, y-2, wv.leftTop, collision.Blocked)
					setWall(x, y-1, wv.leftBottom, collision.Blocked)
				case rightMouth:
					setWall(x, y-2, wv.rightTop, collision.Blocked)
					setWall(x, y-1, wv.rightBottom, collision.Blocked)
				case tunnelFloor[y][x]:
					setWall(x, y-2, wv.top, collision.Blocked)
					setWall(x, y-1, wv.face, collision.Blocked)
				default:
					setWall(x, y-2, wv.topBorder, collision.Blocked)
					setWall(x, y-1, wv.face, collision.Blocked)
				}
			}
			if !walkAt(x, y+1) {
				setWall(x, y+1, wc.bottom, collision.BlockedBot)
			}
			if !walkAt(x-1, y) {
				setWall(x-1, y, wc.left, collision.Blocked)
			}
			if !walkAt(x+1, y) {
				setWall(x+1, y, wc.right, collision.Blocked)
			}
		}
	}

	// 4c. Convex corners overwrite the edge tiles placed above.
	for y := 0; y < gridH; y++ {
		for x := 0; x < gridW; x++ {
			if !walk[y][x] {
				continue
			}
			n, s := !walkAt(x, y-1), !walkAt(x, y+1)
			w, e := !walkAt(x-1, y), !walkAt(x+1, y)
			if n && w {
				setWall(x-1, y-2, wc.cornerLeftTop, collision.Blocked)
				setWall(x-1, y-1, wc.left, collision.Blocked)
			}
			if n && e {
				setWall(x+1, y-2, wc.cornerRightTop, collision.Blocked)
				setWall(x+1, y-1, wc.right, collision.Blocked)
			}
			if s && w {
				setWall(x-1, y+1, wc.cornerLeftBot, collision.Blocked)
			}
			if s && e {
				setWall(x+1, y+1, wc.cornerRightBot, collision.Blocked)
			}
		}
	}

	// 4d. Concave (inner) corners. A blocked cell with floor on its north side
	// plus floor to the west or east is where two walls meet at the bottom of an
	// opening (tunnel mouth, room edge joining a passage). Round it with the
	// matching big corner.
	for y := 0; y < gridH; y++ {
		for x := 0; x < gridW; x++ {
			if walk[y][x] || !walkAt(x, y-1) {
				continue
			}
			switch {
			case walkAt(x-1, y):
				setWall(x, y, wc.bigCornerLeftTop, collision.BlockedBot)
			case walkAt(x+1, y):
				setWall(x, y, wc.bigCornerLeftRight, collision.BlockedBot)
			}
		}
	}

	// 5. Place one button on the inner-top wall face of every room (like V1).
	for _, r := range rooms {
		buttonCol := r.x0 + rnd.Intn(r.width())
		itemsGrid[r.y0-1][buttonCol] = ButtonInactiveTileID
	}

	// Dump the generated grids to a file for inspection.
	//_ = dumpMap(floorGrid, wallsGrid, itemsGrid, collisionGrid)

	return floorGrid, wallsGrid, itemsGrid, collisionGrid
}

// randRange returns a random int in [lo, hi].
func randRange(rnd *rand.Rand, lo, hi int) int {
	return lo + rnd.Intn(hi-lo+1)
}

// newGrid allocates a height*width grid filled with fill.
func newGrid(width, height, fill int) [][]int {
	g := make([][]int, height)
	for y := range g {
		g[y] = make([]int, width)
		for x := range g[y] {
			g[y][x] = fill
		}
	}
	return g
}

// overlapsAny reports whether cand touches any existing room within margin.
func overlapsAny(cand room, rooms []room, margin int) bool {
	for _, r := range rooms {
		if roomsOverlap(cand, r, margin) {
			return true
		}
	}
	return false
}

// roomsOverlap reports whether two room footprints intersect within margin.
func roomsOverlap(a, b room, margin int) bool {
	ax0, ay0, ax1, ay1 := a.footprint()
	bx0, by0, bx1, by1 := b.footprint()
	return ax0-margin <= bx1 && bx0 <= ax1+margin &&
		ay0-margin <= by1 && by0 <= ay1+margin
}

// boundingBox returns the min/max footprint corners across all rooms.
func boundingBox(rooms []room) (minX, minY, maxX, maxY int) {
	first := true
	for _, r := range rooms {
		x0, y0, x1, y1 := r.footprint()
		if first {
			minX, minY, maxX, maxY = x0, y0, x1, y1
			first = false
			continue
		}
		minX, minY = min(minX, x0), min(minY, y0)
		maxX, maxY = max(maxX, x1), max(maxY, y1)
	}
	return minX, minY, maxX, maxY
}

// ownerOr returns owner[y][x], or fallback when the cell has no owner.
func ownerOr(owner [][]int, x, y, fallback int) int {
	if o := owner[y][x]; o >= 0 {
		return o
	}
	return fallback
}
