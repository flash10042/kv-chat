# kv-chat

`kv-chat` is a small, in-memory key–value store written in Go, inspired by Redis but not compatible with it.
The primary purpose of this project is learning Go by building a networked, concurrent system with a clean internal architecture.

The store is intended to be used as a lightweight backend for chat history storage in Python applications.

## Goals

* Learn idiomatic Go project structure
* Implement a TCP server
* Design and enforce a simple text-based protocol
* Implement concurrent, in-memory data structures
* Support basic key–value and list operations
* Provide a usable backend for chat history storage


## High-Level Design

* Single-process Go server

* In-memory data store

* TCP-based, line-oriented protocol

* One goroutine per client connection

* Shared store protected by mutexes

* Lazy expiration for keys with TTL

## Supported Commands (Initial Scope)

### General
* `PING`
* `SET key value`
* `GET key`
* `DEL key`
* `EXISTS key`
### Expiration
* `SETEX key ttl value`
* `TTL key`
### Lists (for chat history)
* `LPUSH key value`
* `LRANGE key start end`

Command syntax and responses are Redis-inspired but intentionally simplified.