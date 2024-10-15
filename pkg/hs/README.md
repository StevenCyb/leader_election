# Hirschenberg and Sinclair (HS)
The HS algorithm is used for leader election in synchronous ring networks. 
Each node sends its unique ID in both directions around the ring. 
The messages travel a distance that doubles each phase. 
Nodes forward IDs greater than their own and discard smaller ones. 
The process continues until a node receives its own ID back, confirming it as the leader. 
The leader then informs all other nodes.

## Usage
On you service setup the HS service like this:
```go
TODO
```

It is recommended to use TLS or even better mutual TLS to secure the communication. 
This can be configured using the `grpc.DialOption`. 
A simple example for that is show in the unit tests (`TestClientServer_TLS` and `TestClientServer_MutualTLS`) [here](internal/client_server_test.go).