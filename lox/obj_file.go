package lox

import (
	"bufio"
	"os"
	"strings"
)

type FileObject struct {
	file   *os.File
	closed bool
	eof    bool
	reader *bufio.Reader
	writer *bufio.Writer
}

func makeFileObject(file *os.File) *FileObject {

	return &FileObject{
		file:   file,
		reader: bufio.NewReader(file),
		writer: bufio.NewWriter(file),
	}
}

func (FileObject) isObject() {}

func (FileObject) getType() ObjectType {

	return OBJECT_FILE
}

func (f *FileObject) String() string {

	return "<file>"
}

func (f *FileObject) close() {
	f.writer.Flush()
	f.file.Close()
	f.closed = true
}

func (f *FileObject) readLine() Value {

	if f.eof {
		return makeNilValue()
	}

	line, err := f.reader.ReadString('\n')
	line = strings.TrimRight(line, "\r\n")
	if err != nil {
		if err.Error() == "EOF" {
			f.eof = true
			if len(line) > 0 {
				return makeObjectValue(makeStringObject(line), false)
			}
		}
	}
	return makeObjectValue(makeStringObject(line), false)
}

func (f *FileObject) write(str Value) {

	s := str.asString()
	s = strings.ReplaceAll(s, `\n`, "\n")
	f.writer.WriteString(s)

}

//-------------------------------------------------------------------------------------------

//-------------------------------------------------------------------------------------------

//-------------------------------------------------------------------------------------------
