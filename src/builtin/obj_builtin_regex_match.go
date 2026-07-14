package builtin

import (
	"fmt"
	"glox/src/core"
)

// RegexMatchObject represents the result of a successful re.search/match/fullmatch call.
type RegexMatchObject struct {
	core.BuiltInObject
	Source  string
	Indices []int    // pairs [start0,end0, start1,end1, ...]; -1,-1 for a non-participating group
	Names   []string // group names, index-aligned with Indices/2; Names[0] == ""
	Methods map[int]*core.BuiltInObject
}

func MakeRegexMatchObject(source string, indices []int, names []string) *RegexMatchObject {
	o := &RegexMatchObject{
		Source:  source,
		Indices: indices,
		Names:   names,
	}
	RegisterAllRegexMatchMethods(o)
	return o
}

func (o *RegexMatchObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}

func (o *RegexMatchObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *RegexMatchObject) GroupCount() int {
	return len(o.Indices)/2 - 1
}

// GroupText returns the text of group n and whether it participated in the match.
func (o *RegexMatchObject) GroupText(n int) (string, bool) {
	if n < 0 || n > o.GroupCount() {
		return "", false
	}
	start, end := o.Indices[2*n], o.Indices[2*n+1]
	if start < 0 || end < 0 {
		return "", false
	}
	return o.Source[start:end], true
}

// GroupIndexByName resolves a named group to its index, or -1 if not found.
func (o *RegexMatchObject) GroupIndexByName(name string) int {
	for i, n := range o.Names {
		if n == name {
			return i
		}
	}
	return -1
}

func (o *RegexMatchObject) String() string {
	text, _ := o.GroupText(0)
	return fmt.Sprintf("<Match '%s'>", text)
}

func (o *RegexMatchObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *RegexMatchObject) IsBuiltIn() bool {
	return true
}
