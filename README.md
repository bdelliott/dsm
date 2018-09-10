# Distributed state machine 

DSM provides machinery to execute a standard standard machine in distributed fashion.

# How it works

## Actors

* Client - process that submits state machine instances for execution
* Worker - process that polls etcd for state machine instances that are ready to be worked.
* etcd (3) - coordination layer providing the shared locks utilized by worker processes.

## Process

The basic process for distributed state machine processing is as follows:

1. The client submit a state machine instance M to etcd 3.
2. M is saved under a unique key.
3. A worker process W will poll etcd and discover M.
4. W will attempt to lock M.  
5. If W successfully locks M, it will call the M's registered "tick" function to perform the state machine's logic.
6. M will be advanced to a new state with an updated payload.
7. W will save M's updates back to etcd
8. W unlocks M.

These steps will continue until M reaches its terminal state.  Multiple worker processes can operate concurrently, each attempting to perform work on unlocked instances of state machines.

# Concurrency

Each individual machine instance is currently serial on a per instance basis.  An instance is locked and unlocked through each state transition.  This model is a bit contrived for distributed processing, but is basically equivalent to the non-distributed model of state machine processing.

An extension of this work could be done to create states that allow for parallel execution.  This would require an ability to describe a state transition to a set of "next states" that can run in parallel and when to join back to a single state.

i.e.:

Serial state transition: 

`A->B`

Parallel state transition:

```
  ->B
A	  ->D
  ->C
```
Meaning, A transitions to both states B and C.  B and C may be completed by parallel workers before running sequentially again in state D.

# Running

The repository comes with a `Dockerfile` and `docker-compose.yml` for easy execution.

1. `docker-compose build` will build the necessary Docker container image.
2. `docker-compose up --scale work=2` will start the application components (etcd, worker, and client) with 2 worker processes.

# Tests

Unit tests can be executed with: `go test -v`