package ca

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	cx509 "crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	mathRand "math/rand"
	"time"

	"github.com/youmark/pkcs8"
)

const (
	rsaPrivateKeyLocation string = "cert/private/ca.private.key"
	rsaPrivateKeyPassword string = "123456"
	rootCALocation        string = "cert/ca/root.crt"
)

var CA CertificateAuthority

func init() {
	CA = CertificateAuthority{}
	CA.load()
}

type CertificateAuthority struct {
	RootCA     cx509.Certificate
	PrivateKey *rsa.PrivateKey
}

/*
从磁盘加载根证书和私钥信息
*/
func (ca *CertificateAuthority) load() {
	//加载 rootCA 的 private key
	bytes, err := ioutil.ReadFile(rsaPrivateKeyLocation)
	if err != nil {
		panic("can't load ca private key")
	}
	pemBlocks, _ := pem.Decode(bytes)
	if pemBlocks.Type != "ENCRYPTED PRIVATE KEY" {
		panic("ca private key type should be ENCRYPTED")
	}
	data, err := pkcs8.ParsePKCS8PrivateKeyRSA(pemBlocks.Bytes, []byte(rsaPrivateKeyPassword)) //need package pkcs8 to parse
	if err != nil {
		panic("can't parse private key bytes via pkcs8")
	}
	ca.PrivateKey = data
	//加载 rootCA
	rootCABytes, err := ioutil.ReadFile(rootCALocation)
	if err != nil {
		panic("can't load root ca")
	}
	pemBlocks, _ = pem.Decode(rootCABytes)
	rootCA, err := cx509.ParseCertificate(pemBlocks.Bytes)
	if err != nil {
		panic("can't parse root ca")
	}
	ca.RootCA = *rootCA
}

/*
用根证书签署一个证书签发请求CSR。CSR是以我自己的Struct表达的
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
把以我的Struct表述的 CSR 转化为 x509 package 定义的 CSR
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
