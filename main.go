package main

import (
	"io"
	"net"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/lzambarda/hbt/graph/naive"
)

const cacheName = ".hbtcache"
const usage = "usage: hbt <port> <cache_path> [--verbose]"

var graph = naive.NewGraph(10)
var verbose = false

func main() {
	arguments := os.Args
	if len(arguments) < 3 {
		println(usage)
		os.Exit(1)
	}
	if len(arguments) == 4 {
		if arguments[3] != "--verbose" {
			println(usage)
			os.Exit(1)
		}
		verbose = true
		println("starting hbt in verbose mode")
	}
	err := graph.Load(path.Join(arguments[2], cacheName))
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err = graph.Save(path.Join(arguments[2], cacheName))
		if err != nil {
			println(err.Error())
		}
		os.Exit(0)
	}()
	l, err := net.Listen("tcp4", ":"+arguments[1])
	if err != nil {
		println(err.Error())
		exitCode = 1
		return
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			println(err.Error())
			exitCode = 1
			return
		}
		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()
	buf := make([]byte, 0, 4096) // arbitrary
	tmp := make([]byte, 64)      // arbitrary
	for {
		n, err := c.Read(tmp)
		if err != nil {
			if err != io.EOF {
				println(err.Error())
				return
			}
			break
		}
		buf = append(buf, tmp[:n]...)
	}
	args := strings.Split(string(buf), "\n")
	if len(args) != 4 {
		println("needs 4 args")
		return
	}
	shellID := args[0]
	method := args[1]
	shellWd := args[2]
	shellCmd := args[3]
	if verbose {
		println(shellID, "\t", method, "\t", shellWd, "\t", shellCmd)
	}
	switch method {
	case "track":
		graph.Track(shellID, shellWd, shellCmd)
	case "hint":
		c.Write([]byte(graph.Hint(shellID, shellWd)))
	default:
		println("unknown: " + method)
	}
}
