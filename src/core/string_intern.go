package core

var (
	nameToID = make(map[string]int)
	idToName = make([]string, 0)
)

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

// InternName takes a string and returns an integer ID for it.
func InternName(name string) int {
	if id, ok := nameToID[name]; ok {
		return id
	}
	id := len(idToName)
	nameToID[name] = id
	idToName = append(idToName, name)
	return id
}

func NameFromID(id int) string {
	return idToName[id]
}
