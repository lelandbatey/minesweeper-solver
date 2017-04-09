package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/lelandbatey/minesweeper-solver/client"
	"github.com/lelandbatey/minesweeper-solver/defusedivision"
	"github.com/lelandbatey/minesweeper-solver/solver"

	"github.com/davecgh/go-spew/spew"
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
	fmt.Printf("We did it, we opened a client named %s!\n", c.Name)
	q.Q("Now we wait for a message:")
	// Get the first message, which is the player struct for ourself. This also
	// causes our client to modify itself by changing it's name to the name of
	// the player sent by the server.
	c.Message()
	// The second message will be the full state from the server.
	time.Sleep(400 * time.Millisecond)
	fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
	c.MoveRight()
	c.MoveDown()
	c.MoveRight()
	c.MoveDown()
	c.MoveRight()
	for {
		time.Sleep(400 * time.Millisecond)
		c.Send("PROBE")

		// Will contain a new state
		msg := c.Message()
		state := msg.(defusedivision.State)
		player := state.Players[c.Name]
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
			time.Sleep(400 * time.Millisecond)
			c.Send("FLAG")
			fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
			time.Sleep(400 * time.Millisecond)
		}

		// find lowest-probability cell to probe
		safest := solver.GetSafestCell(sboard)
		// if probability isn't 0, then we should try hypothetical scenarios
		// to see if we can nail down a cell that cannot be a mine
		// we do this by setting a mine probability to 1.0 and then running
		// running the probability calculations. Then we ask if any
		// "impossible" scenario has happened: such as too many nearby
		// mines...
		// If so, then our hypothetical mine cannot be a mine, since it
		// violates our knowledge of the board by putting too many mines
		// nearby a revealed #
		if safest.MineProb > 0.0 {
			primedCells := solver.GetPrimedCells(sboard)
			for _, hypothetical := range primedCells {
				// skip altering a cell whose mine-status is already determined
				if hypothetical.MineProb == 0.0 ||
					hypothetical.MineProb == 1.0 {
					continue
				}
				savestate := solver.PreserveProbabilities(sboard)
				// mark this cell as a mine
				hypothetical.MineProb = 1.0
				// re-calculate probabilities... See if this is possible
				solver.PrimedFieldProbability(sboard)
				// do these probabilities have an error?
				hasError := solver.HasImpossibleProbability(sboard)
				// but first -- restore probabilities before manipulation
				solver.RestoreProbabilities(sboard, savestate)
				if hasError {
					// this is it! This cell cannot be a mine...
					// so let's probe this cell instead of our previously
					// calculated low-probability mine
					fmt.Printf("found a hypothetical non-mine at ")
					fmt.Printf("(%v, %v)\n", safest.X, safest.Y)
					safest = hypothetical
					break
				}
			}
		}
		fmt.Printf("%v\n", safest.MineProb)
		x := safest.X
		y := safest.Y
		c.MoveToXY(x, y)
	}

	// and now it looks like we again step to the right
	time.Sleep(400 * time.Millisecond)

	c.MoveRight()
	time.Sleep(400 * time.Millisecond)
	c.MoveRight()
	time.Sleep(400 * time.Millisecond)
	c.MoveRight()
	time.Sleep(400 * time.Millisecond)
	c.MoveRight()
	time.Sleep(400 * time.Millisecond)
	fmt.Println("We are waiting for a message:")
	spew.Dump(c.Message())
	fmt.Println("We successfully read a message!")
	c.Send("PROBE")
	time.Sleep(1 * time.Second)
	spew.Dump(c.Message()) // don't exit just yet or else we won't see the changes
}
