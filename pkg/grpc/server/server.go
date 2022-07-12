package server

import (
	"fmt"
	"log"
	"net"

	mygrpc "github.com/jackyzhangfudan/sidecar/pkg/grpc"
	googlegrpc "google.golang.org/grpc"
)

const (
	port int32 = 8112
)

var server *certificateServiceServer = &certificateServiceServer{}

/*
Start gRPC server to accept certificate related request
*/
func Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := googlegrpc.NewServer()
	mygrpc.RegisterCertificateServiceServer(s, server)
	log.Printf("server listening at %v, gRpc", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
