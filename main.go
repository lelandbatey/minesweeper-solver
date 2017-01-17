package main

import (
	"fmt"
	"os"

	"github.com/lelandbatey/minesweeper-solver/client"

	"github.com/davecgh/go-spew/spew"
	"github.com/y0ssar1an/q"
)

func main() {

	host := os.Args[1]
	port := os.Args[2]
	c, err := client.New(host, port)
	go client.NetReader(c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("We did it, we opened a client named %s!\n", c.Name)
	q.Q("Now we wait for a message:")
	spew.Dump(c.Message())
}
