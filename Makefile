# Rule for building the Go binary
.PHONY: build
build: ./cmd/main.go generate
	go build -o main ./cmd/main.go

# Rule for running the executable 
.PHONY: run
run: build generate
	./main

.PHONY: generate
generate: 
	go generate ./...

.PHONY: test
test: generate 
	go test ./...

# Rule to clean your project by removing the binary and object files
clean:
	rm -f main
	rm -rf ./cmd/zz.generated_prod_migrations
	rm -rf ./internal/pkg/usecases/tasks/zz.generated_test_migrations