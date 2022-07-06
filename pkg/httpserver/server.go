package httpserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jackyzhangfudan/sidecar/pkg/ca"
)

var server *http.Server
var running bool

func init() {
}

func Run() {
	if running {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/csr-template", getCsrTemplateHandler)
	mux.HandleFunc("/csr", signCsrHandler)
	server = &http.Server{
		Addr:    ":8111",
		Handler: mux,
	}

	running = true
	if server.ListenAndServe() != nil {
		running = false
		log.Fatal("can't start http server @ 8111")
	}
	running = false
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("error when reading body"))
		return
	}
	w.Write(body)
}

func getCsrTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	csr := ca.CertificateSigningRequest{
		SubjectCountry:            []string{"China", "US"},
		SubjectOrganization:       []string{"Qinghua", "Beida"},
		SubjectOrganizationalUnit: []string{"ComputerScience", "Mathematics"},
		SubjectProvince:           []string{"Shanghai"},
		SubjectLocality:           []string{"上海"},

		PublicKeyAlg:       1,
		SignatureAlgorithm: 4,
		SubjectCommonName:  "www.fudan.edu.cn",
		EmailAddresses:     []string{"ex@example.com"},
	}

	csrBytes, err := json.Marshal(csr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json")
	w.Write(csrBytes)
}

func signCsrHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	csr := &ca.CertificateSigningRequest{}
	err = json.Unmarshal(reqBody, csr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	go signCsrRoutine(w, csr)

}

func signCsrRoutine(w http.ResponseWriter, csr *ca.CertificateSigningRequest) {
	theCert, err := ca.CA.SignX509(csr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error happen: %v", err)
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "certificated is generated: %v", theCert.AuthorityKeyId)
}
