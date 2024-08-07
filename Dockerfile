FROM --platform=linux/amd64 golang:1.22-bookworm

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      build-essential \
      make \
      libsqlite3-dev \
      sshpass \
      openssh-client


WORKDIR /app
COPY . .
RUN go generate ./...
RUN  --mount=type=cache,target="/.cache/go" \
    GOCACHE=/.cache/go/go \
    GOMODCACHE=/.cache/go/mod \
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main_linux ./cmd/main.go

ARG SSHPASS
ARG REMOTE_USER
ARG REMOTE_HOST
RUN make deploy