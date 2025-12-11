GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
COMMIT=$(shell git describe --always)

# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	rm -f msgproto/*.pb.go && protoc --go_out=. --go_opt=paths=source_relative msgproto/*

run:
	@go run main.go

build:
	mkdir -p bin/ && go build -ldflags "-X version.Version=$(VERSION) -X version.Commit=$(COMMIT)" -o ./bin/ ./...