# Chang and Roberts (CR)
## **DO NOT USE THIS: This implementation is not accurate to the algorithm. I will leave it here for now but this implementation might be recreated or removed later.**
   
The Chang and Roberts (CR) algorithm is used for leader election in unidirectional ring networks. 
Each process starts as a non-participant. 
A process noticing no leader starts an election by sending its UID clockwise. 
Processes forward larger UIDs and replace smaller ones if they haven’t participated yet. 
If a process receives its own UID, it becomes the leader and sends an elected message around the ring. 
All processes mark themselves as non-participants and acknowledge the leader.

## Usage
On you service setup the LCR service like this:
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