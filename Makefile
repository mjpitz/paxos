default: install

build-deps:
	GO111MODULE=off go get -u github.com/golang/protobuf/protoc-gen-go
	GO111MODULE=off go get -u github.com/gogo/protobuf/protoc-gen-gogo
	GO111MODULE=off go get -u golang.org/x/lint/golint

deps:
	go get -v ./...

test:
	go vet ./...
	#golint -set_exit_status ./...
	go test -v ./...

generate:
	protoc -I=. -I=$(GOPATH)/src --gogo_out=plugins=grpc:$(GOPATH)/src api/paxos.proto

build:
	go build
