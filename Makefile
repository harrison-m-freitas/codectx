BINARY=codectx

.PHONY: build test fmt vet run clean

build:
	go build -o $(BINARY) ./cmd/codectx

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

run: build
	./$(BINARY) -h

clean:
	rm -f $(BINARY)
