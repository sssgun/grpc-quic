# gRPC over QUIC

The Go language implementation of gRPC over QUIC.

* gRPC-Go + QUIC-Go
  * https://github.com/grpc/grpc-go
  * https://github.com/lucas-clemente/quic-go
* Improved 'github.com/gfanton/grpc-quic'

## Prerequisites

- **[Go][]**: any one of the **three latest major** [releases][go-releases].

## Installation

With [Go module][] support (Go 1.11+), simply add the following import

```go
import "github.com/sssgun/grpc-quic"
```

## Usage
* gRPC-Go + QUIC-Go
  * https://github.com/grpc/grpc-go/tree/master/examples/helloworld
  * https://github.com/lucas-clemente/quic-go/tree/master/example

### As a server
See the example server.

### As a client
See the example client.