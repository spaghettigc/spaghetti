default: build run
build:
	go build -o ./bin/main ./cmd/spaghetti/main.go
run:
	./bin/main
test:
	gotest ./...

install:
	go mod tidy