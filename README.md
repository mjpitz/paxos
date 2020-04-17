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
$ mkdir -p logs/0 logs/1 logs/2

$ ./paxos \
    --server-id 0 \
    --members localhost:8080,localhost:8081,localhost:8082 \
    --bind-address localhost:8080 \
    --promise-log boltdb://logs/0/promise.log \
    --accept-log boltdb://logs/0/accept.log \
    --decision-log boltdb://logs/0/decision.log

$ ./paxos \
    --server-id 1 \
    --members localhost:8080,localhost:8081,localhost:8082 \
    --bind-address localhost:8081 \
    --promise-log boltdb://logs/1/promise.log \
    --accept-log boltdb://logs/1/accept.log \
    --decision-log boltdb://logs/1/decision.log
    
$ ./paxos \
    --server-id 2 \
    --members localhost:8080,localhost:8081,localhost:8082 \
    --bind-address localhost:8082 \
    --promise-log boltdb://logs/2/promise.log \
    --accept-log boltdb://logs/2/accept.log \
    --decision-log boltdb://logs/2/decision.log
```

As the program runs, you should be able to inspect shas of the log files:

```bash
$ find logs/ -type f | xargs sha256sum 
7f9fc12159567d21113663035918699f4b7e73ab00af3570067427617925e687  logs/0/accept.log
7f9fc12159567d21113663035918699f4b7e73ab00af3570067427617925e687  logs/0/decision.log
604bbf65a60b99732b055a760adff32e422101ceccfc633994ece5dbe3547c57  logs/0/promise.log
7f9fc12159567d21113663035918699f4b7e73ab00af3570067427617925e687  logs/1/accept.log
7f9fc12159567d21113663035918699f4b7e73ab00af3570067427617925e687  logs/1/decision.log
604bbf65a60b99732b055a760adff32e422101ceccfc633994ece5dbe3547c57  logs/1/promise.log
7f9fc12159567d21113663035918699f4b7e73ab00af3570067427617925e687  logs/2/accept.log
7f9fc12159567d21113663035918699f4b7e73ab00af3570067427617925e687  logs/2/decision.log
604bbf65a60b99732b055a760adff32e422101ceccfc633994ece5dbe3547c57  logs/2/promise.log
```
