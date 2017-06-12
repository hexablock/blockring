# blockring
blockring is a building block for sharding, replicating and storing data.  It has a peer-to-peer master-less design, which is highly-available and scalable with no single point of failure.

## Data Categories
There are 2 main data categories.  This determines how data is written to the ring.

- Immutable
- Mutable

### Immutable
Immutable data does not go through the consensus process and is directly written.  Data blocks are written
in this manner.

### Mutable (experimental)
Mutable data goes through the consensus process i.e. the log.

### Roadmap

- [ ] Node failure
- [ ] Node leave
- [ ] Replication
- [ ] Compaction
- [ ] Optimizations

## References
- http://nms.lcs.mit.edu/papers/chord.pdf
- https://arxiv.org/pdf/1006.3465.pdf
