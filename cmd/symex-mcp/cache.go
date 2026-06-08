package main

import (
	"sync"
	"time"

	"indexmcp/symex/extractor"
)

type cacheEntry struct {
	mtime time.Time
	syms  []extractor.Symbol
}

var (
	cacheMu   sync.RWMutex
	fileCache = map[string]cacheEntry{}
)

func cacheGet(path string, mtime time.Time) ([]extractor.Symbol, bool) {
	cacheMu.RLock()
	entry, ok := fileCache[path]
	cacheMu.RUnlock()
	if ok && entry.mtime.Equal(mtime) {
		return entry.syms, true
	}
	return nil, false
}

func cachePut(path string, mtime time.Time, syms []extractor.Symbol) {
	cacheMu.Lock()
	fileCache[path] = cacheEntry{mtime: mtime, syms: syms}
	cacheMu.Unlock()
}
