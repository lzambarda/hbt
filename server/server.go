// Package server contains all logic related to the server.
package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lzambarda/hbt/internal"
)

func saveRoutines(g Graph, cachePath string) {
	// Intercept termination signal to save the most recent knowledge
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if internal.Debug {
			fmt.Println("Saving graph at", cachePath)
		}
		err := g.Save(cachePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	// Periodically save the file
	go func() {
		for {
			time.Sleep(internal.SaveInterval)
			if internal.Debug {
				fmt.Println("Saving graph at", cachePath)
			}
			err := g.Save(cachePath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}()
}

// Start the hbt server with the given graph and cache path.
func Start(g Graph, cachePath string) error {
	saveRoutines(g, cachePath)
	if internal.Debug {
		fmt.Println("Starting server at", internal.Port)
	}
	l, err := net.Listen("tcp4", ":"+internal.Port)
	if err != nil {
		return err
	}
	defer l.Close() //nolint:errcheck // It is okay.
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConnection(c, g)
	}
}

// Stop the server.
func Stop() error {
	return nil
}

func handleConnection(c net.Conn, g Graph) {
	defer c.Close()              //nolint:errcheck // It is okay.
	buf := make([]byte, 0, 4096) // arbitrary
	tmp := make([]byte, 64)      // arbitrary
	for {
		n, err := c.Read(tmp)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Println(err)
				return
			}
			break
		}
		buf = append(buf, tmp[:n]...)
	}
	args := strings.Split(string(buf), "\n")
	if internal.Debug {
		fmt.Println(args)
	}
	switch args[0] {
	case "track":
		if len(args) != 4 {
			fmt.Printf("wrong number of arguments, expected 4, got %d\n", len(args))
			return
		}
		g.Track(args[1], args[2], args[3])
	case "hint":
		if len(args) != 4 {
			fmt.Printf("wrong number of arguments, expected 4, got %d\n", len(args))
			return
		}
		// args[3] is unused for now
		c.Write([]byte(g.Hint(args[1], args[2]))) //nolint:errcheck,gosec // It is okay.
	case "end":
		if len(args) != 2 {
			fmt.Printf("wrong number of arguments, expected 2, got %d\n", len(args))
			return
		}
		g.End(args[1])
	case "del":
		if len(args) != 4 {
			fmt.Printf("wrong number of arguments, expected 4, got %d\n", len(args))
			return
		}
		g.Delete(args[1], args[2], args[3])
	default:
		println("unknown command: " + args[0])
	}
}
