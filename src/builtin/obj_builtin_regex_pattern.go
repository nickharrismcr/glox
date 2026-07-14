package builtin

import (
	"fmt"
	"glox/src/core"
	"regexp"
)

// RegexPatternObject is the reusable compiled-pattern object returned by re.compile().
type RegexPatternObject struct {
	core.BuiltInObject
	Source        string
	Re            *regexp.Regexp
	startAnchored *regexp.Regexp // lazily built \A(?:Source)
	fullAnchored  *regexp.Regexp // lazily built \A(?:Source)\z
	Methods       map[int]*core.BuiltInObject
}

func MakeRegexPatternObject(re *regexp.Regexp, source string) *RegexPatternObject {
	o := &RegexPatternObject{
		Source: source,
		Re:     re,
	}
	RegisterAllRegexPatternMethods(o)
	return o
}

func (o *RegexPatternObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}

func (o *RegexPatternObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *RegexPatternObject) StartAnchored() (*regexp.Regexp, error) {
	if o.startAnchored == nil {
		re, err := regexp.Compile(`\A(?:` + o.Source + `)`)
		if err != nil {
			return nil, err
		}
		o.startAnchored = re
	}
	return o.startAnchored, nil
}

func (o *RegexPatternObject) FullAnchored() (*regexp.Regexp, error) {
	if o.fullAnchored == nil {
		re, err := regexp.Compile(`\A(?:` + o.Source + `)\z`)
		if err != nil {
			return nil, err
		}
		o.fullAnchored = re
	}
	return o.fullAnchored, nil
}

func (o *RegexPatternObject) String() string {
	return fmt.Sprintf("<Pattern '%s'>", o.Source)
}

func (o *RegexPatternObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *RegexPatternObject) IsBuiltIn() bool {
	return true
}
