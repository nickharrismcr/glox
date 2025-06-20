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
