# Leader Election in GoLang
This repository contains some leader election algorithms written in Go. 
Most of them were implemented just as a warm-up and may not be suitable for production use. 
In general, I plan to further develop one of the algorithms in order to use this implementation in another project (might be Raft or Zab).

# Algorithm
| Algorithm | Status       | Doc                        |
|-----------|--------------|----------------------------|
| Zab       | Experimental | TODO                       |
| Raft      | Experimental | [DOC](pkg/raft/README.md)  |
| Bully     | ~Stable      | [DOC](pkg/bully/README.md) |
| CR        | Experimental | [DOC](pkg/cr/README.md)    |
| HS        | Experimental | [DOC](pkg/hs/README.md)    |
| LCR       | Experimental | [DOC](pkg/lcr/README.md)   |

*Experimental algorithms are implemented in a basic form without optimization and may not fully adhere to the algorithm's exact specifications. They are subject to change or removal in the future.*

*~Stable algorithms might be used for production, but make sure to tests it before usage, to decide if it is stable enough for you.*
