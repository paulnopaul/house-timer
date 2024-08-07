# Rule for building the Go binary
.PHONY: build
build: ./cmd/main.go generate
	go mod tidy
	go build -o main ./cmd/main.go

.PHONY: build_linux
build_linux: ./cmd/main.go generate
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main_linux ./cmd/main.go

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

# Rule to clean your project by removikjng the binary and object files
clean:
	rm -f main
	rm -rf ./cmd/zz.generated_prod_migrations
	rm -rf ./internal/pkg/usecases/tasks/zz.generated_test_migrations

.PHONY: deploy
deploy: build_linux
	 sshpass -e scp -r ./main_linux  ${REMOTE_USER}@${REMOTE_HOST}:~/house-timer
