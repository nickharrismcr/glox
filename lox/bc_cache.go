package lox

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//functions for cacheing and retrieval of compiled bytecode in .lxc files

func writeToLxc(vm *VM, serialised string) {

	dir := filepath.Dir(vm.ScriptName)

	// Create the cache subdirectory
	cacheDir := filepath.Join(dir, "__loxcache__")
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		panic(fmt.Errorf("failed to create cache dir: %w", err))
	}

	// Remove .lox extension from filename
	base := filepath.Base(vm.ScriptName)
	name := strings.TrimSuffix(base, ".lox")
	cacheFile := filepath.Join(cacheDir, name+".lxc")

	// Write to file
	err = os.WriteFile(cacheFile, []byte(serialised), 0644)
	if err != nil {
		panic(fmt.Errorf("failed to write cache file: %w", err))
	}
}

func (c *Chunk) serialise() string {

	var b strings.Builder

	// Encode code bytes
	b.WriteString("CHUNK|")
	b.WriteString(hex.EncodeToString(c.code))
	b.WriteString("|")

	// Encode line numbers
	lineStrs := make([]string, len(c.lines))
	for i, l := range c.lines {
		lineStrs[i] = strconv.Itoa(l)
	}
	b.WriteString(strings.Join(lineStrs, ","))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(len(c.constants)))
	b.WriteString("\n")

	// Write constants
	for _, v := range c.constants {
		v.serialise(&b)
	}

	return b.String()
}

func (v *Value) serialise(b *strings.Builder) {

	switch v.Type {
	case VAL_FLOAT:
		fmt.Fprintf(b, "VAL|NUMBER|%f\n", v.Float)
	case VAL_INT:
		fmt.Fprintf(b, "VAL|NUMBER|%d\n", v.Int)
	case VAL_OBJ:
		switch v.Obj.getType() {
		case OBJECT_STRING:
			fmt.Fprintf(b, "VAL|STRING|%s\n", escape(v.Obj.String()))
		case OBJECT_FUNCTION:
			fo := v.Obj.(*FunctionObject)
			fmt.Fprintf(b, "VAL|FUNC|%s|%d\n", escape(fo.name.String()), fo.arity)
			b.WriteString(fo.chunk.serialise())
			b.WriteString("END_FUNC\n")
		}
	case VAL_BOOL:
		fmt.Fprintf(b, "VAL|BOOL|%v\n", v.Bool)
	case VAL_NIL:
		b.WriteString("VAL|NIL|\n")
	default:
		panic("serialise value not handled")
	}
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, `"`, "")
	return s
}
