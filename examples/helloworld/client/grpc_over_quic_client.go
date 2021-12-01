package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	qnet "github.com/sssgun/grpc-quic"
	pb "github.com/sssgun/grpc-quic/examples/helloworld/helloworld"
	"google.golang.org/grpc"
)

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