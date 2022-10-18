package server

// Graph has all the functions a suggestion graph needs to be implemented.
type Graph interface {
	// Track adds to the graph the command cmd performed at path wd by the id
	// user/process.
	Track(id, wd, cmd string)
	// Hint returns the next suggestion for user/process id at path wd.
	Hint(id, wd string) string
	// End clears a session for user/process id. This is useful to reset a
	// stateful graph.
	End(id string)
	// Delete removes a previously tracked command. It should not return an
	// error.
	Delete(id, wd, cmd string)
	// Save serialises the graph to the given file path.
	Save(filePath string) error
	// Load initialises the graph with a serialiastion at the give file path.
	Load(filePath string) error
	// Prune removes unused edges. Should return whether any edge has been
	// removed.
	// If the graph implementation has no prune strategy, it should return
	// false,nil.
	Prune() (removed bool, err error)
}
