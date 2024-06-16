
# Rule for building the Go binary
.PHONY: main
main: ./cmd/main.go generate
	go build -o $@ ./cmd/main.go

# Rule for running the executable
run: main
	./main

.PHONY: generate
generate: 
	go generate ./...

.PHONY: test
test: generate 
	go test ./...

# Rule to clean your project by removing the binary and object files
clean:
	rm main
	rm $(ag -g "zz.generated.*k")