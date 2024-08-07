# Rule for building the Go binary
.PHONY: build
build: ./cmd/main.go generate
	go mod tidy
	go build -o main ./cmd/main.go

.PHONY: build_linux
build_linux: ./cmd/main.go generate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main_linux ./cmd/main.go

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
deploy: 
	 sshpass -e ssh -o "StrictHostKeyChecking=no" ${REMOTE_USER}@${REMOTE_HOST} "rm ~/house-timer"
	 sshpass -e scp -o "StrictHostKeyChecking=no" ./main_linux  ${REMOTE_USER}@${REMOTE_HOST}:~/house-timer
	 sshpass -e ssh -o "StrictHostKeyChecking=no" ${REMOTE_USER}@${REMOTE_HOST} "systemctl restart house-timer.service"

.PHONY: backup
backup: 
	sshpass -e scp ${REMOTE_USER}@${REMOTE_HOST}:~/tasks.sqlite ./tasks.sqlite

.PHONY: backup
debug_remote: 
	sshpass -e ssh ${REMOTE_USER}@${REMOTE_HOST}


.PHONY: deploy_docker
deploy_docker: 
	docker build -t house-timer --build-arg SSHPASS=${SSHPASS} --build-arg REMOTE_USER=${REMOTE_USER} --build-arg REMOTE_HOST=${REMOTE_HOST} . 