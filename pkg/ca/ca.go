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
	"os"
	"time"

	"github.com/youmark/pkcs8"
)

const (
	rsaPrivateKeyLocation string = rootCAFolder + "/root.private.key"
	//rsaPrivateKeyPassword string = "123456"
	rootCALocation string = rootCAFolder + "/root.crt"

	localKeyLocation        string = localCAFolder + "/local.private.key"
	localCertLocation       string = localCAFolder + "/local.crt"
	localPrivateKeyPassword string = "123456"

	rootCAFolder   string = "cert/rootCA"
	clientCAFolder string = "cert/clientCert"
	localCAFolder  string = "cert/localCert"
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
	//如果没有配置根证书，我们自签一个
	if !checkFileExist(rootCALocation) || !checkFileExist(rsaPrivateKeyLocation) {
		if err := ca.makeRootCA(); err != nil {
			log.Print("can't create self-signed root CA")
			return
		}
		//我们需要同时签发本地server的certificate，用于后续的mTLS
		os.Remove(localCertLocation)
		os.Remove(localKeyLocation)
	}

	//加载 rootCA 的 private key
	bytes, err := ioutil.ReadFile(rsaPrivateKeyLocation)
	if err != nil {
		panic("can't load ca private key")
	}
	pemBlocks, _ := pem.Decode(bytes)
	if pemBlocks.Type != "ENCRYPTED PRIVATE KEY" {
		panic("ca private key type should be ENCRYPTED")
	}
	data, err := pkcs8.ParsePKCS8PrivateKeyRSA(pemBlocks.Bytes) //need package pkcs8 to parse
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

	//我们检查是否需要生成本地server的certificate
	if !checkFileExist(localCertLocation) || !checkFileExist(localKeyLocation) {
		if err := ca.signLocalCert(); err != nil {
			log.Print("can't create local certificate")
			return
		}
	}
}

/*
CA 做一个自签名证书，作为自己的根证书，当配置没有在cert\rootCA下提供根证书和私钥时，我们就自己做一个
*/
func (ca *CertificateAuthority) makeRootCA() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Print("error happens when generate private key to create root CA")
		return err
	}

	mathRand.Seed(time.Now().UnixNano())
	rootCertificateTemplate := cx509.Certificate{
		Version:      1,
		SerialNumber: big.NewInt((int64)(mathRand.Int())),
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Organization:       []string{"Fudan"},
			OrganizationalUnit: []string{"Mathematics"},
			Locality:           []string{"上海"},
			Province:           []string{"Shanghai"},
			StreetAddress:      []string{"Handan Road #200"},
			PostalCode:         []string{"200201"},
			CommonName:         "Fudan CA",
		},

		EmailAddresses: []string{"jacky01.zhang@outlook.com"},
		DNSNames:       []string{"localhost"},

		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	buf, err := cx509.CreateCertificate(rand.Reader, &rootCertificateTemplate, &rootCertificateTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Print("sign the root ca fail")
		return err
	}
	err = saveToPEM(buf, rootCAFolder, "root.crt", "CERTIFICATE")
	if err != nil {
		log.Print("persistent the root ca fail")
		return err
	}

	buf, err = pkcs8.MarshalPrivateKey(privateKey, nil, nil)
	if err != nil {
		log.Print("marshal ca private key fail")
		return err
	}
	err = saveToPEM(buf, rootCAFolder, "root.private.key", "ENCRYPTED PRIVATE KEY")
	if err != nil {
		log.Print("persistent the root ca private key fail")
		return err
	}

	return nil
}

/*
用CA自己的根证书，为自己签发一张证书，用于和客户端做mTLS
*/
func (ca *CertificateAuthority) signLocalCert() error {
	csr := &CertificateSigningRequest{
		SubjectCountry:            []string{"China"},
		SubjectOrganization:       []string{"Fudan"},
		SubjectOrganizationalUnit: []string{"ComputerScience"},
		SubjectProvince:           []string{"Shanghai"},
		SubjectLocality:           []string{"上海"},

		SubjectCommonName: "localhost", //这里需要填写所在Server的实际域名，但我们这里没有
		EmailAddresses:    []string{"jacky01.zhang@outlook.com"},
		DNSNames:          []string{"localhost"},
	}

	cert, err := ca.SignX509(csr)
	if err != nil {
		return err
	}

	err = os.Rename(clientCAFolder+"/"+cert.ID+".crt", localCertLocation)
	if err != nil {
		log.Print("move local cert file fail")
		return err
	}
	err = os.Rename(clientCAFolder+"/"+cert.ID+".key", localKeyLocation)
	if err != nil {
		log.Print("move local key file fail")
		return err
	}

	return nil
}

/*
用根证书签署一个证书签发请求CSR。CSR是以我自己的Struct表达的
*/
func (ca *CertificateAuthority) SignX509(csr *CertificateSigningRequest) (*Certificate, error) {

	csrPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Print("error happens when generate private key to sign CSR")
		return nil, err
	}

	cx509CSR := csr.toCX509CSR(csrPrivateKey)

	mathRand.Seed(time.Now().UnixNano())
	cx509CertificateTemplate := cx509.Certificate{
		Version:            cx509CSR.Version,
		SerialNumber:       big.NewInt((int64)(mathRand.Int())),
		Signature:          cx509CSR.Signature,
		SignatureAlgorithm: cx509CSR.SignatureAlgorithm,
		PublicKey:          cx509CSR.PublicKey,
		PublicKeyAlgorithm: cx509CSR.PublicKeyAlgorithm,
		Subject:            cx509CSR.Subject,

		URIs:           cx509CSR.URIs,
		DNSNames:       cx509CSR.DNSNames,
		EmailAddresses: cx509CSR.EmailAddresses,
		IPAddresses:    cx509CSR.IPAddresses,

		Extensions: cx509CSR.Extensions,

		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		BasicConstraintsValid: true,
	}

	buf, err := cx509.CreateCertificate(rand.Reader, &cx509CertificateTemplate, &ca.RootCA, cx509CSR.PublicKey, ca.PrivateKey)
	if err != nil {
		log.Print("sign the x509 csr fail")
		return nil, err
	}
	_, err = cx509.ParseCertificate(buf) //just to verify the generated byte[] is a qualified certification
	if err != nil {
		log.Print("verify the cx509 certificate fail")
		return nil, err
	}

	var fileNamePrefix string = time.Now().Format("2006-01-02_15-04-05")
	err = saveToPEM(buf, clientCAFolder, fileNamePrefix+".crt", "CERTIFICATE")
	if err != nil {
		log.Print("persistent generated certificate file fail")
		return nil, err
	}
	err = saveToPEM(cx509.MarshalPKCS1PrivateKey(csrPrivateKey), clientCAFolder, fileNamePrefix+".key", "PRIVATE KEY")
	if err != nil {
		log.Print("persistent generated private key file fail")
		return nil, err
	}

	res := &Certificate{ID: fileNamePrefix}
	return res, err
}

/*
return the generated client certificate file
*/
func (ca *CertificateAuthority) GetCertFile(id string) ([]byte, error) {
	contents, err := os.ReadFile(clientCAFolder + "/" + id + ".crt")
	if err != nil {
		return nil, err
	}
	return contents, nil
}

/*
return the generated client key file
*/
func (ca *CertificateAuthority) GetKeyFile(id string) ([]byte, error) {
	contents, err := os.ReadFile(clientCAFolder + "/" + id + ".key")
	if err != nil {
		return nil, err
	}
	return contents, nil
}

/*
把以我的Struct表述的 CSR 转化为 x509 package 定义的 CSR
x509包支持的csr属性都在这里了，不支持的没有包含
*/
func (csr *CertificateSigningRequest) toCX509CSR(signer crypto.Signer) *cx509.CertificateRequest {
	cx509CSR := &cx509.CertificateRequest{
		SignatureAlgorithm: csr.SignatureAlgorithm,

		//一下属性，加上uris，会形成subject alternative names
		DNSNames:       csr.DNSNames,
		EmailAddresses: csr.EmailAddresses,
		IPAddresses:    csr.IPAddresses,
	}
	for _, uri := range csr.URIs {
		cx509CSR.URIs = append(cx509CSR.URIs, &uri)
	}

	cx509CSR.Subject.CommonName = csr.SubjectCommonName
	cx509CSR.Subject.Country = csr.SubjectCountry
	cx509CSR.Subject.Province = csr.SubjectProvince
	cx509CSR.Subject.StreetAddress = csr.SubjectStreetAddress
	cx509CSR.Subject.PostalCode = csr.SubjectPostalCode
	cx509CSR.Subject.Locality = csr.SubjectLocality
	cx509CSR.Subject.Organization = csr.SubjectOrganization
	cx509CSR.Subject.OrganizationalUnit = csr.SubjectOrganizationalUnit

	for _, ex := range csr.Extensions {
		cx509CSR.Extensions = append(cx509CSR.Extensions, pkix.Extension{
			Id:       ex.ID,
			Critical: ex.Critical,
			Value:    ex.Value,
		})
	}

	buf, err := cx509.CreateCertificateRequest(rand.Reader, cx509CSR, signer)
	if err != nil {
		log.Print("error when create csr")
		return nil
	}
	cx509CSR, err = cx509.ParseCertificateRequest(buf)
	if err != nil {
		log.Print("error when parse x50 CSR")
		return nil
	}

	return cx509CSR
}

/*
保存一个私钥或证书到PEM文件
*/
func saveToPEM(bytes []byte, folder string, fileName string, fileType string) error {
	file, err := os.Create(folder + "/" + fileName)
	defer file.Close()
	if err != nil {
		return err
	}

	pemBlocks := &pem.Block{Bytes: bytes, Type: fileType}
	return pem.Encode(file, pemBlocks)
}

func checkFileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else {
		return os.IsExist(err)
	}
}
