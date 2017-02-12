package solver

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/lelandbatey/minesweeper-solver/defusedivision"
)

var rawblob = []byte(`{"width": 5, "height": 5, "cells": [{"neighbors": {"SW": null, "E": false, "SE": false, "NE": null, "S": false, "N": null, "W": null, "NW": null}, "y": 0, "x": 0, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": null, "E": false, "SE": false, "NE": false, "S": false, "N": false, "W": null, "NW": null}, "y": 1, "x": 0, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": null, "E": false, "SE": true, "NE": false, "S": false, "N": false, "W": null, "NW": null}, "y": 2, "x": 0, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": null, "E": true, "SE": false, "NE": false, "S": false, "N": false, "W": null, "NW": null}, "y": 3, "x": 0, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": null, "E": false, "SE": null, "NE": true, "S": null, "N": false, "W": null, "NW": null}, "y": 4, "x": 0, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": true, "SE": false, "NE": null, "S": false, "N": null, "W": false, "NW": null}, "y": 0, "x": 1, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": false, "SE": false, "NE": true, "S": false, "N": false, "W": false, "NW": false}, "y": 1, "x": 1, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": false, "SE": false, "NE": false, "S": true, "N": false, "W": false, "NW": false}, "y": 2, "x": 1, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": false, "SE": false, "NE": false, "S": false, "N": false, "W": false, "NW": false}, "y": 3, "x": 1, "contents": " b ", "flagged": false, "probed": false}, {"neighbors": {"SW": null, "E": false, "SE": null, "NE": false, "S": null, "N": true, "W": false, "NW": false}, "y": 4, "x": 1, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": false, "SE": false, "NE": null, "S": false, "N": null, "W": false, "NW": null}, "y": 0, "x": 2, "contents": " b ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": false, "SE": true, "NE": false, "S": false, "N": true, "W": false, "NW": false}, "y": 1, "x": 2, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": true, "E": true, "SE": false, "NE": false, "S": false, "N": false, "W": false, "NW": false}, "y": 2, "x": 2, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": false, "SE": false, "NE": true, "S": false, "N": false, "W": true, "NW": false}, "y": 3, "x": 2, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": null, "E": false, "SE": null, "NE": false, "S": null, "N": false, "W": false, "NW": true}, "y": 4, "x": 2, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": false, "SE": false, "NE": null, "S": false, "N": null, "W": true, "NW": null}, "y": 0, "x": 3, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": false, "SE": false, "NE": false, "S": true, "N": false, "W": false, "NW": true}, "y": 1, "x": 3, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": false, "SE": false, "NE": false, "S": false, "N": false, "W": false, "NW": false}, "y": 2, "x": 3, "contents": " b ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": false, "SE": false, "NE": false, "S": false, "N": true, "W": false, "NW": false}, "y": 3, "x": 3, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": null, "E": false, "SE": null, "NE": false, "S": null, "N": false, "W": false, "NW": false}, "y": 4, "x": 3, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": null, "SE": null, "NE": null, "S": false, "N": null, "W": false, "NW": null}, "y": 0, "x": 4, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": true, "E": null, "SE": null, "NE": null, "S": false, "N": false, "W": false, "NW": false}, "y": 1, "x": 4, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": null, "SE": null, "NE": null, "S": false, "N": false, "W": true, "NW": false}, "y": 2, "x": 4, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": false, "E": null, "SE": null, "NE": null, "S": false, "N": false, "W": false, "NW": true}, "y": 3, "x": 4, "contents": "   ", "flagged": false, "probed": false}, {"neighbors": {"SW": null, "E": null, "SE": null, "NE": null, "S": null, "N": false, "W": false, "NW": false}, "y": 4, "x": 4, "contents": "   ", "flagged": false, "probed": false}], "selected": [0, 0], "mine_count": 3}`)

func TestNewMinefield(t *testing.T) {
	var ddc defusedivision.Minefield
	err := json.Unmarshal(rawblob, &ddc)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewMinefield(ddc)
	if err != nil {
		t.Fatal(err)
	}

}

func TestCellTouchingCounts(t *testing.T) {
	var ddc defusedivision.Minefield
	err := json.Unmarshal(rawblob, &ddc)
	if err != nil {
		t.Fatal(err)
	}
	sort.Sort(ByCoords(ddc.Cells))

	// Mark all the cells as probed
	for _, cell := range ddc.Cells {
		cell.Probed = true
	}

	mf, err := NewMinefield(ddc)
	if err != nil {
		t.Fatal(err)
	}

	countmines := func(x map[string]*bool) int {
		rv := 0
		for _, v := range x {
			if v != nil && *v == true {
				rv += 1
			}
		}
		return rv
	}

	for idx, c := range ddc.Cells {
		limitedCell := mf.Cells[idx]
		// Ensure that the ordering of cells in the new minefield is the same
		// as the order of cells in the old minefield.
		if !(c.X == limitedCell.X && c.Y == limitedCell.Y) {
			t.Errorf("Cells have differing coords; expected (%d, %d), found (%d, %d)", c.X, c.Y, limitedCell.X, limitedCell.Y)
		}
		// Ensure the calculated number of mines this cell is touching is the
		// same as the number from the original cell.
		if !(countmines(c.Neighbors) == limitedCell.MineTouch) {
			t.Errorf("Cells touching different number of mines; expected %d, found %d", countmines(c.Neighbors), limitedCell.MineTouch)
		}

		if c.X == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf(" %d ", countmines(c.Neighbors))
	}
	fmt.Printf("\n")
}

func TestCellInfoLimits(t *testing.T) {
	fake := defusedivision.Cell{}
	fc, err := NewCell(&fake)
	if err != nil {
		t.Fatal(err)
	}
	// verify can't see # of touching mines
	if fc.MineTouch >= 0 {
		t.Errorf("Cell has a defined # of touching mines. Expected 0, found %d", fc.MineTouch)
	}
}

func TestCellNeighbors(t *testing.T) {
	var ddc defusedivision.Minefield
	err := json.Unmarshal(rawblob, &ddc)
	if err != nil {
		t.Fatal(err)
	}
	// NewMinefield sorts the cells
	mf, err := NewMinefield(ddc)
	// now test that directional neighbors work as expected
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
	for _, cell := range mf.Cells {
		neighborCount := 0
		for direction, neighbor := range cell.Neighbors {
			delt := deltas[direction]
			dX, dY := delt[0], delt[1]
			calcX, calcY := dX+cell.X, dY+cell.Y
			if neighbor == nil {
				oor := func(i, limit int) bool {
					return i >= limit || i < 0
				}
				// if neighbor is nil, at least one of the coordinates must be out of range
				if !(oor(calcX, mf.Width) || oor(calcY, mf.Height)) {
					t.Errorf("Neighbor is nil, but X and Y (%d, %d) are in range", calcX, calcY)
				}
				// Skip comparing coordinates of nil neighbor
				continue
			}
			// verify that calculated neighbor coordinates match server-issued coordinates
			if !(calcX == neighbor.X && calcY == neighbor.Y) {
				t.Errorf("Neighbor coords (%d, %d) don't match server coords (%d, %d)",
					calcX, calcY, neighbor.X, neighbor.Y)
			}
			neighborCount += 1
		}
		// Ensure that not *all* the neighbors are nil
		if neighborCount < 3 || neighborCount > 9 {
			t.Errorf("number of neighbors should be > 3 and < 9")
		}
	}
}
