
#BRANCH = $(shell git rev-parse --abbrev-ref HEAD || echo unknown)
#COMMIT = $(shell git rev-parse --short HEAD || echo unknown)
#BUILDTIME = $(shell date +%Y-%m-%dT%T%z)

#LD_OPTS = -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w"
#LD_OPTS = -ldflags="-X difuse.version=${VERSION} -w"
#BUILD_CMD = CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo

clean:
	go clean -i ./...

protoc-structs:
	@rm -f structs/structs.pb.go
	protoc -I ../../../ -I ./structs ./structs/structs.proto --go_out=plugins=grpc:./structs

protoc-rpc:
	@rm -f rpc/rpc.pb.go
	protoc -I ../../../ -I ./rpc ./rpc/rpc.proto --go_out=plugins=grpc:./rpc


protoc: protoc-structs protoc-rpc

test:
	go test -cover ./...
