# Raft
The Raft algorithm is a consensus algorithm and leader election in distributed systems. Each server starts as a follower. If a follower doesn’t receive a heartbeat from the leader within a timeout, it becomes a candidate and starts an election by requesting votes from other servers. Servers vote for the candidate with the most up-to-date log. If a candidate receives a majority of votes, it becomes the leader and sends heartbeats to maintain its authority. If no candidate wins, the process repeats. The leader manages log replication to ensure consistency across the cluster.

An visualization of this algorithm is provide on the [Raft Homepage](https://raft.github.io/) and a more precise explanation [here](https://thesecretlivesofdata.com/raft/).

## Usage
On you service setup the Raft service as follow. 
This implementation does not handle log replication, therefore the `resolveLeaderElection` is needed to resolve conflicts,
where the log may decide the new leader. 
```go
uid, err := strconv.ParseUint(os.Getenv("LB_ID"), 10, 64)
// ...
var resolveLeaderElection = func(potentialLeader uint64) bool {
    return potentialLeader > uid
}
// ...
logger := log.New().Build()
instance, err := New(uid, listen, logger, resolveLeaderElection)
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
