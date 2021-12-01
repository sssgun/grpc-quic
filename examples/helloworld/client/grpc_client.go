package main

import (
	"context"
	"log"
	"time"

	pb "github.com/sssgun/grpc-quic/examples/helloworld/helloworld"
	"google.golang.org/grpc"
)

func echoGrpcClient() error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Printf("Client: did not connect. %s", err.Error())
		return err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Client: failed to close - grpc.Dial. %s", err.Error())
		}
	}(conn)
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Printf("Client: could not greet. %s", err.Error())
		return err
	}
	log.Printf("Greeting: %s", r.GetMessage())

	return nil
}