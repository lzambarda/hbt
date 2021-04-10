package naive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNaive(t *testing.T) {
	t.Run("Track", TestNaiveTrack)
	t.Run("Hint", TestNaiveHint)
	t.Run("SaveLoad", TestNaiveSaveLoad)
}

func TestNaiveTrack(t *testing.T) {
	g := NewGraph(10)
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

func TestNaiveHint(t *testing.T) {
	g := NewGraph(10)
	id := "1"
	wd1 := "d1"
	got := g.Hint(id, wd1)
	assert.Equal(t, shrug, got, "no matching node, no tracking info")

	cmd1 := "c1"
	g.Track(id, wd1, cmd1)
	got = g.Hint(id, wd1)
	assert.Equal(t, cmd1, got, "single matching node")

	wd2 := "d2"
	got = g.Hint(id, wd2)
	assert.Equal(t, shrug, got, "no matching node, tracking info")

	cmd2 := "c2"
	g.Track(id, wd1, cmd2)
	got = g.Hint(id, wd1)
	assert.Equal(t, cmd1, got, "two commands, same hits")

	g.Track(id, wd1, cmd2)
	got = g.Hint(id, wd1)
	assert.Equal(t, cmd2, got, "two commands, cmd2 greater hits")
}

func TestNaiveSaveLoad(t *testing.T) {

}
