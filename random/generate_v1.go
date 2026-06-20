package random

import (
	"math/rand"

	"github.com/dqso/after-the-last/collision"
)

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
		wallsGrid[0][col] = wall.topBorder
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
