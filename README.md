# blockring
blockring is a building block for sharding, replicating and storing data.  It has a peer-to-peer master-less design, which is highly-available and scalable with no single point of failure.

## Data Categories
There are 2 main data categories.  This determines how data is written to the ring.

- Immutable
- Mutable

### Immutable

### Mutable

## TODO

- [x] Peer store
- [x] Client library
- [ ] Data
    - [x] Routing
        - [ ] Optimizations
    - [x] Datastore
- [ ] Graceful shutdown

### Roadmap

- [ ] Replication
- [ ] Mutable data
    - [ ] Log
    - [ ] Consensus

## References
- http://nms.lcs.mit.edu/papers/chord.pdf
- https://arxiv.org/pdf/1006.3465.pdf
