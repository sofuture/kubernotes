PROGRAM = kubernotes
PREFIX  = bin

default: build test

build:
	@CGO_ENABLED=0 godep go build -a -installsuffix cgo -o $(PREFIX)/$(PROGRAM)
	@cp $(PREFIX)/$(PROGRAM) $(GOPATH)/$(PREFIX)/$(PROGRAM) || true

test:
	@godep go test ./... -cover -parallel=4 -race --timeout=300s

clean:
	@rm -rf bin/*

.PHONY: default build test clean
