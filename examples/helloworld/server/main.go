package main

import (
	"context"
	"flag"
	pb "github.com/sssgun/grpc-quic/examples/helloworld/helloworld"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	addr = flag.String("addr", ":4242", "server address")
)

// server is used to implement hello.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements hello.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	var certFile string
	var keyFile string
	var enableGRPC bool

	flag.BoolVar(&enableGRPC, "grpc", false, "enable grpc server")
	flag.StringVar(&certFile, "cert", "", "cert file")
	flag.StringVar(&keyFile, "key", "", "key file")

	flag.Parse()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	if enableGRPC {
		go func() {
			err := echoGrpcServer()
			if err != nil {
				log.Printf("failed to Echo Server. %s", err.Error())
				return
			}
		}()
	}

	go func() {
		err := echoGrpcQuicServer(certFile, keyFile)
		if err != nil {
			log.Printf("failed to Echo QUIC Server. %s", err.Error())
			return
		}
	}()

	exitChan := make(chan int)
	go func() {
		for {
			s := <-signalChan
			switch s {
			// kill -SIGHUP XXXX
			case syscall.SIGHUP:
				log.Println("hungup")
				exitChan <- 0

			// kill -SIGINT XXXX or Ctrl+c
			case syscall.SIGINT:
				log.Println("Warikomi")
				exitChan <- 0

			// kill -SIGTERM XXXX
			case syscall.SIGTERM:
				log.Println("force stop")
				exitChan <- 0

			// kill -SIGQUIT XXXX
			case syscall.SIGQUIT:
				log.Println("stop and core dump")
				exitChan <- 0

			default:
				log.Println("Unknown signal.")
				exitChan <- 1
			}
		}
	}()

	log.Println("awaiting signal")
	code := <-exitChan
	os.Exit(code)
}