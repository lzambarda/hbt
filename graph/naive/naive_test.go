//nolint:goconst // Prefer readability.
package naive

import (
	"encoding/json"
	"os"
	"path"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNaive(t *testing.T) {
	t.Run("Node", testNaiveNode)
	t.Run("Track", testNaiveTrack)
	t.Run("Hint", testNaiveHint)
	t.Run("Save", testNaiveSave)
	t.Run("Load", testNaiveLoad)
	t.Run("Delete", testNaiveDelete)
}

func testNaiveNode(t *testing.T) {
	cmd1 := "c1"
	cmd2 := "c2"
	cmd3 := "c3"
	n := node{
		edges: map[string]*edge{},
	}
	n.edges[cmd1] = &edge{Hits: 2}
	n.edges[cmd2] = &edge{Hits: 3}
	n.edges[cmd3] = &edge{Hits: 1}
	assert.Equal(t, cmd2, n.getBestCommand())
	e := n.getSortedEdges()
	assert.Len(t, e, 3)
	assert.EqualValues(t, cmd2, e[0].cmd)
	assert.EqualValues(t, cmd1, e[1].cmd)
	assert.EqualValues(t, cmd3, e[2].cmd)
}

func DummyNow() time.Time {
	return time.Unix(482196050, 0).UTC() // 1985-04-12T23:20:50Z
}

func testNaiveTrack(t *testing.T) {
	g := NewGraph(10, 3, DummyNow)
	id := "1"
	wd := "d1"
	cmd1 := "c1"
	g.Track(id, wd, cmd1)
	assert.Len(t, g.Nodes, 1)
	assert.Contains(t, g.Nodes, wd)
	assert.Contains(t, g.Nodes[wd].edges, cmd1)
	assert.Equal(t, g.Nodes[wd].edges[cmd1].Hits, 1)
	assert.Equal(t, g.Nodes[wd].edges[cmd1].From, g.Nodes[wd])
	assert.Nil(t, g.Nodes[wd].edges[cmd1].To)

	assert.Len(t, g.walkers, 1)
	assert.Contains(t, g.walkers, id)

	assert.Len(t, g.walkers[id], 1)
	assert.Equal(t, g.walkers[id][0].lastNode, g.Nodes[wd])

	cmd2 := "c2"
	g.Track(id, wd, cmd2)
	assert.Len(t, g.Nodes, 1)
	assert.Contains(t, g.Nodes, wd)
	assert.Contains(t, g.Nodes[wd].edges, cmd1)
	assert.Contains(t, g.Nodes[wd].edges, cmd2)
	assert.Equal(t, g.Nodes[wd].edges[cmd1].Hits, 1)
	assert.Equal(t, g.Nodes[wd].edges[cmd1].From, g.Nodes[wd])
	// Now this has been updated
	assert.Equal(t, g.Nodes[wd].edges[cmd1].To, g.Nodes[wd])

	assert.Equal(t, g.Nodes[wd].edges[cmd2].Hits, 1)
	assert.Equal(t, g.Nodes[wd].edges[cmd2].From, g.Nodes[wd])
	assert.Nil(t, g.Nodes[wd].edges[cmd2].To)

	assert.Len(t, g.walkers, 1)
	assert.Contains(t, g.walkers, id)

	assert.Len(t, g.walkers[id], 2)

	// Run to only increment Hits on graph
	g.Track(id, wd, cmd2)
	assert.Len(t, g.Nodes, 1)
	assert.Equal(t, g.Nodes[wd].edges[cmd1].Hits, 1)
	// Should increase the count
	assert.Equal(t, g.Nodes[wd].edges[cmd2].Hits, 2)

	assert.Len(t, g.walkers[id], 3)
	assert.Equal(t, g.walkers[id][0], g.walkers[id][1])
}

func testNaiveHint(t *testing.T) {
	t.Run("Base", testNaiveHintBasic)
	t.Run("Breakdown", testNaiveHintBreakdown)
}

func testNaiveHintBasic(t *testing.T) {
	g := NewGraph(10, 3, DummyNow)
	id := "1"
	wd1 := "d1"
	got := g.Hint(id, wd1)
	assert.Equal(t, shrug, got, "no matching node, no tracking info")

	got = g.Hint(id, "d1/d2/d3/d4") // shold be longer than minCommonPath
	assert.Equal(t, shrug, got, "no matching node, no tracking info, long path")

	cmd1 := "c1"
	g.Track(id, wd1, cmd1)
	got = g.Hint(id, wd1)
	assert.Equal(t, cmd1, got, "single matching node")

	id2 := "2"
	got = g.Hint(id2, wd1)
	assert.Equal(t, cmd1, got, "single matching node, for another provider")

	wd2 := "d2"
	got = g.Hint(id, wd2)
	assert.Equal(t, shrug, got, "no matching node, tracking info")

	cmd2 := "c2"
	g.Track(id, wd1, cmd2)
	got = g.Hint(id, wd1)
	assert.Contains(t, []string{cmd1, cmd2}, got, "two commands, same hits, random result")

	g.Track(id, wd1, cmd2)
	got = g.Hint(id, wd1)
	assert.Equal(t, cmd2, got, "two commands, cmd2 greater hits")
	got = g.Hint(id, wd1)
	assert.Equal(t, cmd1, got, "two commands, cmd2 greater hits, cycle hints 1")
	got = g.Hint(id, wd1)
	assert.Equal(t, cmd2, got, "two commands, cmd2 greater hits, cycle hints 2")
}

func testNaiveHintBreakdown(t *testing.T) {
	g := NewGraph(10, 3, DummyNow)
	id := "1"
	wd1 := "/foo/bar/baz"
	cmd1 := "binary arg1 arg2 --flag1 flag1value -t"
	g.Track(id, wd1, cmd1)
	got := g.Hint(id, wd1)
	assert.Equal(t, cmd1, got, "same dir match")

	wd2 := "/another/foo/bar/baz"
	got = g.Hint(id, wd2)
	assert.Equal(t, cmd1, got, "different dir match")
}

func testNaiveSave(t *testing.T) {
	runs := map[string]func(g *Graph){
		"simple": func(g *Graph) {
			g.Track("1", "dir1", "cmd1")
		},
		"cyclic_single": func(g *Graph) {
			g.Track("1", "dir1", "cmd1")
			g.Track("1", "dir1", "cmd2")
		},
		"cyclic_multi": func(g *Graph) {
			g.Track("1", "dir1", "cmd1")
			g.Track("1", "dir2", "cmd2")
			g.Track("1", "dir1", "cmd3")
		},
	}
	for name, setup := range runs {
		t.Run(name, func(t *testing.T) {
			g := NewGraph(10, 3, DummyNow)
			setup(g)
			base := path.Join("testdata", name)
			actualFile := base + "_actual.json"
			err := g.Save(actualFile)
			assert.NoError(t, err)
			be, err := os.ReadFile(base + ".json")
			require.NoError(t, err)
			var expected map[string]interface{}
			err = json.Unmarshal(be, &expected)
			require.NoError(t, err)
			ba, err := os.ReadFile(actualFile)
			require.NoError(t, err)
			var actual map[string]interface{}
			err = json.Unmarshal(ba, &actual)
			require.NoError(t, err)
			if assert.EqualValues(t, expected, actual) {
				err = os.Remove(actualFile)
				require.NoError(t, err)
			}
		})
	}
}

func testNaiveLoad(t *testing.T) {
	runs := map[string]func(g *Graph){
		"simple": func(g *Graph) {
			g.Track("1", "dir1", "cmd1")
		},
		// "cyclic_single": func(g *Graph) {
		// 	g.Track("1", "dir1", "cmd1")
		// 	g.Track("1", "dir1", "cmd2")
		// },
		// "cyclic_multi": func(g *Graph) {
		// 	g.Track("1", "dir1", "cmd1")
		// 	g.Track("1", "dir2", "cmd2")
		// 	g.Track("1", "dir1", "cmd3")
		// },
	}
	for name, setup := range runs {
		t.Run(name, func(t *testing.T) {
			expected := NewGraph(10, 3, DummyNow)
			setup(expected)
			actual := NewGraph(10, 3, DummyNow)
			err := actual.Load(path.Join("testdata", name+".json"))
			assert.NoError(t, err)
			if diff := cmp.Diff(expected, actual, cmp.AllowUnexported(Graph{}, node{}, edge{}), cmpopts.IgnoreFields(Graph{}, "nowFunc", "walkers", "suggestionState")); diff != "" {
				t.Errorf("(-expected +actual):\n%s", diff)
			}
		})
	}
}

func testNaiveDelete(t *testing.T) {
	g := NewGraph(10, 3, DummyNow)
	g.Track("123", "abc", "def")
	assert.EqualValues(t, "def", g.Hint("123", "abc"))
	g.Delete("123", "xxx", "def")
	assert.EqualValues(t, "def", g.Hint("123", "abc"))
	g.Delete("123", "abc", "xxx")
	assert.EqualValues(t, "def", g.Hint("123", "abc"))
	g.Delete("123", "abc", "def")
	assert.EqualValues(t, shrug, g.Hint("123", "abc"))
}
