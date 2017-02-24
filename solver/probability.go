package solver

// calculate probability on all cells bordering probed cells
func PrimedFieldProbability(mf *Minefield) bool {
	dirtyProbability := false

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
