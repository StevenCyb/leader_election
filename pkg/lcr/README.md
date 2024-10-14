# LeLann, Chang and Roberts (LCR)
## **DO NOT USE THIS: This implementation is not accurate to the algorithm. I will leave it here for now but this implementation might be recreated or removed later.**
   
The LCR (LeLann, Chang, and Roberts) algorithm is a leader election method used in ring networks. 
Each node sends its unique ID around the ring, and the node with the highest ID becomes the leader. 
Once a node identifies itself as the leader, it sends a termination message around the ring to inform all other nodes.

## Usage
On you service setup the LCR service like this:
```go
id, err := strconv.ParseUint(os.Getenv("LB_ID"), 10, 64)
// ...
logger := log.New().Build()
instance, err := New(id, listen, logger)
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