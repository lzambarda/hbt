package naive

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type edge struct {
	Hits int   `json:"c"`
	From *node `json:"f"`
	To   *node `json:"t"`
}

// cmd -> node
type node struct {
	id    int
	edges map[string]*edge
}

type walkerNode struct {
	lastNode *node
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

// The zero value of this structure cannot be used. Please use NewGraph to
// obtain a valid one.
type Graph struct {
	// wd -> node
	// Must assess how efficient this implementation is.
	Nodes map[string]*node `json:"nodes"`
	// How many recent commands must each walker keep track of.
	MaxWalkerHistory int `json:"max_walker_history"`
	// How many path components (directories) are at least needed to be a match
	// of a different path.
	// This value should be a positive integer.
	MinCommonPath int               `json:"min_common_path"`
	walkers       map[string]walker // not saved to file
}

// Create a new graph with the given parameters.
// If minCommonPath is set to a value <=0, it will default to 1.
func NewGraph(maxWalkerHistory int, minCommonPath int) *Graph {
	if minCommonPath <= 0 {
		minCommonPath = 1
	}
	return &Graph{
		Nodes:            map[string]*node{},
		MaxWalkerHistory: maxWalkerHistory,
		MinCommonPath:    minCommonPath,
		walkers:          map[string]walker{},
	}
}

func (g *Graph) newNode(wd, cmd string, parent *node) (*node, *edge) {
	n := &node{
		id:    len(g.Nodes), // this will eventually break
		edges: map[string]*edge{},
	}
	e := &edge{
		Hits: 1,
		From: n,
		To:   parent,
	}
	n.edges[cmd] = e
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
	if _, ok := n.edges[cmd]; !ok {
		// TODO: should really check permutations of the command (maybe even
		// just the binary name)
		e := &edge{
			Hits: 1,
			From: n,
			To:   nil,
		}
		n.edges[cmd] = e
		g.walkers[id] = walker.progress(&walkerNode{
			lastNode: n,
			lastEdge: e,
		})
		return
	}
	n.edges[cmd].Hits++
	// Reference to itself
	if len(walker) == 0 {
		g.walkers[id] = walker.progress(&walkerNode{
			lastNode: n,
			lastEdge: n.edges[cmd],
		})
	} else {
		g.walkers[id] = walker.progress(walker[0])
	}
}

const shrug = "¯\\_(ツ)_/¯"

func (g *Graph) Hint(id, wd string) string {
	if _, ok := g.Nodes[wd]; !ok {
		// Try to see if we have a node with a similar structure
		if strings.HasPrefix(wd, "/") {
			wd = wd[1:]
		}
		pathComponents := strings.Split(wd, "/")
		if len(pathComponents) > g.MinCommonPath {
			// Reduce the path to the common path and check again
			return g.Hint(id, "/"+path.Join(pathComponents[len(pathComponents)-g.MinCommonPath:]...))

		}
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
	for cmd, e := range n.edges {
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

// These structs are what we need to move from a programmer-friendly structure
// to a serialisable one. Unfortunately pointers, cyclic references and
// serialisation do not play well together.
type serialisableGraph struct {
	Wds   []string                      `json:"wds"`
	Edges []map[string]serialisableEdge `json:"edges"`
}
type serialisableEdge struct {
	Hits int `json:"h"`
	To   int `json:"t"`
}

func (g *Graph) Save(filePath string) error {
	// We first need to build a model which doesn't contain pointers nor cycles.
	nodes := make([]*node, len(g.Nodes))
	sg := serialisableGraph{
		Wds:   make([]string, len(g.Nodes)),
		Edges: make([]map[string]serialisableEdge, len(g.Nodes)),
	}
	// First pass, get all node indices
	for wd, n := range g.Nodes {
		nodes[n.id] = n
		sg.Wds[n.id] = wd
	}
	// Second pass, do the same with edges
	for fromIndex, n := range nodes {
		sg.Edges[fromIndex] = map[string]serialisableEdge{}
		for cmd, e := range n.edges {
			se := serialisableEdge{
				Hits: e.Hits,
				To:   -1,
			}
			if e.To != nil {
				se.To = e.To.id
			}
			sg.Edges[fromIndex][cmd] = se
		}
	}
	b, err := json.Marshal(sg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, b, os.ModePerm)
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
	sg := serialisableGraph{}
	err = json.Unmarshal(b, &sg)
	if err != nil {
		return err
	}
	// Here we must do the opposite, where we start from the serialisable model
	// and build the programmer-friendly one.
	g.Nodes = map[string]*node{}
	// First pass, lay down all node pointers
	for id, wd := range sg.Wds {
		g.Nodes[wd] = &node{
			id:    id,
			edges: map[string]*edge{},
		}
	}
	// Second pass, do the same with edges
	for nodeID, edges := range sg.Edges {
		n := g.Nodes[sg.Wds[nodeID]]
		for cmd, se := range edges {
			e := &edge{
				Hits: se.Hits,
				From: n,
				To:   nil,
			}
			if se.To != -1 {
				e.To = g.Nodes[sg.Wds[se.To]]
			}
			n.edges[cmd] = e
		}
	}
	return nil
}
