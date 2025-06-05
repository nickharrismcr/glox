set -x

if [ $HOSTNAME = "THL-3LX7HS3" ]
then
	go build -o glox main.go
else
	go build -o glox.exe main.go
fi
