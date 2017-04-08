package solver

import (
	"fmt"
	"sort"

	"github.com/lelandbatey/minesweeper-solver/defusedivision"
)

// Minefield contains a more "limited" view of a minefield within a game of
// minesweeper. Instead of *all* the information about the state of each cell,
// it will have more limited info based on what we could see if we were purely
// a player.

type Cell struct {
	// The probability that *this* cell contains a mine, as accumulated from
	// it's neighbors.
	MineProb float64
	// The number of neighbors which contain a mine
	MineTouch int
	X         int
	Y         int
	Probed    bool
	Flagged   bool
	Neighbors map[string]*Cell
}

type Minefield struct {
	Height int
	Width  int
	// Cells will be a correctly-sorted slice of pointers to
	// limited-information Cells.
	Cells []*Cell
}

type Player struct {
	Name   string
	Living bool
	Field  Minefield
}

type State struct {
	Ready   bool
	Players map[string]Player
}

// Pre-sort the cells ordering in minefield

type ByCoords []*defusedivision.Cell

func (c ByCoords) Len() int {
	return len(c)
}
func (c ByCoords) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c ByCoords) Less(i, j int) bool {
	// return true if (Y,X) < (Y2,X2)
	if c[i].Y < c[j].Y {
		return true
	}
	if c[i].Y > c[j].Y {
		return false
	}
	if c[i].X < c[j].X {
		return true
	}
	return false
}

func NewMinefield(mf defusedivision.Minefield) (*Minefield, error) {
	var Cells []*Cell
	// Sort the incoming Cells by their X, Y coordinates
	sort.Sort(ByCoords(mf.Cells))
	for _, c := range mf.Cells {
		nc, err := NewCell(c)
		if err != nil {
			return nil, err
		}
		Cells = append(Cells, nc)
	}
	m := Minefield{
		Height: mf.Height,
		Width:  mf.Width,
		Cells:  Cells,
	}
	// Build each Cell's record of its neighbors
	// (This is the "second sweep" for NewCell)
	deltas := map[string][]int{
		"N":  []int{0, -1},
		"S":  []int{0, 1},
		"W":  []int{-1, 0},
		"E":  []int{1, 0},
		"NW": []int{-1, -1},
		"NE": []int{1, -1},
		"SW": []int{-1, 1},
		"SE": []int{1, 1},
	}
	for _, c := range m.Cells {
		for direction, delta := range deltas {
			dX, dY := delta[0], delta[1]
			Y := c.Y + dY
			X := c.X + dX
			// check bounds: skip if (Y,X) is out-of-bounds of minefield
			if Y >= mf.Height || Y < 0 {
				c.Neighbors[direction] = nil
				continue
			}
			if X >= mf.Width || X < 0 {
				c.Neighbors[direction] = nil
				continue
			}
			// offset is converting 2D coordinates (x,y) into 1D
			// since m.Cells is a linear array
			offset := X + Y*m.Width
			neighbor := m.Cells[offset]
			c.Neighbors[direction] = neighbor
		}

	}
	return &m, nil
}

func NewCell(ddc *defusedivision.Cell) (*Cell, error) {
	var minetouch int
	if !ddc.Probed {
		// we COULD calculate minetouch like below even for this cell
		// but we don't because we aren't supposed to know how many mines
		// an unrevealed cell touches
		minetouch = -1
	} else {
		// ddc.Neighbors is a map of neighbor-direction and boolean
		// indicating if it contains a mine or not. We only want
		// info to tally up how many mines a revealed-cell is touching
		for _, val := range ddc.Neighbors {
			if val != nil && *val == true {
				minetouch += 1
			}
		}
	}
	c := Cell{
		MineProb:  -1.0,
		MineTouch: minetouch,
		X:         ddc.X,
		Y:         ddc.Y,
		Probed:    ddc.Probed,
		Flagged:   ddc.Flagged,
		// Neighbors is initialized in a second sweep
		Neighbors: map[string]*Cell{},
	}
	return &c, nil
}

func (mf *Minefield) Render() string {
	rv := ""
	for _, c := range mf.Cells {
		if c.MineProb == -1.0 {
			rv += " ?  "
		} else if c.Probed {
			if c.MineTouch == 0 {
				rv += " .  "
			} else {
				rv += fmt.Sprintf(" %d  ", c.MineTouch)
			}
		} else {
			rv += fmt.Sprintf("%.1f ", c.MineProb)
		}
		if c.X == mf.Width-1 {
			rv += "\n\n"
		}
	}
	rv += "\n"
	return rv
}
