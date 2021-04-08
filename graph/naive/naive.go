package naive

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type edge struct {
	Hits int  `json:"c"`
	From node `json:"f"`
	To   node `json:"t"`
}

// cmd -> node
type node map[string]*edge

type walkerNode struct {
	lastNode node
	lastEdge *edge
}

type walker []*walkerNode

func (w walker) progress(next *walkerNode) walker {
	if len(w) == 0 {
		return append(w, next)
	}
	w[0].lastEdge.To = next.lastNode
	max := len(w)
	if len(w) == cap(w) {
		max = cap(w)
	}
	return append([]*walkerNode{next}, w[:max]...)
}

type Graph struct {
	// wd -> node
	// Must assess how efficient this implementation is.
	Nodes            map[string]node   `json:"nodes"`
	MaxWalkerHistory int               `json:"max_walker_history"`
	walkers          map[string]walker // not saved to file
}

func NewGraph(maxWalkerHistory int) *Graph {
	return &Graph{
		Nodes:            map[string]node{},
		MaxWalkerHistory: maxWalkerHistory,
		walkers:          map[string]walker{},
	}
}

func (g *Graph) newNode(wd, cmd string, parent node) (node, *edge) {
	n := node{}
	e := &edge{
		Hits: 1,
		From: n,
		To:   parent,
	}
	n[cmd] = e
	g.Nodes[wd] = n
	return n, e
}

func (g *Graph) Track(id, wd, cmd string) {
	// Check if this is a new session we are creating
	walker := g.walkers[id]
	if walker == nil {
		walker = make([]*walkerNode, 0, g.MaxWalkerHistory)
	}
	// Check if there is a node matching the current wdectory
	if _, ok := g.Nodes[wd]; !ok {
		// TODO: for now don't do anything, but we should try a cmd hook
		n, e := g.newNode(wd, cmd, nil)
		g.walkers[id] = walker.progress(&walkerNode{
			lastNode: n,
			lastEdge: e,
		})
		return
	}
	n := g.Nodes[wd]
	// Now check if there is a known edge with the run command
	if _, ok := n[cmd]; !ok {
		// TODO: should really check permutations of the command (maybe even
		// just the binary name)
		e := &edge{
			Hits: 1,
			From: n,
			To:   nil,
		}
		n[cmd] = e
		g.walkers[id] = walker.progress(&walkerNode{
			lastNode: n,
			lastEdge: e,
		})
		return
	}
	n[cmd].Hits++
	// Reference to itself
	g.walkers[id] = walker.progress(walker[0])
}

const shrug = "¯\\_(ツ)_/¯"

func (g *Graph) Hint(id, wd string) string {
	if _, ok := g.Nodes[wd]; !ok {
		// TODO: for now don't do anything, but we should try fuzzy matching by
		// scanning the wd components
		// Maybe even check the walker's history
		return shrug
	}
	n := g.Nodes[wd]
	// TODO: maybe we could use the walker to get the next action???
	// walker := g.walkers[id]
	// if walker == nil {
	// 	return shrug
	// }
	var max = -1
	var best string
	for cmd, e := range n {
		if e.Hits > max {
			max = e.Hits
			best = cmd
		}
	}
	if best == "" {
		return shrug
	}
	return best
}

func (g *Graph) End(id string) {
	delete(g.walkers, id)
}

func (g *Graph) Load(filePath string) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Nothing to load
			return nil
		}
		return err
	}
	return json.Unmarshal(b, g)
}

func (g *Graph) Save(filePath string) error {
	b, err := json.Marshal(g)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, b, os.ModePerm)
}
