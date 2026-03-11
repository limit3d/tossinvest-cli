BINARY := bin/tossctl

.PHONY: build run test fmt tidy clean

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/tossctl

run:
	go run ./cmd/tossctl

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

clean:
	rm -rf bin coverage.out

