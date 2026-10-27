.PHONY: test build

test:
	go test ./...

build:
	go build -o adsb-matrix-display ./cmd/server
