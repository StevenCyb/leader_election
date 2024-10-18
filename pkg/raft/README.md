https://chelseatroy.com/2020/08/30/raft-10-leader-elections-part-1/

https://raft.github.io/


Leader election with distributed consensus.

NODE:
* Node A
* Term: 1
* Voted For: C

Three states:
* Follower
* Candidate
* Leader

majority = totalNodes/2 <= totalRespondingNodes+deadNodes

Two timeouts for leader election:
* election timeout -> amount of time a follower waits until becoming a candidate (random between 150ms and 300ms)
* heartbeat timeout -> 

1) At start all nodes are follower
2) First that finds no leader or that the leader is down, they become a candidate
3) Leader election
   3.1) New leader start a new term
   3.2) Candidates request vote from other
   3.3) Follower nodes will respond with there vote (follower can only vote to the first one they get a request + reset there election timeout)
   3.4) The candidate becomes the leader if it gets votes from a majority of nodes
        * a split vote can occur
          3.4.1) Will get to the next term and pick new election timeout
   3.5) Leader sends Heartbeat Append Entries message to his followers (they respond and reset there election timeout)
        * If a "leader" gets a leader heartbeat from higher term
          3.5.1) its accept him as leader
          * based on state:
            3.5.1) Leader A has higher term and fewer committed logs:
                   Action: Leader A will send its log entries to the followers. Followers will compare the logs and send back any missing committed entries to Leader A. Leader A will update its log to include these committed entries.
            3.5.2) Leader A has higher term and fewer uncommitted logs:
                   Action: Leader A will send its log entries to the followers. Followers will respond with their logs, and Leader A will update its log to match the followers’ logs, ensuring all uncommitted entries are included.
            3.5.3) Leader A has higher term and more uncommitted logs:
                   Action: Leader A will send its log entries to the followers. Followers will accept these entries and update their logs to match Leader A’s log. Uncommitted entries will be replicated across the cluster.
            3.5.4) Leader A has higher term and more committed logs:
                   Action: Leader A will send its log entries to the followers. Followers will accept these entries and update their logs to match Leader A’s log. Committed entries will be propagated to ensure consistency. 
4) Log Replication
  4.4) Changes are going trough leader and they are added to the log
  4.5) At first he log is uncommitted
  4.6) This log is than send to all follower with the next Append Entries message (heartbeat)
  4.7) Respond to client sended the change request (not needed for me?)
  4.8) Leader is waiting till the majority have written the entry
  4.9) The log on the leader is now committed
  4.10) Leader notify follower that log is now committed, so also for them