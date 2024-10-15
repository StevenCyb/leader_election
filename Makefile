proto:
	protoc --go_out=. --go-grpc_out=. -I pkg/lcr/internal/rpc pkg/lcr/internal/rpc/lcr.proto
	protoc --go_out=. --go-grpc_out=. -I pkg/hs/internal/rpc pkg/hs/internal/rpc/hs.proto
	protoc --go_out=. --go-grpc_out=. -I pkg/cr/internal/rpc pkg/cr/internal/rpc/cr.proto