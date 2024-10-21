proto:
	protoc --go_out=. --go-grpc_out=. -I pkg/lcr/internal/rpc pkg/lcr/internal/rpc/lcr.proto
	protoc --go_out=. --go-grpc_out=. -I pkg/hs/internal/rpc pkg/hs/internal/rpc/hs.proto
	protoc --go_out=. --go-grpc_out=. -I pkg/cr/internal/rpc pkg/cr/internal/rpc/cr.proto
	protoc --go_out=. --go-grpc_out=. -I pkg/bully/internal/rpc pkg/bully/internal/rpc/bully.proto
	protoc --go_out=. --go-grpc_out=. -I pkg/raft/internal/rpc pkg/raft/internal/rpc/raft.proto
	protoc --go_out=. --go-grpc_out=. -I pkg/zab/internal/rpc pkg/zab/internal/rpc/zab.proto