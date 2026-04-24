.PHONY: build install test clean

build:
	go build -o gnosis ./cmd/gnosis

install:
	go install ./cmd/gnosis

test:
	go test ./...

clean:
	rm -f gnosis
