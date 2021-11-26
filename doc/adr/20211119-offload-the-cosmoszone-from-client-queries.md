# Offload the Cosmos App from client queries

- Status: accepted
- Deciders: the Nolus dev team
- Date: 2021-11-19
- Tags:

## Context and Problem Statement

The clients need data that is usually retrieved from the Cosmos App by using its native query mechanism. There are two drawbacks:
- the Nolus Validator Nodes would be overwhelmed by a lot of requests to serve which would collide with their main responsibility, and
- the data structure and detailness would not be convenient for clients to operate on.

Shall we bother Nolus Validator Nodes with client queries?

Another aspect is system scalability. The system should be able to handle growing number of client requests without affecting the main work of the validators.

## Decision Outcome

Provide a separate service that would collect, filter, re-structure and aggregate data from Nolus Validator Nodes into content that is suitable for client needs. We call the container providing that service `Application Server`. All client queries are served by that service.

The service sources data from any of the nodes and builds its own internal read-only data. The data may be persisted for quick startup and reused by horizontally scaled instances.

The service should not contain any client-related state to allow perfect scalability.
