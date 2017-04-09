// Package client provides a client struct which can communicate with a
// DefuseDivision server.
package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"reflect"
	"time"

	"github.com/lelandbatey/minesweeper-solver/defusedivision"
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

// A Client struct holds a connection and some basic information about a player
// that this connection represents.
type Client struct {
	Connection net.Conn
	// Name will be "example" when initially created, but after the first
	// message which is a Player struct is read, the name of this Client struct
	// will match the name of the Player.
	Name string
	Msgs chan []byte
	X    int
	Y    int
}

func New(host string, port string) (*Client, error) {
	addr := fmt.Sprintf("%s:%s", host, port)
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	rv := Client{
		Connection: c,
		Name:       "example",
		Msgs:       make(chan []byte),
		X:          0,
		Y:          0,
	}
	return &rv, nil
}

// NetReader accepts a Client and reads byte-sequence-delimited gzipped messages from
// that Clients Connection, placing the uncompressed contents of each message
// into the "Msgs" channel on the provided Client struct. It will repeat this
// process, doing this forever until an error occurs/the Connection closes.
func NetReader(client *Client) {
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
		//if bytes.Contains(tmp[:n], []byte{'\x00', '\x01', '\x00'}) {
		// possibleBuf is buffer of everything received after last msg
		possibleBuf := []byte{}
		possibleBuf = append(possibleBuf, buf...)
		possibleBuf = append(possibleBuf, tmp[:n]...)
		if bytes.Contains(possibleBuf, []byte{'\x00', '\x01', '\x00'}) {
			msgs := bytes.Split(possibleBuf, []byte{'\x00', '\x01', '\x00'})

			// Set buf as the beginning of the as-yet incomplete message, which
			// may be an empty slice
			buf = msgs[len(msgs)-1]
			for _, m := range msgs[:len(msgs)-1] {
				gzr, err := gzip.NewReader(bytes.NewReader(m))
				if err != nil {
					panic(err)
					err = nil
					continue
				}
				err = gzr.Close() // close gzr to fully read until EOF
				if err != nil {
					panic(err)
				}
				ungzipped, err := ioutil.ReadAll(gzr)
				if err != nil {
					panic(err)
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

// move* functions send a command to client, then consume the
// corresponding confirmation message
func (c *Client) MoveUp() error {
	err := c.Send("UP")
	c.Y = c.Y - 1
	fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
	return err
}
func (c *Client) MoveDown() error {
	err := c.Send("DOWN")
	c.Y = c.Y + 1
	fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
	return err
}
func (c *Client) MoveLeft() error {
	err := c.Send("LEFT")
	c.X = c.X - 1
	fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
	return err
}
func (c *Client) MoveRight() error {
	err := c.Send("RIGHT")
	c.X = c.X + 1
	fmt.Printf("%v\n", reflect.TypeOf(c.Message()))
	return err
}

// WARNING. I'm using *solver.Cell. Almost all of the other client
// functions use defusedivision.* structs. I'd like this to accept a
// generic interface that supports both Cell types (since it only needs
// X and Y components). Is there a way to do that?
func (c *Client) MoveToXY(X int, Y int) error {
	// move up to cell
	for c.Y > Y {
		c.MoveUp()
	}
	// move down to cell
	for c.Y < Y {
		c.MoveDown()
	}
	// move left to cell
	for c.X > X {
		c.MoveLeft()
	}
	// move right to cell
	for c.X < X {
		c.MoveRight()
	}
	return nil
}

func (c *Client) Send(toSend interface{}) error {
	data, err := json.Marshal(toSend)
	if err != nil {
		return err
	}

	var packet []byte
	tmp := bytes.NewBuffer(packet)

	w := gzip.NewWriter(tmp)
	if _, err := w.Write(data); err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	_, err = tmp.Write([]byte{'\x00', '\x01', '\x00'})
	if err != nil {
		return err
	}
	_, err = tmp.WriteTo(c.Connection)
	if err != nil {
		return err
	}
	return nil
}

// Method Message will block until it returns the next message from the server.
// That return value is an interface of either type Player, type State, or a
// slice of interfaces (an 'update-selected' message). If a timeout occurs,
// returns nil.
func (c *Client) Message() interface{} {
	// Try to read a message for 50 ms, then time out
	var msg []byte
	select {
	case msg = <-c.Msgs:
	case <-time.After(500 * time.Millisecond):
		return nil
	}

	q.Q(string(msg))
	// Check if the server sent a 'configuration' message, a Player object
	// about ourselves.
	var player defusedivision.Player
	var err error
	if err = json.Unmarshal(msg, &player); err == nil {
		c.Name = player.Name
		return player
	}
	// Implicitly there was an error, because the message was not a player
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
			var state defusedivision.State
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
