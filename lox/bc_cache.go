package lox

import (
	"bytes"
	bin "encoding/binary"
	"fmt"
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

func loadLxc(scriptPath string) (*Chunk, bool) {

	if ForceModuleCompile {
		return nil, false
	}
	Debugf("Attempting to load lxc for %s", scriptPath)

	// Determine cache path
	dir := filepath.Dir(scriptPath)
	cacheDir := filepath.Join(dir, "__loxcache__")
	base := strings.TrimSuffix(filepath.Base(scriptPath), ".lox")
	cachePath := filepath.Join(cacheDir, base+".lxc")

	// Check timestamps
	sourceInfo, err := os.Stat(scriptPath)
	if err != nil {
		Debug("lxc not found.")
		return nil, false
	}

	cacheInfo, err := os.Stat(cachePath)
	useCache := err == nil && cacheInfo.ModTime().After(sourceInfo.ModTime())

	if useCache {
		// Load from cache
		reader, err := os.Open(cachePath)
		if err != nil {
			Debug("lxc not found.")
			return nil, false
		}
		Debug("loading lxc.")
		chunk := readChunk(reader)
		return chunk, true
	}
	return nil, false
}

func (c *Chunk) serialise(b *bytes.Buffer) {

	writeMarker(b)
	bin.Write(b, bin.LittleEndian, uint32(len(c.code)))
	b.Write(c.code)
	writeMarker(b)
	bin.Write(b, bin.LittleEndian, uint32(len(c.lines)))
	for _, line := range c.lines {
		bin.Write(b, bin.LittleEndian, uint32(line))
	}
	writeMarker(b)
	bin.Write(b, bin.LittleEndian, uint32(len(c.constants)))
	for _, v := range c.constants {
		v.serialise(b)
	}
	writeMarker(b)
}

func (v *Value) serialise(buffer *bytes.Buffer) {

	switch v.Type {
	case VAL_FLOAT:
		buffer.Write([]byte{0x01})
		bin.Write(buffer, bin.LittleEndian, v.Float)
	case VAL_INT:
		buffer.Write([]byte{0x02})
		bin.Write(buffer, bin.LittleEndian, uint32(v.Int))
	case VAL_OBJ:
		switch v.Obj.getType() {
		case OBJECT_STRING:
			buffer.Write([]byte{0x03})
			s := v.asString().get()
			bin.Write(buffer, bin.LittleEndian, uint32(len(s)))
			buffer.Write([]byte(s))

		case OBJECT_FUNCTION:
			fo := v.asFunction()
			buffer.Write([]byte{0x04})
			writeString(buffer, fo.name.get())
			bin.Write(buffer, bin.LittleEndian, uint32(fo.arity))
			bin.Write(buffer, bin.LittleEndian, uint32(fo.upvalueCount))
			fo.chunk.serialise(buffer)
		default:
			panic("serialise object value not handled")
		}
	case VAL_BOOL:
		buffer.Write([]byte{0x05})
		b := byte(0)
		if v.Bool {
			b = byte(1)
		}
		buffer.Write([]byte{b})
	case VAL_NIL:
		buffer.Write([]byte{0x06})
	default:
		panic("serialise value not handled")
	}
}

func writeString(b *bytes.Buffer, s string) {
	bin.Write(b, bin.LittleEndian, uint32(len(s)))
	b.Write([]byte(s))
}

func readChunk(reader io.Reader) *Chunk {

	var codeLen uint32
	readMarker(reader)
	bin.Read(reader, bin.LittleEndian, &codeLen)
	Debugf("Code len %d", codeLen)
	code := make([]byte, codeLen)
	io.ReadFull(reader, code)
	readMarker(reader)

	var lineCount uint32
	bin.Read(reader, bin.LittleEndian, &lineCount)
	Debugf("Line count %d", lineCount)
	lines := make([]int, lineCount)
	for i := range lines {
		var l uint32
		bin.Read(reader, bin.LittleEndian, &l)
		lines[i] = int(l)
	}
	readMarker(reader)

	var constCount uint32
	bin.Read(reader, bin.LittleEndian, &constCount)
	Debugf("Const count %d", constCount)
	constants := make([]Value, constCount)
	for i := range constants {
		constants[i] = readValue(reader)
	}
	readMarker(reader)

	chunk := &Chunk{code: code, lines: lines, constants: constants}
	return chunk
}

func readValue(r io.Reader) Value {
	var tag [1]byte
	r.Read(tag[:])
	Debugf("Tag : %d", tag[0])
	switch tag[0] {
	case 0x01:
		var n float64
		bin.Read(r, bin.LittleEndian, &n)
		Debugf("Float %f", n)
		return makeFloatValue(n, false)
	case 0x02:
		var n uint32
		bin.Read(r, bin.LittleEndian, &n)
		Debugf("Int %d", n)
		return makeIntValue(int(n), false)
	case 0x03:
		var len uint32
		bin.Read(r, bin.LittleEndian, &len)
		buf := make([]byte, len)
		r.Read(buf)
		Debugf("String %s ", string(buf))
		return makeObjectValue(makeStringObject(string(buf)), false)
	case 0x04:
		s := readString(r)
		name := makeStringObject(s)
		Debugf("Function %s", s)
		var arity uint32
		bin.Read(r, bin.LittleEndian, &arity)
		Debugf("Arity %d", arity)
		var upvalueCount uint32
		bin.Read(r, bin.LittleEndian, &upvalueCount)
		Debugf("Arity %d", arity)
		chunk := readChunk(r)
		return makeObjectValue(&FunctionObject{name: name, arity: int(arity), upvalueCount: int(upvalueCount), chunk: chunk}, false)
	case 0x05:
		var b [1]byte
		r.Read(b[:])
		Debugf("Bool %s", b[0] == 1)
		return makeBooleanValue(b[0] == 1, false)
	case 0x06:
		Debugf("Nil")
		return makeNilValue()
	default:
		panic("unknown tag")
	}
}

func writeMarker(b *bytes.Buffer) {
	b.Write([]byte{0xFF})
}

func readMarker(r io.Reader) {
	buf := make([]byte, 1)
	r.Read(buf)
	if buf[0] != 0xFF {
		panic(fmt.Sprintf("Expected marker, got %d", buf[0]))
	}
}

func readString(r io.Reader) string {
	var len uint32
	bin.Read(r, bin.LittleEndian, &len)
	buf := make([]byte, len)
	r.Read(buf)
	return string(buf)
}
