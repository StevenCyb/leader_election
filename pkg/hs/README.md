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

* HS https://www.cs.utexas.edu/~lorenzo/corsi/cs380d/past/03F/notes/9-6.pdf