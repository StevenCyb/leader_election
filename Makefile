proto:
	protoc --go_out=. --go-grpc_out=. -I pkg/lcr/internal/rpc pkg/lcr/internal/rpc/lcr.proto