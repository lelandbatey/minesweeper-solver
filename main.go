package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/y0ssar1an/q"
)

// DefuseDivision -- How does the client server model work?
//
// The client sends messages over a tcp socket. The boundary between messages
// is a null byte. Each message is gzipped before it's sent over the wire.
//
// As the client, we'll need to open a tcp socket then wait for an initial
// message telling us about ourselves. From there on out, the server will
// respond with messages outlined here:
// https://github.com/lelandbatey/defuse_division/blob/master/architecture.md

type Client struct {
	Connection net.Conn
	Name       string
	Msgs       chan []byte
}

func NetReader(client *Client) {
	//
	buf := []byte{}
	tmp := make([]byte, 50)
	for {
		n, err := client.Connection.Read(tmp)
		if n == 0 || err != nil {
			q.Q(err)
			break
		}
		// We definitely have at least one message, now to process that one
		// message.
		if bytes.Contains(tmp[:n], []byte{'\x00', '\x01', '\x00'}) {
			inpt := []byte{}
			inpt = append(inpt, buf...)
			inpt = append(inpt, tmp[:n]...)
			msgs := bytes.Split(inpt, []byte{'\x00', '\x01', '\x00'})

			q.Q(len(msgs))
			for _, m := range msgs[:len(msgs)-1] {
				gzr, err := gzip.NewReader(bytes.NewReader(m))
				if err != nil {
					q.Q(err)
					err = nil
					continue
				}
				ungzipped, err := ioutil.ReadAll(gzr)
				if err != nil {
					q.Q(err)
					err = nil
					continue
				}
				client.Msgs <- ungzipped
			}
		} else {
			buf = append(buf, tmp[:n]...)
		}

	}
}

func NewClient(host string, port string) (*Client, error) {
	addr := fmt.Sprintf("%s:%s", host, port)
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	rv := Client{
		Connection: c,
		Name:       "example",
		Msgs:       make(chan []byte),
	}
	return &rv, nil
}

// Method Message will block until it returns the next message from the server.
// That return value is an interface of either type Player, type State, or a
// slice of interfaces (an 'update-selected' message).
func (c *Client) Message() interface{} {
	msg := <-c.Msgs
	//fmt.Printf("%s\n", msg[:30])

	q.Q(string(msg))
	// The server sent a 'configuration' message, a Player object about ourselves
	var player Player
	var err error
	if err = json.Unmarshal(msg, &player); err == nil {
		return player
	}
	//panic(err)
	// Implicitly, there was an error, because the message was not a player
	// object. So now, we test the other two cases, both of which are lists.
	var items []interface{}
	err = json.Unmarshal(msg, &items)
	if err != nil {
		// We don't know what the heck the structure of this message could be
		// if it's not a list of stuff!
		panic(err)
	}
	if first, ok := items[0].(string); ok {
		if first == "new-state" {
			thebits, err := json.Marshal(items[1])
			if err != nil {
				panic("BUT WE JUST GOT THIS FROM JSON, HOW CNA WE NOT MARSHAL IT AGAIN?!?!")
			}
			var state State
			json.Unmarshal(thebits, &state)
			return state
		} else if first == "update-selected" {
			return items[1]
		} else {
			panic("HOW COULD IT NOT BE ONE OF THOSE TWO!?!")
		}
	} else {
		panic("WHAT ELSE CAN THE FIRST BE IF NOT A STRING?!?!?!")
	}
}

func main() {
	//fmt.Println("vim-go")
	//fmt.Println("asdfsdf")

	host := os.Args[1]
	port := os.Args[2]
	client, err := NewClient(host, port)
	go NetReader(client)
	if err != nil {
		panic(err)
	}
	fmt.Printf("We did it, we opened a client named %s!\n", client.Name)
	q.Q("Now we wait for a message:")
	spew.Dump(client.Message())
}
