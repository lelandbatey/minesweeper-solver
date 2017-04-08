package solver

/* let's talk about symantics. When we say "Primed Field", we refer to a
 * minefield that has been partially exposed. The cells bordering on numbers
 * are "primed" because they may or may not contain a mine, and we therefore
 * call it primed because it's potentially an explodable mine
 */

func GetSafestCell(mf *Minefield) *Cell {
	unprobed := GetUnmarkedCells(mf)
	var safest *Cell = nil
	for _, cell := range unprobed {
		// skip iteration if we have an unkown mine probability
		// (means it's far away from revealed #'s)
		// This effectively filters out all non-primed cells
		if cell.MineProb == -1.0 {
			continue
		}
		if safest == nil {
			safest = cell
		}
		// keep trading safest with cell of lesser MineProb.
		// this has the effect that we'll choose the safest, lowest (x,y) cell
		// due to how we iterate with increasing x,y coords
		if cell.MineProb < safest.MineProb {
			safest = cell
		}
	}
	return safest
}

// return "unmarked" cells that have neither been probed nor flagged
func GetUnmarkedCells(mf *Minefield) []*Cell {
	unprobed := []*Cell{}
	for _, cell := range mf.Cells {
		if !cell.Probed && !cell.Flagged {
			unprobed = append(unprobed, cell)
		}
	}
	return unprobed
}

func UnflaggedMines(mf *Minefield) []*Cell {
	unflagged := []*Cell{}
	for _, cell := range mf.Cells {
		if cell.MineProb == 1.0 && !cell.Flagged {
			unflagged = append(unflagged, cell)
		}
	}
	return unflagged
}

// fully calculate minefield probability given current conditions
// this modified the minefiled by pointer, so nothing needs returning
func PrimedFieldProbability(mf *Minefield) {
	for IterPrimedFieldProbability(mf) == true {
	}
	return
}

// calculate probability on all cells bordering probed cells
// It returns each iteration with a boolean indicating that it still has
// calculations to do. This is because it must recalculate the board after
// it calculates with 100% certainty that a particular cell has a mine.
func IterPrimedFieldProbability(mf *Minefield) bool {
	dirtyProbability := false

	// primed cells is a map of (x,y) to cell which gives us unique cells
	// due to the fact that re-adding the same cell will overwrite itself
	// because the key (x,y) is the same
	primedCells := map[[2]int]*Cell{}
	for _, cell := range mf.Cells {
		if cell.Probed {
			cell.MineProb = 0.0
			// add all unique surrounding unprobed cells to `primed` list
			for _, neighbor := range cell.Neighbors {
				if neighbor != nil && !neighbor.Probed {
					primedCells[[2]int{neighbor.X, neighbor.Y}] = neighbor
				}
			}
		}
	}
	// first pass calculating probabilities
	for _, omega := range primedCells {
		prob := MineProbability(omega)
		if prob == 1.0 && omega.MineProb != 1.0 {
			// omega has just been declared a mine! Recalculate board
			dirtyProbability = true
		}
		omega.MineProb = prob
	}
	return dirtyProbability
}

// MineProbability accepts a Cell (omega) and calculates the probability that
// omega contains a mine by querying its neighbors to determine the probability
// they see for each neighbor.
func MineProbability(omega *Cell) float64 {
	if omega.MineProb == 1.0 || omega.MineProb == 0.0 {
		// probability is already known (for sure!)
		return omega.MineProb
	}
	validNeighborsCount := 0
	probSum := 0.0
	for _, neighbor := range omega.Neighbors {
		// Neighbors touching zero mines are neighbors we know nothing about.
		if neighbor == nil || neighbor.MineTouch == -1 {
			continue
		}
		validNeighborsCount += 1
		prob := GetNeighborsUnknownProbability(neighbor)
		if prob == 1.0 || prob == 0.0 {
			return prob
		}
		probSum += prob
		// this neighbor now queries its neighbors to determine the # of
		// surrounding un-accounted-for mines
	}
	// return an average of the reported probabilities
	return probSum / float64(validNeighborsCount)
}

// GetNeighborsUnknownProbability accepts a cell (gamma) and returns the
// probability of gamma's unknown neighbors containing mines.
func GetNeighborsUnknownProbability(gamma *Cell) float64 {
	neighborsCount := 0
	unknownNeighborsCount := 0
	unknownMinesCount := gamma.MineTouch
	for _, cell := range gamma.Neighbors {
		if cell == nil {
			continue
		}
		neighborsCount += 1
		if cell.Probed {
			continue
		}
		if cell.MineProb != 1.0 || cell.MineProb != 0.0 {
			unknownNeighborsCount += 1
		}
		if cell.MineProb == 1.0 {
			unknownMinesCount -= 1
		}
	}
	if unknownMinesCount <= 0 {
		return 0
	}
	if unknownNeighborsCount == 0 {
		return 0
	}
	// probability that unknown neighbors contain a mine
	return float64(unknownMinesCount) / float64(unknownNeighborsCount)
}
