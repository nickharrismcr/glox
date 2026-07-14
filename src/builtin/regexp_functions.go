package builtin

import (
	"glox/src/core"
	"regexp"
	"sync"
)

// patternCache caches compiled *regexp.Regexp by their literal pattern text
// (including the \A / \A...\z wrapped variants used for match/fullmatch),
// so repeated calls with the same pattern string don't recompile every time.
var patternCache sync.Map // map[string]*regexp.Regexp

func compileCached(pattern string) (*regexp.Regexp, error) {
	if v, ok := patternCache.Load(pattern); ok {
		return v.(*regexp.Regexp), nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	patternCache.Store(pattern, re)
	return re, nil
}

func compileAnchoredStart(pattern string) (*regexp.Regexp, error) {
	return compileCached(`\A(?:` + pattern + `)`)
}

func compileFullAnchored(pattern string) (*regexp.Regexp, error) {
	return compileCached(`\A(?:` + pattern + `)\z`)
}

func matchFromRegex(re *regexp.Regexp, s string) core.Value {
	loc := re.FindStringSubmatchIndex(s)
	if loc == nil {
		return core.NIL_VALUE
	}
	names := re.SubexpNames()
	match := MakeRegexMatchObject(s, loc, names)
	return core.MakeObjectValue(match, false)
}

// subWithCount replaces up to count matches (count<=0 means all) of re in s with repl
// (Go's $1 / ${name} replacement syntax), returning the result and the number of replacements made.
func subWithCount(re *regexp.Regexp, repl string, s string, count int) (string, int) {
	n := -1
	if count > 0 {
		n = count
	}
	matches := re.FindAllStringSubmatchIndex(s, n)
	if len(matches) == 0 {
		return s, 0
	}
	var buf []byte
	last := 0
	for _, m := range matches {
		buf = append(buf, s[last:m[0]]...)
		buf = re.ExpandString(buf, repl, s, m)
		last = m[1]
	}
	buf = append(buf, s[last:]...)
	return string(buf), len(matches)
}

func splitWithMax(re *regexp.Regexp, s string, maxsplit int) []string {
	n := -1
	if maxsplit > 0 {
		n = maxsplit + 1
	}
	return re.Split(s, n)
}

// findallResults mirrors Python's re.findall: no groups -> whole-match strings,
// one group -> that group's strings, 2+ groups -> tuples of the groups.
func findallResults(re *regexp.Regexp, s string) core.Value {
	numGroups := re.NumSubexp()
	all := re.FindAllStringSubmatch(s, -1)
	items := make([]core.Value, 0, len(all))
	for _, m := range all {
		switch numGroups {
		case 0:
			items = append(items, core.MakeStringObjectValue(m[0], false))
		case 1:
			items = append(items, core.MakeStringObjectValue(m[1], false))
		default:
			tupleItems := make([]core.Value, 0, numGroups)
			for i := 1; i <= numGroups; i++ {
				tupleItems = append(tupleItems, core.MakeStringObjectValue(m[i], false))
			}
			items = append(items, core.MakeObjectValue(core.MakeListObject(tupleItems, true), false))
		}
	}
	return core.MakeObjectValue(core.MakeListObject(items, false), false)
}

func argAsPattern(vm core.VMContext, v core.Value, what string) (string, bool) {
	if !v.IsStringObject() {
		vm.RunTimeError("%s must be a string.", what)
		return "", false
	}
	return v.AsString().Get(), true
}

func RegexSearchBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to re.search.")
		return core.NIL_VALUE
	}
	pattern, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "re.search pattern")
	if !ok {
		return core.NIL_VALUE
	}
	s, ok := argAsPattern(vm, vm.Stack(arg_stackptr+1), "re.search string")
	if !ok {
		return core.NIL_VALUE
	}
	re, err := compileCached(pattern)
	if err != nil {
		vm.RunTimeError("re: invalid pattern: %v", err)
		return core.NIL_VALUE
	}
	return matchFromRegex(re, s)
}

func RegexMatchBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to re.match.")
		return core.NIL_VALUE
	}
	pattern, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "re.match pattern")
	if !ok {
		return core.NIL_VALUE
	}
	s, ok := argAsPattern(vm, vm.Stack(arg_stackptr+1), "re.match string")
	if !ok {
		return core.NIL_VALUE
	}
	re, err := compileAnchoredStart(pattern)
	if err != nil {
		vm.RunTimeError("re: invalid pattern: %v", err)
		return core.NIL_VALUE
	}
	return matchFromRegex(re, s)
}

func RegexFullmatchBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to re.fullmatch.")
		return core.NIL_VALUE
	}
	pattern, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "re.fullmatch pattern")
	if !ok {
		return core.NIL_VALUE
	}
	s, ok := argAsPattern(vm, vm.Stack(arg_stackptr+1), "re.fullmatch string")
	if !ok {
		return core.NIL_VALUE
	}
	re, err := compileFullAnchored(pattern)
	if err != nil {
		vm.RunTimeError("re: invalid pattern: %v", err)
		return core.NIL_VALUE
	}
	return matchFromRegex(re, s)
}

func RegexSubBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 3 && argCount != 4 {
		vm.RunTimeError("re.sub expects 3 or 4 arguments.")
		return core.NIL_VALUE
	}
	pattern, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "re.sub pattern")
	if !ok {
		return core.NIL_VALUE
	}
	repl, ok := argAsPattern(vm, vm.Stack(arg_stackptr+1), "re.sub repl")
	if !ok {
		return core.NIL_VALUE
	}
	s, ok := argAsPattern(vm, vm.Stack(arg_stackptr+2), "re.sub string")
	if !ok {
		return core.NIL_VALUE
	}
	count := 0
	if argCount == 4 {
		cv := vm.Stack(arg_stackptr + 3)
		if !cv.IsInt() {
			vm.RunTimeError("re.sub count must be an integer.")
			return core.NIL_VALUE
		}
		count = cv.AsInt()
	}
	re, err := compileCached(pattern)
	if err != nil {
		vm.RunTimeError("re: invalid pattern: %v", err)
		return core.NIL_VALUE
	}
	result, _ := subWithCount(re, repl, s, count)
	return core.MakeStringObjectValue(result, false)
}

func RegexSubnBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 3 && argCount != 4 {
		vm.RunTimeError("re.subn expects 3 or 4 arguments.")
		return core.NIL_VALUE
	}
	pattern, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "re.subn pattern")
	if !ok {
		return core.NIL_VALUE
	}
	repl, ok := argAsPattern(vm, vm.Stack(arg_stackptr+1), "re.subn repl")
	if !ok {
		return core.NIL_VALUE
	}
	s, ok := argAsPattern(vm, vm.Stack(arg_stackptr+2), "re.subn string")
	if !ok {
		return core.NIL_VALUE
	}
	count := 0
	if argCount == 4 {
		cv := vm.Stack(arg_stackptr + 3)
		if !cv.IsInt() {
			vm.RunTimeError("re.subn count must be an integer.")
			return core.NIL_VALUE
		}
		count = cv.AsInt()
	}
	re, err := compileCached(pattern)
	if err != nil {
		vm.RunTimeError("re: invalid pattern: %v", err)
		return core.NIL_VALUE
	}
	result, n := subWithCount(re, repl, s, count)
	items := []core.Value{
		core.MakeStringObjectValue(result, false),
		core.MakeIntValue(n, false),
	}
	return core.MakeObjectValue(core.MakeListObject(items, true), false)
}

func RegexSplitBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 2 && argCount != 3 {
		vm.RunTimeError("re.split expects 2 or 3 arguments.")
		return core.NIL_VALUE
	}
	pattern, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "re.split pattern")
	if !ok {
		return core.NIL_VALUE
	}
	s, ok := argAsPattern(vm, vm.Stack(arg_stackptr+1), "re.split string")
	if !ok {
		return core.NIL_VALUE
	}
	maxsplit := 0
	if argCount == 3 {
		mv := vm.Stack(arg_stackptr + 2)
		if !mv.IsInt() {
			vm.RunTimeError("re.split maxsplit must be an integer.")
			return core.NIL_VALUE
		}
		maxsplit = mv.AsInt()
	}
	re, err := compileCached(pattern)
	if err != nil {
		vm.RunTimeError("re: invalid pattern: %v", err)
		return core.NIL_VALUE
	}
	parts := splitWithMax(re, s, maxsplit)
	items := make([]core.Value, 0, len(parts))
	for _, p := range parts {
		items = append(items, core.MakeStringObjectValue(p, false))
	}
	return core.MakeObjectValue(core.MakeListObject(items, false), false)
}

func RegexFindallBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to re.findall.")
		return core.NIL_VALUE
	}
	pattern, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "re.findall pattern")
	if !ok {
		return core.NIL_VALUE
	}
	s, ok := argAsPattern(vm, vm.Stack(arg_stackptr+1), "re.findall string")
	if !ok {
		return core.NIL_VALUE
	}
	re, err := compileCached(pattern)
	if err != nil {
		vm.RunTimeError("re: invalid pattern: %v", err)
		return core.NIL_VALUE
	}
	return findallResults(re, s)
}

func RegexCompileBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to re.compile.")
		return core.NIL_VALUE
	}
	pattern, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "re.compile pattern")
	if !ok {
		return core.NIL_VALUE
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		vm.RunTimeError("re: invalid pattern: %v", err)
		return core.NIL_VALUE
	}
	patternObj := MakeRegexPatternObject(re, pattern)
	return core.MakeObjectValue(patternObj, false)
}
