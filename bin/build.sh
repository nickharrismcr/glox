set -x

go build -o bin/glox main.go
cp bin/glox bin/glox.exe
