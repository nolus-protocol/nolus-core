# Offload the Cosmos App from client queries

- Status: accepted
- Deciders: the Nolus dev team
- Date: 2021-00-19
- Tags:

## Context and Problem Statement

The clients need data that could be retrieved from the Cosmos App by using its native query mechanism. There are two drawbacks:
- the Nolus Validator Nodes would be overwhelmed by a lot of requests to serve which would collide with their main responsibility, and
- the data structure and detailness would not be convenient for clients to operate on.

Shall we bother Nolus Validator Nodes with client queries?
Read-only view of the application state
The separation of queries from commands/transactions aims to

offload the core service,
enable perfect horizontal scalability, and
keep the data in a form that allows faster retrieval, e.g. relations, aggegations, filtrations, etc.

## Decision Outcome

Provide a separate service that would collect, filter, re-structure and aggregate data from Nolus Validator Nodes into content that is suitable for client needs. We call the container providing that service `Application Server`. All client queries are served by that service.

The service sources data from any of the nodes and builds its own internal data. The data may be persisted for quick startup and reused by horizontally scaled instances.

The service should not contain any client-related state to allow perfect scalability.
