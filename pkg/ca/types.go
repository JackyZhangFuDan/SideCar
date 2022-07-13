package ca

import (
	"crypto/x509"
	"encoding/asn1"
	"net"
	"net/url"
)

/*
contain CSR from http request
*/
type CertificateSigningRequest struct {
	Version int

	SubjectCountry            []string //subject attributes
	SubjectOrganization       []string
	SubjectOrganizationalUnit []string
	SubjectLocality           []string
	SubjectProvince           []string
	SubjectStreetAddress      []string
	SubjectPostalCode         []string
	SubjectSerialNumber       string
	SubjectCommonName         string
	SubjectExtraNames         []DistinguishedName

	PublicKeyAlg       x509.PublicKeyAlgorithm //public key的生成算法：rsa，ecdsa，dsa
	SignatureAlgorithm x509.SignatureAlgorithm //签名算法，int

	//一下四个属性会构成证书中的SANs
	DNSNames       []string
	EmailAddresses []string
	IPAddresses    []net.IP
	URIs           []url.URL

	Extensions []Extension
}

type DistinguishedName struct {
	Type  asn1.ObjectIdentifier
	Value interface{}
}

type Extension struct {
	ID       asn1.ObjectIdentifier
	Critical bool
	Value    []byte
}

type Certificate struct {
	ID string `json:"certificateId" `
}
