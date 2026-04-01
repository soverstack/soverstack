VERSION := $(shell cat VERSION)
BINARY  := soverstack
LDFLAGS := -s -w -X main.Version=$(VERSION)

.PHONY: build test snapshot clean

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

test:
	go vet ./...
	go test ./...

snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -f $(BINARY)
	rm -rf dist/
