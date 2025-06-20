package lox

import (
	"bytes"
	bin "encoding/binary"
	"fmt"
	"glox/src/core"

	"glox/src/util"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//functions for cacheing and retrieval of compiled bytecode in .lxc files

func writeToLxc(vm *VM, serialised *bytes.Buffer) {

	dir := filepath.Dir(vm.script)

	// Create the cache subdirectory
	cacheDir := filepath.Join(dir, "__loxcache__")
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		panic(fmt.Errorf("failed to create cache dir: %w", err))
	}

	// Remove .lox extension from filename
	base := filepath.Base(vm.script)
	name := strings.TrimSuffix(base, ".lox")
	cacheFile := filepath.Join(cacheDir, name+".lxc")

	// Write to file
	err = os.WriteFile(cacheFile, serialised.Bytes(), 0644)
	if err != nil {
		panic(fmt.Errorf("failed to write cache file: %w", err))
	}
}

func loadLxc(scriptPath string) (*core.Chunk, *core.Environment, bool) {

	if core.ForceModuleCompile {
		return nil, nil, false
	}
	core.LogFmt(core.INFO, "Attempting to load lxc for %s", scriptPath)

	// Determine cache path
	dir := filepath.Dir(scriptPath)
	cacheDir := filepath.Join(dir, "__loxcache__")
	base := strings.TrimSuffix(filepath.Base(scriptPath), ".lox")
	cachePath := filepath.Join(cacheDir, base+".lxc")

	// Check timestamps
	sourceInfo, err := os.Stat(scriptPath)
	if err != nil {
		//Debug("lxc not found.")
		return nil, nil, false
	}

	cacheInfo, err := os.Stat(cachePath)
	useCache := err == nil && cacheInfo.ModTime().After(sourceInfo.ModTime())

	if useCache {
		// Load from cache
		reader, err := os.Open(cachePath)
		if err != nil {
			//Debug("lxc not found.")
			return nil, nil, false
		}
		core.LogFmt(core.INFO, "loading lxc. %s", base)
		env := core.NewEnvironment(base)
		chunk := readChunk(reader, env)
		return chunk, env, true
	}
	return nil, nil, false
}

func readChunk(reader io.Reader, env *core.Environment) *core.Chunk {

	var codeLen uint32
	util.ReadMarker(reader)
	bin.Read(reader, bin.LittleEndian, &codeLen)
	//Debugf("Code len %d", codeLen)
	code := make([]byte, codeLen)
	io.ReadFull(reader, code)
	util.ReadMarker(reader)

	var lineCount uint32
	bin.Read(reader, bin.LittleEndian, &lineCount)
	//Debugf("Line count %d", lineCount)
	lines := make([]int, lineCount)
	for i := range lines {
		var l uint32
		bin.Read(reader, bin.LittleEndian, &l)
		lines[i] = int(l)
	}
	util.ReadMarker(reader)

	var constCount uint32
	bin.Read(reader, bin.LittleEndian, &constCount)
	//Debugf("Const count %d", constCount)
	constants := make([]core.Value, constCount)
	for i := range constants {
		constants[i] = readValue(reader, env)
	}
	var filenameLen uint32
	util.ReadMarker(reader)
	bin.Read(reader, bin.LittleEndian, &filenameLen)
	//Debugf("Filename len %d", filenameLen)
	filename := make([]byte, filenameLen)
	reader.Read(filename)
	//Debugf("String %s ", string(filename))
	util.ReadMarker(reader)
	chunk := core.MakeChunk(string(filename), code, constants, lines)
	return chunk
}

func readValue(r io.Reader, env *core.Environment) core.Value {
	var tag [1]byte
	r.Read(tag[:])
	//Debugf("Tag : %d", tag[0])
	switch tag[0] {
	case 0x01:
		var n float64
		bin.Read(r, bin.LittleEndian, &n)
		//Debugf("Float %f", n)
		return core.MakeFloatValue(n, false)
	case 0x02:
		var n uint32
		bin.Read(r, bin.LittleEndian, &n)
		//Debugf("Int %d", n)
		return core.MakeIntValue(int(n), false)
	case 0x03:
		var len uint32
		bin.Read(r, bin.LittleEndian, &len)
		buf := make([]byte, len)
		r.Read(buf)
		//Debugf("String %s ", string(buf))
		return core.MakeStringObjectValue(string(buf), false)
	case 0x04:
		s := util.ReadString(r)
		name := core.MakeStringObject(s)
		//Debugf("Function %s", s)
		var arity uint32
		bin.Read(r, bin.LittleEndian, &arity)
		//Debugf("Arity %d", arity)
		var upvalueCount uint32
		bin.Read(r, bin.LittleEndian, &upvalueCount)
		//Debugf("Arity %d", arity)
		chunk := readChunk(r, env)
		fo := core.MakeFunctionObject(name.Get(), env)
		fo.Name = name
		//Debugf("Function %s arity %d upvalueCount %d", fo.name.get(), arity, upvalueCount)
		fo.Arity = int(arity)
		fo.UpvalueCount = int(upvalueCount)
		fo.Chunk = chunk
		return core.MakeObjectValue(fo, false)
	case 0x05:
		var b [1]byte
		r.Read(b[:])
		//Debugf("Bool %s", b[0] == 1)
		return core.MakeBooleanValue(b[0] == 1, false)
	case 0x06:
		//Debugf("Nil")
		return core.NIL_VALUE
	default:
		panic("unknown tag")
	}
}
