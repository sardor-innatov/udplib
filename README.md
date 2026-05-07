# udplib

A lightweight UDP library for Go that adds reliability on top of raw UDP — packet splitting, checksums, retransmission, and ACK/NACK handling. Not TCP. Faster than TCP where it matters.

---

## What It Does

Raw UDP gives you speed but no guarantees. udplib adds:

- **Packet splitting** — messages larger than 1400 bytes are automatically split into chunks and reassembled on the other side
- **Checksum validation** — every packet is fingerprinted with Ones Complement Sum (TCP checksum), corrupted packets are detected and discarded
- **ACK/NACK system** — sender waits for acknowledgement, receiver signals corruption explicitly
- **Exponential backoff retry** — lost packets are retransmitted with increasing timeouts (100ms → 200ms → 400ms → ...)
- **Single reader routing** — one goroutine owns the socket, routes ACKs and data to separate channels internally

---

## Project Structure

```
udplib/
│
├── packet.go                  Packet struct, serialize, deserialize
├── checksum.go                TCP checksum logic
├── conn.go                    Conn struct, ListenOn, Dial, Close
├── reader.go                  Read // internal read loop, ACK/data routing, assembling packets
├── writer.go                  WriteTo // sendWithRetry, exponential backoff, deviding into packets
└── go.mod
```

---

## Installation

```bash
go get github.com/sardor-innatov/udplib
```

---

## Quick Start

**Server:**
```go
package main

import (
    "fmt"
    "github.com/yourname/udplib"
)

func main() {
    conn, err := udplib.Listen(":8080")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    for {
        data, addr, err := conn.Read()
        if err != nil {
            continue
        }
        fmt.Printf("received from %s: %s\n", addr, string(data))
    }
}
```

**Client:**
```go
package main

import (
    "github.com/yourname/udplib"
)

func main() {
    conn, err := udplib.Dial("127.0.0.1:8080")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    lost, err := conn.WriteTo([]byte("Hello from client"), addr)
    if err != nil {
        panic(err)
    }

    if lost > 0 {
        fmt.Printf("warning: %d bytes failed to deliver\n", lost)
    }
}
```

---

## API

### `udplib.ListenOn(addr string) (*Conn, error)`
Binds a UDP socket to the given address and starts the internal read loop. Ready to receive immediately after return.

### `udplib.Dial(addr string) (*Conn, error)`
Creates a UDP socket on a random free port and starts the internal read loop.

### `(*Conn) WriteTo(data []byte, addr net.Addr) (uint32, error)`
Splits data into 1400 byte chunks, sends each with retry logic, returns number of bytes that failed to deliver after all attempts exhausted.

### `(*Conn) Read() ([]byte, net.Addr, error)`
Blocks until a complete message is assembled from all chunks. Returns the full reassembled payload and sender address.

### `(*Conn) Close() error`
Closes the underlying UDP socket.

---

## Packet Layout

Every packet serializes to:

```
    [0-3]      uint32    Seq            (4 bytes)
    [4-7]      uint32    Ack            (4 bytes)
    [8]        bool      IsAck          (1 byte)
    [9]	       bool      IsNack         (1 byte)
    [10]	   bool      IsFinal        (1 byte)
    [11-12]    uint16    Checksum       (2 bytes)
    [13-14]    uint16    Data Len       (2 bytes)
    [15-N]     []byte    The Payload    (actual data)
```

Total header: 15 bytes. Everything after is payload.

---

## Reliability Model

udplib uses **Stop-and-Wait ARQ** — each chunk is sent and acknowledged before the next is sent. Simple, correct, not the fastest possible approach.

```
sender                        receiver
  │                               │
  │  chunk seq=0  ─────────────►  │
  │               ◄─────────────  │  ACK
  │  chunk seq=1  ─────────────►  │
  │               ◄─────────────  │  ACK
  │  chunk seq=2  ─────────────►  │  (corrupted)
  │               ◄─────────────  │  NACK → resend
  │  chunk seq=2  ─────────────►  │
  │               ◄─────────────  │  ACK
```

Timeout uses exponential backoff starting at 100ms, up to 7 attempts per chunk.

---

## What udplib Is Not

- Not a replacement for TCP when you need guaranteed ordered delivery at scale
- Not suitable for high-throughput streaming (Stop-and-Wait is slow for large data)
- Not production hardened — built for learning and small scale use

For high performance reliable UDP at production scale look at **QUIC** (used by HTTP/3).

---

## Why Not Just Use TCP

Sometimes you need control over the reliability layer itself — custom timeouts, knowing exactly how many bytes failed, lower level visibility into what the network is doing. udplib gives you that. TCP hides everything from you. Sometimes that hiding is the problem.