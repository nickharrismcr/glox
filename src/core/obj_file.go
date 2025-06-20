package core

import (
	"bufio"
	"os"
	"strings"
)

type FileObject struct {
	File   *os.File
	Closed bool
	Eof    bool
	Reader *bufio.Reader
	Writer *bufio.Writer
}

func MakeFileObject(file *os.File) *FileObject {

	return &FileObject{
		File:   file,
		Reader: bufio.NewReader(file),
		Writer: bufio.NewWriter(file),
	}
}

func (FileObject) IsObject() {}

func (FileObject) GetType() ObjectType {

	return OBJECT_FILE
}

func (f *FileObject) String() string {

	return "<file>"
}

func (f *FileObject) Close() {
	f.Writer.Flush()
	f.File.Close()
	f.Closed = true
}

func (f *FileObject) ReadLine() Value {

	if f.Eof {
		return NIL_VALUE
	}

	line, err := f.Reader.ReadString('\n')
	line = strings.TrimRight(line, "\r\n")
	if err != nil {
		if err.Error() == "EOF" {
			f.Eof = true
			if len(line) > 0 {
				return MakeObjectValue(MakeStringObject(line), false)
			}
		}
	}
	return MakeObjectValue(MakeStringObject(line), false)
}

func (f *FileObject) Write(str Value) {

	s := str.AsString().Get()
	s = strings.ReplaceAll(s, `\n`, "\n")
	f.Writer.WriteString(s)

}

// -------------------------------------------------------------------------------------------
func (t FileObject) IsBuiltIn() bool {
	return true
}

//-------------------------------------------------------------------------------------------

//-------------------------------------------------------------------------------------------
