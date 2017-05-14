
clean:
	go clean -i ./...

deps:
	[ -e glide ] || curl https://glide.sh/get | sh
	glide restore

protoc-structs:
	@rm -f structs/structs.pb.go
	protoc -I ../../../ -I ./structs ./structs/structs.proto --go_out=plugins=grpc:./structs

protoc-rpc:
	@rm -f rpc/rpc.pb.go
	protoc -I ../../../ -I ./rpc ./rpc/rpc.proto --go_out=plugins=grpc:./rpc


protoc: protoc-structs protoc-rpc

test:
	go test -cover ./...
