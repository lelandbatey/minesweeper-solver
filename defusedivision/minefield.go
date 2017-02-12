package defusedivision

import "errors"

// this function is designed to return a cell at the given XY of the minefield
func (mf *Minefield) XY(x int, y int) (*Cell, error) {
	// Bounds check
	if x >= mf.Width || x < 0 {
		return nil, errors.New("x out of range")
	}
	if y >= mf.Height || y < 0 {
		return nil, errors.New("y out of range")
	}

	return mf.Cells[x+(y*mf.Width)], nil
}

// return a slice equal to a 3x3 (or smaller) neighborhood around a given XY coordinate
// if the XY given is on the edge of bounds, it will return a slice that samples
// only the valid cells around the XY given.
func (mf *Minefield) Neighborhood(x int, y int) ([][]*Cell, error) {
	// Bounds check
	if x >= mf.Width || x < 0 {
		return nil, errors.New("x out of range")
	}
	if y >= mf.Height || y < 0 {
		return nil, errors.New("y out of range")
	}
	cells := [][]*Cell{}
	for j := y - 1; j < y+2; j++ {
		row := []*Cell{}
		for i := x - 1; i < x+2; i++ {
			if c, err := mf.XY(i, j); err != nil {
				row = append(row, c)
			} else {
				row = append(row, nil)
			}
		}
		// do not add row if our y-coord is above or below minefield
		if len(row) > 0 {
			cells = append(cells, row)
		}
	}
	return cells, nil
}
