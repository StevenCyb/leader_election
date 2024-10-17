# Bully
## **DO NOT USE THIS: This implementation is not accurate to the algorithm. I will leave it here for now but this implementation might be recreated or removed later.**
   
The Bully algorithm is used for leader election in distributed systems. When a process detects the absence of a leader, it initiates an election by sending an election message to all processes with higher IDs. If no higher-ID process responds, the initiating process becomes the leader and announces its victory. If a higher-ID process responds, it takes over the election process. This continues until the process with the highest ID is elected as the leader and informs all other processes.

## Usage
On you service setup the Bully service like this:
```go
uid, err := strconv.ParseUint(os.Getenv("LB_ID"), 10, 64)
// ...
logger := log.New().Build()
instance, err := New(uid, listen, logger)
// ...
// Add all other nodes from the cluster with there id and listen address
instance.AddNode(id, listen, grpc.WithTransportCredentials(insecure.NewCredentials()))
// ...
// Start the leader service
instance.MustStart(delayDuration, checkIntervalDuration)
// ...
// Optionally register for leader change event if needed
instance.OnLeaderChange(func(leader *uint64) {
  if leader != nil {
    // Got new leader
  }
})
// Get the leader
leader := instance.GetLeader()
// Check if is leader 
if instance.IsLeader() {
  // ...
}
// ...
```

It is recommended to use TLS or even better mutual TLS to secure the communication. 
This can be configured using the `grpc.DialOption`. 
A simple example for that is show in the unit tests (`TestClientServer_TLS` and `TestClientServer_MutualTLS`) [here](internal/client_server_test.go).