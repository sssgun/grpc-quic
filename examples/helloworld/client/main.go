package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "127.0.0.1:4242", "server address")
)

func main() {
	var enableGRPC bool

	flag.BoolVar(&enableGRPC, "grpc", false, "enable grpc client")
	flag.Parse()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	if enableGRPC {
		if err := echoGrpcClient(); err != nil {
			log.Printf("failed to Client. %s", err.Error())
			return
		}
	}

	if err := echoGrpcQuicClient(); err != nil {
		log.Printf("failed to QUIC Client. %s", err.Error())
		return
	}

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