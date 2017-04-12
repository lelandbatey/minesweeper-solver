package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	//"github.com/davecgh/go-spew/spew"
	"github.com/lelandbatey/minesweeper-solver/client"
	"github.com/lelandbatey/minesweeper-solver/defusedivision"
	"github.com/lelandbatey/minesweeper-solver/solver"
	"github.com/y0ssar1an/q"
)

func main() {

	// add default arguments to connect to local-server if none supplied
	os.Args = append(os.Args, "127.0.0.1", "44444")
	host := os.Args[1]
	port := os.Args[2]
	c, err := client.New(host, port)
	go client.NetReader(c)
	if err != nil {
		panic(err)
	}
	time.Sleep(400 * time.Millisecond)
	fmt.Printf("We did it, we opened a client named %s!\n", c.Name)
	q.Q("Now we wait for a message:")
	// Get the first message, which is the player struct for ourself. This also
	// causes our client to modify itself by changing it's name to the name of
	// the player sent by the server.
	c.Message()
	// The second message will be the full state from the server.
	//	c.Send(`
	//{
	//	"new-minefield": {
	//		"height": 25,
	//		"width": 25,
	//		"mine_count": 80
	//	}
	//}
	//	`)
	//	spew.Dump(c.Message())
	//	time.Sleep(400 * time.Millisecond)
	fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
	state := defusedivision.State{}
	for {
		time.Sleep(400 * time.Millisecond)
		c.Send("PROBE")

		// Will contain a new state
		msg := c.Message()
		state = msg.(defusedivision.State)
		player := state.Players[c.Name]
		x := player.Field.Selected[0]
		y := player.Field.Selected[1]
		if player.Living == false {
			fmt.Printf("aw... Exploded @ (%v, %v)\n", x, y)
			break
		}
		board := player.Field
		sboard, err := solver.NewMinefield(board)
		if err != nil {
			panic(err)
		}
		_ = sboard
		// find the probability of cells containing a mine
		solver.PrimedFieldProbability(sboard)
		fmt.Println(sboard.Render())

		// flag all cells that 100% contain a mine
		for _, unflaggedCell := range solver.UnflaggedMines(sboard) {
			x := unflaggedCell.X
			y := unflaggedCell.Y
			c.MoveToXY(x, y)
			time.Sleep(100 * time.Millisecond)
			c.Send("FLAG")
			fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
			time.Sleep(100 * time.Millisecond)
		}

		// find lowest-probability cell to probe
		safest := solver.GetSafestCell(sboard)
		stillNotCertain := true
		if safest.MineProb > 0.0 {
			// if probability isn't 0, then we should try hypothetical scenarios
			// to see if we can nail down a cell that cannot be a mine
			// we do this by setting a mine probability to 1.0 and then
			// running the probability calculations. Then we ask if any
			// "impossible" scenario has happened: such as too many nearby
			// mines...
			// If so, then our hypothetical mine cannot be a mine, since it
			// violates our knowledge of the board by putting too many mines
			// nearby a revealed #
			witnesses := solver.GetWitnesses(sboard)
			primedCells := solver.GetPrimedCells(sboard)
			// get non 1.0 or 0.0 probability for mines & get cells
			// that don't immediately clash with witnesses if they are
			// marked as a mine
			primedCells = solver.GetValidPrimedCells(witnesses, primedCells)
			// keep track of which cells were valid hypothetical mines
			// once we have a success in satisfying all witnesses, add all
			// hypothetically-flagged cells to this list and avoid using them
			// again; it'll just find the same solution (or another) but
			// definitely will NOT find a failure
			// (basically, optimize to skip already known successes)
			validHypotheticalMine := map[[2]int]bool{}
			for _, cell := range primedCells {
				validHypotheticalMine[[2]int{cell.X, cell.Y}] = false
			}
			likelyhood := map[uint]*solver.Cell{}
			bestScenario := uint(0) // get min integer
			fmt.Printf("size of primed cells: %v\n", len(primedCells))
			fmt.Printf("size of witnesses: %v\n", len(witnesses))
			for _, hypothetical := range primedCells {
				coordinates := [2]int{hypothetical.X, hypothetical.Y}
				// skip testing this cell if it's already been proven to be
				// part of a successful hypothetically-flagged scenario
				if validHypotheticalMine[coordinates] {
					continue
				}
				// indicate "thinking". Useful when script takes a looong time
				fmt.Printf(".")
				tempProb := hypothetical.MineProb
				// mark this cell as a mine
				hypothetical.MineProb = 1.0
				flags, scenarios, success, err := solver.SatisfyWitnesses(witnesses, primedCells)
				_ = err
				hypothetical.MineProb = tempProb
				if !success {
					fmt.Printf("found a hypothetical non-mine of ")
					safest = hypothetical
					stillNotCertain = false
					break
				} else {
					for coordinates, _ := range flags {
						validHypotheticalMine[coordinates] = true
					}
				}
				likelyhood[scenarios] = hypothetical
				if scenarios > bestScenario {
					bestScenario = scenarios
				}
			}
			if stillNotCertain {
				fmt.Printf("well I'm still not certain, but the best scenario")
				fmt.Printf(" had %v scenarios\n", bestScenario)
				safest = likelyhood[bestScenario]
			}
		}
		x = safest.X
		y = safest.Y
		fmt.Printf("%v @ (%v, %v)\n", safest.MineProb, x, y)
		c.MoveToXY(x, y)
	}

	// and now it looks like we again step to the right
	time.Sleep(400 * time.Millisecond)
	//	spew.Dump(state)
	time.Sleep(10 * time.Second)
}
