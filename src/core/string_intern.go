package core

import "sync"

var (
	internMu sync.RWMutex
	nameToID = make(map[string]int)
	idToName = make([]string, 0)
)

var ADD = InternName("add")
var INIT = InternName("init")
var NEXT = InternName("__next__")
var ITER = InternName("__iter__")
var TO_STRING = InternName("toString")
var MSG = InternName("msg")
var X = InternName("x")
var Y = InternName("y")
var Z = InternName("z")
var W = InternName("w")
var R = InternName("r")
var G = InternName("g")
var B = InternName("b")
var A = InternName("a")

// InternName takes a string and returns an integer ID for it. Safe to call
// concurrently from multiple VM instances (see thread module) -- the
// common case (name already interned) only takes a read lock.
func InternName(name string) int {
	internMu.RLock()
	if id, ok := nameToID[name]; ok {
		internMu.RUnlock()
		return id
	}
	internMu.RUnlock()

	internMu.Lock()
	defer internMu.Unlock()
	// Another goroutine may have interned this name while we waited for
	// the write lock -- check again before allocating a new id.
	if id, ok := nameToID[name]; ok {
		return id
	}
	id := len(idToName)
	nameToID[name] = id
	idToName = append(idToName, name)
	return id
}

func NameFromID(id int) string {
	internMu.RLock()
	defer internMu.RUnlock()
	return idToName[id]
}
