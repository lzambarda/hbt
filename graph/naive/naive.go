// Package naive contains a simplistic implementation of a suggestion graph.
// Nothing too fancy.
//
//nolint:godox // Okay for now.
package naive

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/lzambarda/hbt/internal"
)

//nolint:govet // Prefer this order of memory efficiency.
type edge struct {
	Hits int   `json:"c"`
	From *node `json:"f"`
	To   *node `json:"t"`
}

// cmd -> node.
//
//nolint:govet // Prefer this order of memory efficiency.
type node struct {
	id    int
	edges map[string]*edge
}

type cmdEdge struct {
	cmd   string
	score int
}

func (c *cmdEdge) String() string {
	return fmt.Sprintf("{ cmd: %q, score: %d }", c.cmd, c.score)
}

func (n *node) getSortedEdges() []*cmdEdge {
	sorted := make([]*cmdEdge, 0, len(n.edges))
	for cmd, e := range n.edges {
		sorted = append(sorted, &cmdEdge{cmd, e.Hits})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})
	return sorted
}

func (n *node) getBestCommand() string {
	max := -1
	var best string
	for cmd, e := range n.edges {
		if e.Hits > max {
			max = e.Hits
			best = cmd
		}
	}
	return best
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

// Graph is a naive implementation of a heuristic system.
// The zero value of this structure cannot be used. Please use NewGraph to
// obtain a valid one.
//
//nolint:govet // Prefer this order of memory efficiency.
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
	// For each session, keep an internal counter to cycle through the possible
	// suggestions.
	suggestionState map[string]int
}

// NewGraph returns usable Graph instances.
// If minCommonPath is set to a value <=0, it will default to 1.
func NewGraph(maxWalkerHistory, minCommonPath int) *Graph {
	if minCommonPath <= 0 {
		minCommonPath = 1
	}
	return &Graph{
		Nodes:            map[string]*node{},
		MaxWalkerHistory: maxWalkerHistory,
		MinCommonPath:    minCommonPath,
		walkers:          map[string]walker{},
		suggestionState:  map[string]int{},
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

// Track adds to the graph the command cmd performed at path wd by the id
// user/process.
func (g *Graph) Track(id, wd, cmd string) {
	// Check if this is a new session we are creating
	walker := g.walkers[id]
	if walker == nil {
		walker = make([]*walkerNode, 0, g.MaxWalkerHistory)
	}
	// Reset the suggestion state
	g.suggestionState[id] = 0
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

func (g *Graph) findNode(wd string) *node {
	if n, ok := g.Nodes[wd]; ok {
		return n
	}
	// Try to see if we have a node with a similar structure
	wd = strings.TrimPrefix(wd, "/")
	pathComponents := strings.Split(wd, "/")
	if len(pathComponents) > g.MinCommonPath {
		// Reduce the path to the common path and check again
		return g.findNode("/" + path.Join(pathComponents[len(pathComponents)-g.MinCommonPath:]...))
	}
	// Maybe even check the walker's history
	return nil
}

// Hint returns the next suggestion for user/process id at path wd.
func (g *Graph) Hint(id, wd string) string {
	n := g.findNode(wd)
	if n == nil {
		// Reset suggestion for session
		g.suggestionState[id] = 0
		return shrug
	}
	// TODO: maybe we could use the walker to get the next action???
	// walker := g.walkers[id]
	// if walker == nil {
	// 	return shrug
	// }
	// Use the suggestion state to cycle through the commands
	if len(n.edges) == 0 {
		// Reset suggestion for session
		g.suggestionState[id] = 0
		return shrug
	}
	sorted := n.getSortedEdges()
	bestIndex := g.suggestionState[id] % len(n.edges)
	if internal.Debug {
		fmt.Println("Sorted Edges:")
		for i, s := range sorted {
			fmt.Printf("  [%d] %s\n", i, s)
		}
		fmt.Printf("Best index %d %% %d = %d\n", g.suggestionState[id], len(n.edges), bestIndex)
	}
	best := sorted[bestIndex].cmd
	if best == "" {
		// Reset suggestion for session
		g.suggestionState[id] = 0
		return shrug
	}
	g.suggestionState[id]++
	return best
}

// Delete removes a previously tracked command. It should not return an
// error.
func (g *Graph) Delete(id, wd, cmd string) {
	n := g.findNode(wd)
	if n == nil {
		return
	}
	if _, ok := n.edges[cmd]; !ok {
		return
	}
	delete(n.edges, cmd)
	// Deleting an edge invalidates the suggestion offset, better to reset it
	// here.
	g.suggestionState[id] = 0
}

// End clears a session for user/process id. This is useful to reset a
// stateful graph.
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

// Save serialises the graph to the given file path.
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
	return os.WriteFile(filePath, b, os.ModePerm)
}

// Load initialises the graph with a serialiastion at the give file path.
func (g *Graph) Load(filePath string) error {
	b, err := os.ReadFile(filePath) //nolint:gosec // It is okay.
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
