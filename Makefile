run:
	go run main.go
build:
	go build -ldflags "-w -s" -o kuchak main.go
fmt:
	go fmt .

