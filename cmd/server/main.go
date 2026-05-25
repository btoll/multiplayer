package main

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"
)

var getColor = colorGenerator()

type client struct {
	id   int
	name string
	conn net.Conn
}

func getDivisibleBy2(n int) int {
	var i int
	for {
		i = rand.IntN(n)
		if i%2 == 0 {
			break
		}
	}
	return i
}

func handleConnection(conn net.Conn, game *game) {
	_, err := conn.Write([]byte("What is your name? \n"))
	if err != nil {
		log.Printf("err=%+v\n", err)
		err = conn.Close()
		if err != nil {
			log.Printf("err=%+v\n", err)
		}
		return
	}
	buf := make([]byte, 256)
	var name string
	var c *client
	for {
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			return
		}

		name = string(buf[:n-1])
		game.mu.Lock()
		_, found := names[name]
		game.mu.Unlock()

		if found {
			_, err := conn.Write([]byte("Name already exits, choose another. \n"))
			if err != nil {
				log.Printf("err=%+v\n", err)
			}
			continue
		}

		game.mu.Lock()
		colore, b := getColor()
		if !b {
			// TODO
			_, err := conn.Write([]byte("No more players allowed. \n"))
			if err != nil {
				log.Printf("err=%+v\n", err)
			}
			err = conn.Close()
			if err != nil {
				log.Printf("err=%+v\n", err)
			}
			return
		}
		c = &client{
			id:   int(colore),
			name: name,
			conn: conn,
		}
		names[name] = c
		game.mu.Unlock()
		break
	}
	_, err = fmt.Fprintf(conn, "Hi %s, welcome to the server!\n", name)
	if err != nil {
		log.Printf("err=%+v\n", err)
	}
	game.mu.Lock()
	game.register(c)
	game.mu.Unlock()
}

func shutdown(listener net.Listener, g *game) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	var err error
	g.mu.RLock()
	clients := make(map[int]*client, len(g.clients))
	maps.Copy(clients, g.clients)
	g.mu.RUnlock()

	for _, client := range clients {
		err = client.conn.Close()
		if err != nil {
			log.Printf("err=%+v\n", err)
		}
	}
	err = listener.Close()
	if err != nil {
		log.Printf("err=%+v\n", err)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":1111")
	if err != nil {
		panic(err)
	}

	done := make(chan struct{})
	visibleWidth, visibleHeight, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	game := newGame(visibleHeight, visibleWidth)

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				game.update()
				game.render()
			}
		}
	}()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					done <- struct{}{}
					return
				}
				continue
			}
			go handleConnection(conn, game)
		}
	}()

	shutdown(listener, game)
}
