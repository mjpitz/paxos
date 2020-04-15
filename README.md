# paxos

Golang implementation of Paxos using protobuf and gRPC.
Currently implemented more as a demonstration, less for practicality.
Most of the concepts are in place.

This was largely implemented off of the Google tech talk found on [youtube](https://www.youtube.com/watch?v=d7nAGI_NZPk).
See the [notes.txt](notes.txt) file for the notes I took prior to implementing.

## Implementation

`Learners` are responsible for observing actions of `Acceptors`.
They do this through the use of a single, server streaming API.
Learners send along their last accepted id and Acceptors stream all records after that point.

`Acceptors` keep track of all the promises they make and proposals they accept.
Currently, this is implemented as an in memory log backed by a btree.
This can easily be replaced with a boltdb implementation for on disk support. 

`Proposer` provides a single `Propose` rpc that implements a single paxos run.
Through this mechanism, you can elect leaders, or get consensus on a value.

## Try it out

I probably won't publish a library or binaries from this for a while.
For now, stick to building locally. 

```bash
$ make build
```

Once the binary is built, you can spin up 3 copies of the server.

```bash
$ ./paxos --server-id 0 --members localhost:8080,localhost:8081,localhost:8082 --bind-address localhost:8080
$ ./paxos --server-id 1 --members localhost:8080,localhost:8081,localhost:8082 --bind-address localhost:8081
$ ./paxos --server-id 2 --members localhost:8080,localhost:8081,localhost:8082 --bind-address localhost:8082
```
