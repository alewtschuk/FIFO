# FIFO Queue Monitor in Go

Author: Alex Lewtschuk

This project implements a thread-safe, monitor-style **bounded FIFO queue** in Go. It is designed to solve the **bounded buffer problem**, where producers and consumers must coordinate safely when sharing a fixed-size queue.

The queue uses **mutexes** and **condition variables** to block producers when full and consumers when empty. It supports graceful shutdown and is fully tested for **concurrency safety** using Go’s `-race` detector.

---

## Features

- Fixed-capacity circular queue
- Blocking `Enqueue()` if full
- Blocking `Dequeue()` if empty
- Graceful `Shutdown()` that unblocks all goroutines

---

## Tests

In order to run standard test please run:

```bash
make test
```

In order to run race testing please run:

```bash
make race
```

