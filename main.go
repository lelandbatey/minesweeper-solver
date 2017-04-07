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
	fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
	c.MoveRight()
	c.MoveRight()
	c.MoveRight()
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

	for _, unflaggedCell := range solver.UnflaggedMines(sboard) {
		c.MoveToCell(unflaggedCell)
		time.Sleep(400 * time.Millisecond)
		c.Send("FLAG")
		time.Sleep(400 * time.Millisecond)
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
