//nolint:revive,stylecheck // Self-explanatory
package internal

import "time"

const (
	DebugName        = "HBT_DEBUG"
	CachePathName    = "HBT_CACHE_PATH"
	PortName         = "HBT_PORT"
	SaveIntervalName = "HBT_SAVE_INTERVAL"
)

const (
	// Found by looking at unused ports at:
	// https://en.wikipedia.org/wiki/List_of_TCP_and_UDP_port_numbers
	DefaultPort         = "43111"
	DefaultCachePath    = "."
	DefaultSaveInterval = time.Minute * 10
)

var (
	Debug        bool
	CachePath    string
	Port         string
	SaveInterval time.Duration
	// Must be var, otherwise -X flag can't modify it.
	Version = "unknown"
)

const (
	CacheName = ".hbtcache"
)
