# LeLann, Chang and Roberts (LCR)
The LCR (LeLann, Chang, and Roberts) algorithm is a leader election method used in ring networks. Each node sends its unique ID around the ring, and the node with the highest ID becomes the leader. Once a node identifies itself as the leader, it sends a termination message around the ring to inform all other nodes.

## Usage
