package server

import (
	"context"
	"log"

	"github.com/jackyzhangfudan/sidecar/pkg/ca"
	mygrpc "github.com/jackyzhangfudan/sidecar/pkg/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type ServerError struct {
	msg string
}

func (e *ServerError) Error() string {
	return e.msg
}

type certificateServiceServer struct {
	mygrpc.UnimplementedCertificateServiceServer
}

/*
return a csr template, to easy requester's work
*/
func (s *certificateServiceServer) CsrTemplate(context.Context, *emptypb.Empty) (*mygrpc.CertificateSigningRequest, error) {
	csr := &mygrpc.CertificateSigningRequest{
		SubjectCountry:            []string{"China"},
		SubjectOrganization:       []string{"Qinghua"},
		SubjectOrganizationalUnit: []string{"ComputerScience"},
		SubjectProvince:           []string{"Beijing"},
		SubjectLocality:           []string{"北京"},

		SubjectCommonName: "www.tsinghua.edu.cn",
		EmailAddresses:    []string{"ex@example.com"},
	}
	return csr, nil
}

/*
Sing a certificate signing request
*/
func (s *certificateServiceServer) SignCsr(ctx context.Context, csrReq *mygrpc.CertificateSigningRequest) (*mygrpc.SignResponse, error) {
	csr := &ca.CertificateSigningRequest{}

	csr.DNSNames = csrReq.DNSNames
	csr.EmailAddresses = csrReq.EmailAddresses
	csr.SubjectCommonName = csrReq.SubjectCommonName
	csr.SubjectCountry = csrReq.SubjectCountry
	csr.SubjectLocality = csrReq.SubjectLocality
	csr.SubjectOrganization = csrReq.SubjectOrganization
	csr.SubjectOrganizationalUnit = csrReq.SubjectOrganizationalUnit
	csr.SubjectPostalCode = csrReq.SubjectPostalCode
	csr.SubjectProvince = csrReq.SubjectProvince
	csr.SubjectSerialNumber = csrReq.SubjectSerialNumber
	csr.SubjectStreetAddress = csrReq.SubjectStreetAddress

	theCert, err := ca.CA.SignX509(csr)

	if err != nil {
		return nil, status.Error(codes.Internal, "singing csr fail")
	}

	result := &mygrpc.SignResponse{CertificateId: theCert.ID}
	return result, nil
}

/*
return the generated certificate
*/
func (s *certificateServiceServer) GetCert(ctx context.Context, in *mygrpc.FileIdentifer) (*mygrpc.FileStream, error) {
	contents, err := ca.CA.GetCertFile(in.Id)
	if err != nil {
		log.Printf("can't find the expected client certificate file %v", err)
		return nil, &ServerError{msg: "can't get the expected file"}
	}
	return &mygrpc.FileStream{Contents: contents}, nil
}

/*
return the generated private key
*/
func (s *certificateServiceServer) GetKey(ctx context.Context, in *mygrpc.FileIdentifer) (*mygrpc.FileStream, error) {
	contents, err := ca.CA.GetKeyFile(in.Id)
	if err != nil {
		log.Printf("can't find the expected client private key file %v", err)
		return nil, &ServerError{msg: "can't get the expected file"}
	}
	return &mygrpc.FileStream{Contents: contents}, nil
}
