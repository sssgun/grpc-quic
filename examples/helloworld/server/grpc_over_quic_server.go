package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/lucas-clemente/quic-go"
	qnet "github.com/sssgun/grpc-quic"
	pb "github.com/sssgun/grpc-quic/examples/helloworld/helloworld"
	"google.golang.org/grpc"
	"log"
	"math/big"
)

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

func generateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	if len(certFile) >0 && len(keyFile) > 0 {
		log.Printf("generateTLSConfig] certFile=%s, keyFile=%s", certFile, keyFile)

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Printf("failed to tls.LoadX509KeyPair. %s", err.Error())
			return nil, err
		}
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"quic-echo-example"},
		}, nil
	} else {
		log.Printf("generateTLSConfig] GenerateKey")
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
