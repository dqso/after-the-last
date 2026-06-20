package random

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
