package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

type cacheEdge struct {
	Count int        `json:"c"`
	To    *cacheNode `json:"t"`
}

// cmd -> node
type cacheNode map[string]*cacheEdge

// dir -> node
var cache = map[string]*cacheNode{}

const cacheName = ".hbtcache"

const maxCacheWalkerSize = 10

var cacheWalker = make([]*cacheNode, 0, maxCacheWalkerSize)

func loadCache(cachePath string) error {
	b, err := ioutil.ReadFile(path.Join(cachePath, cacheName))
	if err != nil {
		if os.IsNotExist(err) {
			// Nothing to load
			return nil
		}
		return err
	}
	return json.Unmarshal(b, &cache)
}

func saveCache(cachePath string) error {
	b, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(cachePath, cacheName), b, os.ModePerm)
}

func main() {
	arguments := os.Args
	if len(arguments) != 3 {
		println("usage: hbt <port> <cache_path>")
		os.Exit(1)
	}
	err := loadCache(arguments[2])
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
		println("\r- Ctrl+C pressed in Terminal")
		err = saveCache(arguments[2])
		if err != nil {
			println(err.Error())
		}
		os.Exit(0)
	}()
	port := ":" + arguments[1]
	l, err := net.Listen("tcp4", port)
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
	if len(args) != 3 {
		println("needs 3 args")
		return
	}
	switch args[0] {
	case "track":
		track(args[1], args[2])
	case "hint":
		c.Write([]byte(hint(args[1], args[2])))
	default:
		println("unknown: " + args[0])
	}
}

func walkTo(n *cacheNode) {
	cacheWalker = append(
		cacheWalker[maxCacheWalkerSize-1:maxCacheWalkerSize],
		cacheWalker[0:maxCacheWalkerSize-1]...)
}

func track(dir, cmd string) {
	if n, ok := cache[dir]; !ok {
		nn = &cacheNode{}
		cache[dir] = nn
		walkTo(nn)
	} else e, ok := n[cmd]; !ok {
		ne = &cacheEdge{}
		n[cmd] = ne
		walkTo(nn)
	}
	if cd, ok := cache[dir]; !ok {
		cache[dir] = map[string]*cacheNode{
			count: 1,
		}
		cache[dir][cmd] = 1
	} else if _, ok := cd[cmd]; !ok {
		cd[cmd] = 1
	} else {
		cd[cmd]++
	}
	println("track " + dir + " " + cmd)
}

func hint(dir, cmd string) string {
	// if cd, ok := cache[dir]; ok {
	// 	if cdc, ok := cd[cmd]; ok {

	// 	}
	// }
	println("hint " + dir + " " + cmd)
	return "hinting"
}
