# VectorMesh 🚀
*A lightweight, distributed vector-retrieval and RAG gateway built in Go with concurrent fan-out and in-memory LRU caching.*

## Architecture Overview
VectorMesh decouples heavy vector similarity searches across distributed database shards while providing a centralized, low-latency API gateway featuring native Go concurrency and caching strategies.

[ Client Request ]
│ (HTTP POST)
▼
[ Go API Gateway ] ──(LRU Cache Check: Hit/Miss)
│
├──► [ Cache Hit ] ──► Returns cached RAG context instantly (Zero Network Latency)
│
└──► [ Cache Miss ] ──(Concurrent Goroutines / Fan-Out)──► [ Worker Shard Node ]
▼
[ RAG Prompt Assembly ]

## Core Features
* **Concurrent Fan-Out Routing:** Leverages Go goroutines and channels to query multiple distributed vector shards simultaneously with built-in timeout handling.
* **Thread-Safe LRU Cache:** Implements an in-memory Least Recently Used caching layer with thread locks to instantly serve frequent queries and minimize downstream network latency.
* **RAG Context Orchestration:** Dynamically aggregates retrieved text chunks and similarity scores into a structured prompt context ready for LLM consumption.

## Tech Stack
* **Language:** Go (API Gateway, Concurrency, LRU Cache), Python (Worker Nodes & Client Script)
* **Concurrency Primitives:** Goroutines, Channels, Mutexes (`sync.Mutex`), Container Lists
* **Protocols:** HTTP/REST
