package ca

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	cx509 "crypto/x509"
	"crypto/x509/pkix"
	"log"
	"math/big"
	mathRand "math/rand"
	"time"
)

var CA CertificateAuthority

func init() {
	CA = CertificateAuthority{}
	CA.load()
}

type CertificateAuthority struct {
	RootCA     cx509.Certificate
	PrivateKey rsa.PrivateKey
}

/*
load ca data
*/
func (ca *CertificateAuthority) load() {
	//TODO
	// if ca.PrivateKey ==  nil {
	// 	ca.PrivateKey, _ = rsa.GenerateKey(rand.Reader, 1024)
	// }
	// ca.PrivateKey.PublicKey
}

/*
sign a csr
*/
func (ca *CertificateAuthority) SignX509(csr *CertificateSigningRequest) (*cx509.Certificate, error) {
	cx509CSR := csr.toCX509CSR(nil)

	mathRand.Seed(time.Now().UnixNano())
	cx509CertificateTemplate := cx509.Certificate{
		Version:            cx509CSR.Version,
		SerialNumber:       big.NewInt((int64)(mathRand.Int())),
		Signature:          cx509CSR.Signature,
		SignatureAlgorithm: cx509CSR.SignatureAlgorithm,
		PublicKey:          cx509CSR.PublicKey,
		PublicKeyAlgorithm: cx509CSR.PublicKeyAlgorithm,
		Subject:            cx509CSR.Subject,
		URIs:               cx509CSR.URIs,
		DNSNames:           cx509CSR.DNSNames,
		Extensions:         cx509CSR.Extensions,
		EmailAddresses:     cx509CSR.EmailAddresses,
		IPAddresses:        cx509CSR.IPAddresses,
	}

	buf, err := cx509.CreateCertificate(rand.Reader, &cx509CertificateTemplate, &ca.RootCA, cx509CSR.PublicKey, ca.PrivateKey)
	if err != nil {
		log.Fatal("sign the x509 csr fail")
	}
	res, err := cx509.ParseCertificate(buf)
	if err != nil {
		log.Fatal("parse the cx509 certificate fail")
	}

	return res, err
}

/*
transfer my CSR to x509 CSR
*/
func (csr *CertificateSigningRequest) toCX509CSR(signer crypto.Signer) *cx509.CertificateRequest {
	cx509CSR := &cx509.CertificateRequest{
		Version:            csr.Version,
		Signature:          nil,
		SignatureAlgorithm: csr.SignatureAlgorithm,

		PublicKeyAlgorithm: csr.PublicKeyAlg,
		DNSNames:           csr.DNSNames,
		EmailAddresses:     csr.EmailAddresses,
		IPAddresses:        csr.IPAddresses,
	}
	cx509CSR.Subject.Country = csr.SubjectCountry
	cx509CSR.Subject.Province = csr.SubjectProvince
	cx509CSR.Subject.StreetAddress = csr.SubjectStreetAddress
	cx509CSR.Subject.PostalCode = csr.SubjectPostalCode
	cx509CSR.Subject.Locality = csr.SubjectLocality
	cx509CSR.Subject.Organization = csr.SubjectOrganization
	cx509CSR.Subject.OrganizationalUnit = csr.SubjectOrganizationalUnit

	for _, uri := range csr.URIs {
		cx509CSR.URIs = append(cx509CSR.URIs, &uri)
	}

	for _, ex := range csr.Extensions {
		cx509CSR.Extensions = append(cx509CSR.Extensions, pkix.Extension{
			Id:       ex.ID,
			Critical: ex.Critical,
			Value:    ex.Value,
		})
	}

	buf, err := cx509.CreateCertificateRequest(rand.Reader, cx509CSR, signer)
	if err != nil {
		log.Fatal("error when create csr")
	}
	cx509CSR, err = cx509.ParseCertificateRequest(buf)
	if err != nil {
		log.Fatal("error when parse x50 CSR")
	}

	return cx509CSR
}
