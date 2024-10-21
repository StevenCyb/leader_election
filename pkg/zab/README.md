* Zab (Zookeeper Atomic Broadcast)


## Notes
# ZooKeeper Atomic Broadcast (Zab) Protocol

The ZooKeeper Atomic Broadcast (Zab) protocol is a consensus algorithm designed specifically for ZooKeeper to ensure data consistency across a distributed system. Here's a brief overview of how it works:

## Key Phases of Zab Protocol

1. **Leader Election**:
   - When the ZooKeeper cluster starts or when the current leader fails, a new leader is elected. All nodes in the cluster vote, and the node with the most up-to-date data is chosen as the leader.

2. **Recovery Phase**:
   - After a leader is elected, it synchronizes its state with the followers to ensure all nodes have the latest data. This phase ensures that any committed transactions are replicated across all nodes.

3. **Broadcast Phase**:
   - The leader handles all write requests from clients. It converts these requests into transaction proposals and broadcasts them to the followers. Once a majority of followers acknowledge the proposal, the leader commits the transaction and informs the followers to do the same.

## Modes of Operation

- **Message Broadcast**:
  - Used for handling client requests. The leader generates a transaction proposal for each write request and broadcasts it to the followers. The transaction is committed once a majority of followers acknowledge it.

- **Crash Recovery**:
  - If the leader crashes, the protocol ensures that any committed transactions are eventually committed by all nodes, and any uncommitted transactions are discarded. This ensures data consistency even in the event of failures.

ZooKeeper uses this protocol to maintain a highly available and consistent distributed system, making it a reliable choice for distributed coordination.

