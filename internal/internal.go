package internal

const (
	DebugName     = "HBT_DEBUG"
	CachePathName = "HBT_CACHE"
	PortName      = "HBT_PORT"
)

const (
	// Found by looking at unused ports at:
	// https://en.wikipedia.org/wiki/List_of_TCP_and_UDP_port_numbers
	DefaultPort      = "43111"
	DefaultCachePath = "."
)

var (
	Debug     bool
	CachePath string
	Port      string
	// Must be var, otherwise -X flag can't modify it
	Version = "unknown"
)

const (
	CacheName = ".hbtcache"
)
