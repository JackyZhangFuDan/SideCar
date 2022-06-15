package ca

import (
	"crypto/x509"
)

type CertificateAuthority struct {
}

/*
sign a csr
*/
func (ca *CertificateAuthority) sign() (Certificate, error) {
	x509.CreateCertificate()
	panic("isn't implemented")
}
