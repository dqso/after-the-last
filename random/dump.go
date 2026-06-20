package random

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// dumpMap writes all grids to map_<datetime>.txt in a plain dump format:
// grids are separated by a blank line, cells in a row by a single space.
// Values are decimal tile indices with the sheet id stripped (low 16 bits).
// Every cell is padded to the same width; empty cells (none) become spaces.
func dumpMap(grids ...[][]int) error {
	// Column width is the widest decimal tile index across all grids.
	width := 1
	for _, grid := range grids {
		for _, row := range grid {
			for _, v := range row {
				if v == none {
					continue
				}
				if w := len(strconv.Itoa(v & 0xFFFF)); w > width {
					width = w
				}
			}
		}
	}

	var b strings.Builder
	for i, grid := range grids {
		if i > 0 {
			b.WriteString("\n")
		}
		for _, row := range grid {
			for x, v := range row {
				if x > 0 {
					b.WriteByte(' ')
				}
				if v == none {
					b.WriteString(strings.Repeat(" ", width))
				} else {
					fmt.Fprintf(&b, "%*d", width, v&0xFFFF)
				}
			}
			b.WriteByte('\n')
		}
	}
	name := fmt.Sprintf("map_%s.txt", time.Now().Format("20060102_150405"))
	return os.WriteFile(name, []byte(b.String()), 0o644)
}
