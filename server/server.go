package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lzambarda/hbt/graph"
	"github.com/lzambarda/hbt/internal"
)

func Start(g graph.Graph, cachePath string) error {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := g.Save(cachePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	if internal.Debug {
		fmt.Println("Starting server at", internal.Port)
	}
	l, err := net.Listen("tcp4", ":"+internal.Port)
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConnection(c, g)
	}
}

func Stop() error {
	return nil
}

func handleConnection(c net.Conn, g graph.Graph) {
	defer c.Close()
	buf := make([]byte, 0, 4096) // arbitrary
	tmp := make([]byte, 64)      // arbitrary
	for {
		n, err := c.Read(tmp)
		if err != nil {
			if err != io.EOF {
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
		c.Write([]byte(g.Hint(args[1], args[2])))
	case "end":
		if len(args) != 2 {
			fmt.Printf("wrong number of arguments, expected 2, got %d\n", len(args))
			return
		}
		g.End(args[1])
	default:
		println("unknown command: " + args[0])
	}
}
