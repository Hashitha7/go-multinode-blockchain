.PHONY: build test test-race run cluster clean

# Build the node binary
build:
	go build -o node ./cmd/node/

# Run all tests
test:
	go test ./...

# Run all tests with race detector
test-race:
	go test -race ./...

# Run a single node on port 3001
run: build
	./node --addr localhost:3001 --difficulty 4 --mine

# Start a 3-node cluster (Linux/Mac)
cluster: build
	bash ./scripts/cluster.sh

# Start a 3-node cluster (Windows)
cluster-win: build
	powershell -ExecutionPolicy Bypass -File ./scripts/cluster.ps1

# Format code
fmt:
	gofmt -w .

# Vet code
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -f node node.exe
	rm -f node1.log node2.log node3.log
	rm -f node1_err.log node2_err.log node3_err.log
