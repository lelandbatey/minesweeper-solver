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
	fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
	// now we try to send a config to resize the minefield
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
	state := defusedivision.State{}
	_ = state
	player := defusedivision.Player{}
	x := 0
	y := 0
	for {
		// ProbeXY also updates the status of x,y, & living
		state, player, _ = c.ProbeXY(x, y)

		if c.Living == false {
			fmt.Printf("aw... Exploded @ (%v, %v)\n", x, y)
			break
		}
		board := player.Field
		sboard, err := solver.NewMinefield(board)
		if err != nil {
			panic(err)
		}
		// find the probability of cells containing a mine
		solver.PrimedFieldProbability(sboard)
		fmt.Println(sboard.Render())

		// flag all cells that 100% contain a mine
		for _, unflaggedCell := range solver.UnflaggedMines(sboard) {
			x := unflaggedCell.X
			y := unflaggedCell.Y
			c.FlagXY(x, y)
			unflaggedCell.Flagged = true
			fmt.Printf("f") //fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
		}

		// find lowest-probability cell to probe
		safest := solver.GetSafestCell(sboard)
		if safest.MineProb > 0.0 {
			fmt.Println("\nsolve with linear algebra")
			flaglist, safelist := solver.SolveWithReducedRowEchellon(sboard)
			for _, unflaggedCell := range flaglist {
				unflaggedCell.MineProb = 1.0
				if unflaggedCell.Flagged {
					// skip toggling flag if we already think it's a flag
					continue
				}
				unflaggedCell.Flagged = true
				x := unflaggedCell.X
				y := unflaggedCell.Y
				c.FlagXY(x, y)
				fmt.Printf("f") //fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
			}
			for _, safe := range safelist {
				safe.MineProb = 0.0
			}
			// REFIND the probability of cells containing a mine, given that
			// we've probably flagged a few cells and found a few safe cells
			solver.PrimedFieldProbability(sboard)
			safest = solver.GetSafestCell(sboard)
		}
		x = safest.X
		y = safest.Y
		fmt.Printf("%v @ (%v, %v)\n", safest.MineProb, x, y)
	}

	//	spew.Dump(state)
	time.Sleep(10 * time.Second)
}
