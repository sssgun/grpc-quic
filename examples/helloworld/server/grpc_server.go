package main

import (
	pb "github.com/sssgun/grpc-quic/examples/helloworld/helloworld"
	"google.golang.org/grpc"
	"log"
	"net"
)

func echoGrpcServer() error {
	log.Println("starting echo server")

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Printf("Server: failed to listen. %s", err.Error())
		return err
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("Server: listening at %s", lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		log.Printf("Server: failed to serve. %s", err.Error())
		return err
	}

	log.Println("stopping echo server")
	return nil
}