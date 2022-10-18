//nolint:godox // Okay for now.
// Package naive contains a simplistic implementation of a suggestion graph.
// Nothing too fancy.
package naive

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/lzambarda/hbt/internal"
)

type NowFunc func() time.Time

func TimeNow() time.Time {
	return time.Now().UTC()
}

//nolint:govet // Prefer this order of memory efficiency.
type edge struct {
	Hits          int       `json:"c"`
	From          *node     `json:"f"`
	To            *node     `json:"t"`
	lastVisitedAt time.Time `json:"v"`
}

//nolint:govet // Prefer this order of memory efficiency.
type node struct {
	id int
	// cmd -> node
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
	var sorted = make([]*cmdEdge, 0, len(n.edges))
	for cmd, e := range n.edges {
		sorted = append(sorted, &cmdEdge{cmd, e.Hits})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})
	return sorted
}

func (n *node) getBestCommand() string {
	var max = -1
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
//nolint:govet // Prefer this order of memory efficiency.
type Graph struct {
	// wd -> node
	// NOTE: Must assess how efficient this implementation is.
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
	// Used to get the current time. For tests only really
	nowFunc NowFunc
}

// NewGraph returns usable Graph instances.
// If minCommonPath is set to a value <=0, it will default to 1.
func NewGraph(maxWalkerHistory, minCommonPath int, nowFunc NowFunc) *Graph {
	if minCommonPath <= 0 {
		minCommonPath = 1
	}
	if nowFunc == nil {
		nowFunc = TimeNow
	}
	return &Graph{
		Nodes:            map[string]*node{},
		MaxWalkerHistory: maxWalkerHistory,
		MinCommonPath:    minCommonPath,
		walkers:          map[string]walker{},
		suggestionState:  map[string]int{},
		nowFunc:          nowFunc,
	}
}

func (g *Graph) newNode(wd, cmd string, parent *node) (*node, *edge) {
	n := &node{
		id:    len(g.Nodes), // this will eventually break
		edges: map[string]*edge{},
	}
	e := &edge{
		Hits:          1,
		From:          n,
		To:            parent,
		lastVisitedAt: g.nowFunc(),
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
			Hits:          1,
			From:          n,
			To:            nil,
			lastVisitedAt: g.nowFunc(),
		}
		n.edges[cmd] = e
		g.walkers[id] = walker.progress(&walkerNode{
			lastNode: n,
			lastEdge: e,
		})
		return
	}
	n.edges[cmd].Hits++
	n.edges[cmd].lastVisitedAt = g.nowFunc()
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
	Hits          int   `json:"h"`
	To            int   `json:"t"`
	LastVisitedAt int64 `json:"v"`
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
	index := 0
	for wd, n := range g.Nodes {
		nodes[index] = n
		sg.Wds[index] = wd
		index++
	}
	// Second pass, do the same with edges
	for fromIndex, n := range nodes {
		sg.Edges[fromIndex] = map[string]serialisableEdge{}
		for cmd, e := range n.edges {
			se := serialisableEdge{
				Hits:          e.Hits,
				To:            -1,
				LastVisitedAt: e.lastVisitedAt.Unix(),
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
	idToNode := map[int]*node{}
	// First pass, lay down all node pointers
	for id, wd := range sg.Wds {
		n := &node{
			id:    id,
			edges: map[string]*edge{},
		}
		g.Nodes[wd] = n
		idToNode[id] = n
	}
	// Second pass, do the same with edges
	for nodeID, edges := range sg.Edges {
		n := g.Nodes[sg.Wds[nodeID]]
		for cmd, se := range edges {
			e := &edge{
				Hits:          se.Hits,
				From:          n,
				To:            nil,
				lastVisitedAt: time.Unix(se.LastVisitedAt, 0),
			}
			// NOTE: Before v0.2.0 there was no lastVisitedAt value
			if e.lastVisitedAt.Unix() == 0 {
				e.lastVisitedAt = g.nowFunc()
			} else {
				e.lastVisitedAt = e.lastVisitedAt.UTC()
			}
			if se.To != -1 {
				e.To = idToNode[se.To]
			}
			n.edges[cmd] = e
		}
	}
	return nil
}

// Prune removes unused edges. Should return whether any edge has been
// removed.
// If the graph implementation has no prune strategy, it should return
// false,nil.
func (g *Graph) Prune() (removed bool, _ error) {
	// Delete edges older than 6 months with a score of 1.
	const maxThreshold = time.Hour * 6 * 30 * 24
	var cdRegexp = regexp.MustCompile(`^cd($|\s[^|\s]+$)`)
	removedEdges := 0
	removedNodes := 0
	for wd, n := range g.Nodes {
		for cmd, e := range n.edges {
			// Also remove anything that is a cd (for older implementations)
			if time.Since(e.lastVisitedAt) >= maxThreshold || cdRegexp.MatchString(cmd) {
				removedEdges++
				delete(n.edges, cmd)
			}
		}
		if len(n.edges) == 0 {
			removedNodes++
			delete(g.Nodes, wd)
		}
	}
	fmt.Printf("Removed edges: %d\nRemoved nodes: %d\n", removedEdges, removedNodes)
	return removedNodes > 0, nil
}
