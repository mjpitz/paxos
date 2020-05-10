default: install

build-deps:
	GO111MODULE=off go get -u github.com/golang/protobuf/protoc-gen-go
	GO111MODULE=off go get -u github.com/gogo/protobuf/protoc-gen-gogo
	GO111MODULE=off go get -u golang.org/x/lint/golint
	GO111MODULE=off go get -u oss.indeed.com/go/go-groups

fmt:
	go-groups -w .
	gofmt -s -w .

clean:
	rm -rf logs/
	rm -rf vendor/
	rm -f paxosc
	rm -f paxosd

vendor: go.mod go.sum
	go mod vendor

test:
	go vet ./...
	#golint -set_exit_status ./...
	go test -v ./...

api/paxos.pb.go: api/paxos.proto
	protoc -I=. -I=$(GOPATH)/src --gogo_out=plugins=grpc:$(GOPATH)/src api/paxos.proto

generate: api/paxos.pb.go

build:
	go build ./cmds/paxosc/
	go build ./cmds/paxosd/

docker:
	docker build -t mjpitz/paxos .
