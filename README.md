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

### As a server
```go
func echoGrpcQuicServer(certFile, keyFile string) error {
	log.Println("starting echo quic server")

	tlsConf, err := generateTLSConfig(certFile, keyFile)
	if err != nil {
		log.Printf("QuicServer: failed to generateTLSConfig. %s", err.Error())
		return err
	}

	ql, err := quic.ListenAddr(*addr, tlsConf, nil)
	if err != nil {
		log.Printf("QuicServer: failed to ListenAddr. %s", err.Error())
		return err
	}
	listener := qnet.Listen(ql)

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("QuicServer: listening at %v", listener.Addr())

	if err := s.Serve(listener); err != nil {
		log.Printf("QuicServer: failed to serve. %v", err)
		return err
	}

	log.Println("stopping echo quic server")
	return nil
}
```


### As a client
```go
func echoGrpcQuicClient() error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}

	creds := qnet.NewCredentials(tlsConf)

	dialer := qnet.NewQuicDialer(tlsConf)
	grpcOpts := []grpc.DialOption{
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(creds),
	}

	conn, err := grpc.Dial(*addr, grpcOpts...)
	if err != nil {
		log.Printf("QuicClient: failed to grpc.Dial. %s", err.Error())
		return err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("QuicClient: failed to close - grpc.Dial. %s", err.Error())
		}
	}(conn)

	c := pb.NewGreeterClient(conn)
	name := defaultName
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Printf("QuicClient: could not greet. %v", err)
		return err
	}
	log.Printf("QuicClient: Greeting=%s", r.GetMessage())

	return nil
}
```
