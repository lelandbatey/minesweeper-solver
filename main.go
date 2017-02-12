package main

import (
	"fmt"
	"os"
	"time"

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
	//spew.Dump(c.Message())
	c.Message()
	c.Send("RIGHT")
	time.Sleep(400 * time.Millisecond)
	//spew.Dump(c.Message())
	c.Send("RIGHT")
	time.Sleep(400 * time.Millisecond)
	c.Send("RIGHT")
	time.Sleep(400 * time.Millisecond)
	c.Send("RIGHT")
	time.Sleep(400 * time.Millisecond)
	fmt.Println("We are waiting for a message:")
	spew.Dump(c.Message())
	fmt.Println("We successfully read a message!")
	c.Send("PROBE")
	time.Sleep(1 * time.Second)
	spew.Dump(c.Message()) // don't exit just yet or else we won't see the changes
}
