.PHONY: build install test clean

build:
	go build -o gn ./cmd/gn

install:
	go install ./cmd/gn

test:
	go test ./...

clean:
	rm -f gn
