# VectorMesh
A lightweight, distributed vector-retrieval and RAG gateway built for high-throughput AI infrastructure.
## Core Features
* **Concurrent Fan-Out Architecture:** Leverages Go goroutines and channels to query multiple distributed vector shards simultaneously with built-in timeout handling.
* **Low-Level Vector Similarity:** In-memory shard execution optimizing cosine similarity and high-dimensional vector lookups.
* **Production-Grade Design:** Structured for horizontal scaling, fault-tolerant node handling, and minimal networking overhead.

## Tech Stack
* **Language:** Go (Gateway & Nodes), Python (Embedding/RAG layer)
* **Concurrency:** Go Routines, Mutexes, Channels
* **Protocols:** HTTP/REST, TCP/gRPC routing patterns
