package server

import (
	"crypto/tls"
	cx509 "crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	mygrpc "github.com/jackyzhangfudan/sidecar/pkg/grpc"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	port int32 = 8112
)

var server *certificateServiceServer = &certificateServiceServer{}

/*
Start gRPC server to accept certificate related request
*/
func Run(enableMTls bool, stopCh <-chan struct{}) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var s *googlegrpc.Server
	if enableMTls {
		tlsCre, err := createTLSCredentials()
		if err != nil {
			return
		}
		s = googlegrpc.NewServer(googlegrpc.Creds(tlsCre))
	} else {
		s = googlegrpc.NewServer()
	}
	mygrpc.RegisterCertificateServiceServer(s, server)

	go func() {
		log.Printf("server listening at %v, gRpc", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-stopCh
	s.Stop()
}

/*
we create the mTLS settings for server
we expect mTLS is enable between grpc client and server.

NOTE: following implementation is just for technical verification, isn't suitable for production,
because we use CA's root certificate as gRPC client and server's trust root certificate, there is a logic circle
*/
func createTLSCredentials() (credentials.TransportCredentials, error) {
	caPEMFile, err := ioutil.ReadFile("cert/rootCA/root.crt") //assume both grpc server and client's certificate are signed by same CA
	if err != nil {
		return nil, err
	}

	caPool := cx509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caPEMFile) {
		return nil, &ServerError{msg: "load local cert fail"}
	}

	localCert, err := tls.LoadX509KeyPair("cert/localCert/local.crt", "cert/localCert/local.private.key")
	if err != nil {
		log.Print("load local certificate and key file fail")
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{localCert},
		ClientAuth:   tls.RequireAndVerifyClientCert, //means mTLS, will check client's certificate
		ClientCAs:    caPool,
	}

	return credentials.NewTLS(config), nil
}
