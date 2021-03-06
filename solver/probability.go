package solver

import "errors"

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

	primedCells := GetPrimedCells(mf)
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

// return a list of cells bordering reveal #'s. These cells potentially have
// a mine and are the ones which you will be probing
func GetPrimedCells(mf *Minefield) []*Cell {
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
	pcells := []*Cell{}
	for _, cell := range primedCells {
		pcells = append(pcells, cell)
	}
	return pcells
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
		// avoid counting unknown neighbors
		if neighbor == nil || neighbor.MineTouch == -1 {
			continue
		}
		validNeighborsCount += 1
		// this probed neighbor now queries its neighbors to determine the
		// # of surrounding un-accounted-for mines
		prob, _ := GetNeighborsUnknownProbability(neighbor)
		if prob == 1.0 || prob == 0.0 {
			return prob
		}
		probSum += prob
	}
	// return an average of the reported probabilities
	return probSum / float64(validNeighborsCount)
}

// GetNeighborsUnknownProbability accepts a cell (gamma) and returns the
// probability of gamma's unknown neighbors containing mines.
func GetNeighborsUnknownProbability(gamma *Cell) (float64, error) {
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
		if cell.MineProb != 1.0 && cell.MineProb != 0.0 {
			unknownNeighborsCount += 1
		}
		if cell.MineProb == 1.0 {
			unknownMinesCount -= 1
		}
	}
	if unknownMinesCount == 0 {
		return 0, nil
	}
	if unknownMinesCount < 0 {
		// this is where we've calculated an impossibility. If < 0, that means
		// we've run into a situation where we expect X mines, but have
		// marked X + 1 mines nearby. So this means we have something
		// incorrectly marked. Usually this means we've marked a mine where
		// there is None, but it could also be an issue from if we've altered
		// the MineTouch property
		return 0, errors.New("we've found too many mines nearby!")
	}
	if unknownNeighborsCount == 0 {
		return 0, nil
	}
	// probability that unknown neighbors contain a mine
	return float64(unknownMinesCount) / float64(unknownNeighborsCount), nil
}

// find if the minefield has a calculation problem.
func HasImpossibleProbability(mf *Minefield) bool {
	for _, cell := range mf.Cells {
		// we check if revealed cell's information clashes with our current
		// scenario
		if cell.MineTouch == -1 {
			continue
		}
		_, err := GetNeighborsUnknownProbability(cell)
		if err != nil {
			return true
		}
	}
	return false
}

/* from an uncompleted board, attempt to completely satisfy all witnesses
* (cells with #s showing how many mines they are touching) by flagging a cell
* as containing a mine, one at a time (only flag if cell won't clash with
* witnesses and currently-flagged cells). Then call SatisfyWitnesses
* recursively in an attempt to successfully satisfy all witnesses. This will
* return a list of flagged-cells-coordinates that satisfied the witnesses,
* along with the total # of recursive calls it took before successfully
* satisfying witnesses. If not able to satisfy witnesses, then the previously
* flagged cell was incorrect.
 */
func SatisfyWitnesses(witnesses []*Cell, primed []*Cell) (map[[2]int]bool, uint, bool, error) {
	// witness are cells to use in determining witness satisfaction
	// primed are cells not yet marked as containing a mine
	flagged := map[[2]int]bool{}
	retries := uint(0)
	// keep track of how many scenarios were iterated upon if failed
	failedScenarios := uint(0)
	// check if witnesses are satisfied. If so, then we're done!
	satisfied, err := WitnessesAreSatisfied(witnesses)
	if err != nil {
		return flagged, retries, false, errors.New("witnesses clashed before I even started")
	}
	if satisfied {
		return flagged, retries, true, nil
	}
	// some cells will never satisfy witnesses in this current environment
	// weed these out and check only a list of cells that are compatible with
	// the currently marked cells
	valids := GetValidPrimedCells(witnesses, primed)
	// now we mark one of valid-primed as a mine, and see if it'll successfully
	// satisfy the witnesses
	for _, cell := range valids {
		tempProb := cell.MineProb
		cell.MineProb = 1.0
		flags, tries, success, _ := SatisfyWitnesses(witnesses, valids)
		retries += 1
		cell.MineProb = tempProb
		//		if err != nil {
		//			return flagged, retries, false, err
		//		}
		if success {
			// SUCCESS! Meaning... this cell is a possible mine.
			// Tally up # of retries
			retries += tries
			// add cells that were flagged in recursive call(s)
			for k, v := range flags {
				flagged[k] = v
			}
			// add cell used in this loop
			flagged[[2]int{cell.X, cell.Y}] = true
			return flagged, retries, true, nil
		}
		// if there's a lot of scenarios, that suggests that this mine has a
		// higher likelyhood of actually being a mine. Basically, if there's
		// only ONE way to arrange the mines, then it's unlikely that it'll
		// be that way, simply because the odds of random mines being arranged
		// in the least likely way is .. well, less likely.
		// so if all scenarios fail to satisfy witnesses, but do so with few
		// scenarios, that suggests that the arrangement is rare, and the
		// first cell to be hypothetically flagged is least likely to be
		// a mine.
		failedScenarios += tries + 1
	}
	return flagged, failedScenarios, false, nil
}

// witnesses are the cells with #'s displaying how many nearby mines there are
func GetWitnesses(mf *Minefield) []*Cell {
	witnesses := []*Cell{}
	for _, cell := range mf.Cells {
		if cell.MineTouch > 0 {
			witnesses = append(witnesses, cell)
		}
	}
	return witnesses
}

// return a list of cells that can be marked as containing a mine
// without clashing with the witnesses (aka without creating an impossible
// scenario given the numbers displayed on the minefield)
func GetValidPrimedCells(witnesses []*Cell, primed []*Cell) []*Cell {
	valids := []*Cell{}
	for _, cell := range primed {
		if cell.MineProb == 1.0 || cell.MineProb == 0.0 {
			continue
		}
		// temporarily mark cell as mine and see if it'll generate an error
		tempProb := cell.MineProb
		cell.MineProb = 1.0
		_, err := WitnessesAreSatisfied(witnesses)
		cell.MineProb = tempProb
		if err == nil {
			valids = append(valids, cell)
		}
	}
	return valids
}

// build a map of X,Y -> list of cell's witness-neighbors
// but specifically only valid witness neighbors. This makes it very easy
// to mark a primed cell and query it's witness neighbors to determine if
// it causes a clash.
func BuildPrimeWitnessMap(primedCells []*Cell) map[[2]int][]*Cell {
	witnesses := map[[2]int][]*Cell{}
	for _, primed := range primedCells {
		for _, neighbor := range primed.Neighbors {
			if neighbor == nil || neighbor.MineTouch <= 0 {
				continue
			}
			xy := [2]int{primed.X, primed.Y}
			witnesses[xy] = append(witnesses[xy], neighbor)
		}
	}
	return witnesses
}

// returns whether witnesses show 0.0 probability for all unknown neighbors
// (which means that all mines are marked). Also returns error if an error
// occurs (which means too many mines are nearby)
func WitnessesAreSatisfied(witnesses []*Cell) (bool, error) {
	for _, witness := range witnesses {
		probability, err := GetNeighborsUnknownProbability(witness)
		if err != nil {
			return false, errors.New("Too many mines nearby witness")
		}
		if probability > 0.0 {
			return false, nil
		}
	}
	return true, nil
}
