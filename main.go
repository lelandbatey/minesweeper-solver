package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/lelandbatey/minesweeper-solver/client"
	"github.com/lelandbatey/minesweeper-solver/defusedivision"
	"github.com/lelandbatey/minesweeper-solver/solver"
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
	state := defusedivision.State{}
	_ = state
	player := defusedivision.Player{}
	x := 0
	y := 0
	Mines := map[[2]int]bool{}
	Safed := map[[2]int]bool{}
	for {
		// ProbeXY also updates the status of x,y, & living
		state, player, _ = c.ProbeXY(x, y)

		if c.Living == false {
			fmt.Printf("aw... Exploded @ (%v, %v)\n", x, y)
			break
		}
		sboard, err := solver.NewMinefield(player.Field)
		if err != nil {
			panic(err)
		}
		// find the probability of cells containing a mine
		solver.PrimedFieldProbability(sboard)
		safest := solver.GetSafestCell(sboard)
		if safest == nil {
			fmt.Println("WE WON!!")
			fmt.Println(sboard.Render())
			break
		}
		// write probabilities to board only if we need the help
		if safest.MineProb > 0.0 {
			fmt.Println("Adding Linear Algebra solutions to minefield")
			solver.ApplyKnownMines(sboard, Mines)
			solver.ApplyKnownSafed(sboard, Safed)
		}
		// find mines / safe-spots using linear algebra
		// (by treating each witness as a start to an equation)
		flaglist, safelist := solver.SolveWithReducedRowEchelon(sboard)
		// update list of mines / safe cells
		for coordinates, _ := range flaglist {
			Mines[coordinates] = true
		}
		for coordinates, _ := range safelist {
			Safed[coordinates] = true
		}
		// again, update mines / safe-spots only if we aren't sure.
		// this makes client look like it flags more difficult deductions
		// when it's having trouble
		if safest.MineProb > 0.0 {
			solver.ApplyKnownMines(sboard, Mines)
			solver.ApplyKnownSafed(sboard, Safed)
		}
		// RECALCULATE the probability of cells containing a mine, given that
		// we've probably flagged a few cells and found a few safe cells
		solver.PrimedFieldProbability(sboard)
		safest = solver.GetSafestCell(sboard)
		// draw board (with our calculations complete)
		fmt.Println(sboard.Render())
		// flag all cells that 100% contain a mine
		for _, unflaggedCell := range solver.UnflaggedMines(sboard) {
			x := unflaggedCell.X
			y := unflaggedCell.Y
			c.FlagXY(x, y)
			unflaggedCell.Flagged = true
			fmt.Printf("f") //fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
		}
		x = safest.X
		y = safest.Y
		fmt.Printf("%v @ (%v, %v)\n", safest.MineProb, x, y)
		if safest.MineProb > 0.0 {
			// give a visual indication of our lack of confidence
			c.HesitateAround(x, y)
		}
	}

	time.Sleep(5 * time.Second)
}
