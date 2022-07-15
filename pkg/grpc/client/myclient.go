package client

import (
	"context"
	"crypto/tls"
	cx509 "crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	mygrpc "github.com/jackyzhangfudan/sidecar/pkg/grpc"
	googlerpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

const (
	port int32 = 8112
)

func createTLSCredentials(clientId string) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	rootCAFile, err := ioutil.ReadFile("cert/rootCA/root.crt")
	if err != nil {
		return nil, err
	}

	certPool := cx509.NewCertPool()
	if !certPool.AppendCertsFromPEM(rootCAFile) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair("cert/clientCert/"+clientId+".crt", "cert/clientCert/"+clientId+".key")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		ServerName:   "localhost", //server 必须是对应的这个值
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}

func Run(clientId string) {
	tlsCredentials, err := createTLSCredentials(clientId)
	if err != nil {
		log.Print("cannot load TLS credentials: ", err)
		return
	}

	cc1, err := googlerpc.Dial(fmt.Sprintf(":%v", port), googlerpc.WithTransportCredentials(tlsCredentials))
	if err != nil {
		log.Print("cannot dial server: ", err)
		return
	}
	defer cc1.Close()

	rpcClient := mygrpc.NewCertificateServiceClient(cc1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	template, err := rpcClient.CsrTemplate(ctx, &emptypb.Empty{})
	if err != nil {
		log.Print("error happen when call gRPC client:" + err.Error())
		return
	}

	fmt.Print(template)
}
