package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lucas-clemente/quic-go"
	qnet "github.com/sssgun/grpc-quic"
	pb "github.com/sssgun/grpc-quic/examples/helloworld/helloworld"
	"google.golang.org/grpc"
)

const (
	addr = "localhost:4242"
	defaultName = "world"
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
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		err := echoServer()
		if err != nil {
			log.Printf("failed to Echo Server. %s", err.Error())
			return
		}
	}()

	go func() {
		err := echoQuicServer()
		if err != nil {
			log.Printf("failed to Echo QUIC Server. %s", err.Error())
			return
		}
	}()

	if err := echoClient(); err != nil {
		log.Printf("failed to Client. %s", err.Error())
		return
	}

	if err := echoQuicClient(); err != nil {
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

func echoServer() error {
	log.Println("starting echo server")

	lis, err := net.Listen("tcp", addr)
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

func echoQuicServer() error {
	log.Println("starting echo quic server")

	tlsConf, err := generateTLSConfig(false)
	if err != nil {
		log.Printf("QuicServer: failed to generateTLSConfig. %s", err.Error())
		return err
	}

	ql, err := quic.ListenAddr(addr, tlsConf, nil)
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

func echoClient() error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
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

func echoQuicClient() error {
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

	conn, err := grpc.Dial(addr, grpcOpts...)
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

func generateTLSConfig(isFile bool) (*tls.Config, error) {
	if isFile {
		cert, err := tls.LoadX509KeyPair(
			"/etc/letsencrypt/live/pangyo-dev01.kro.kr/fullchain.pem",
			"/etc/letsencrypt/live/pangyo-dev01.kro.kr/privkey.pem")
		if err != nil {
			log.Printf("failed to tls.LoadX509KeyPair. %s", err.Error())
			return nil, err
		}
		return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
	} else {
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			log.Printf("failed to rsa.GenerateKey. %s", err.Error())
			return nil, err
		}
		template := x509.Certificate{SerialNumber: big.NewInt(1)}
		certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
		if err != nil {
			log.Printf("failed to x509.CreateCertificate. %s", err.Error())
			return nil, err
		}
		keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
		certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

		tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			log.Printf("failed to tls.X509KeyPair. %s", err.Error())
			return nil, err
		}

		return &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			NextProtos:   []string{"quic-echo-example"},
		}, nil
	}
}
