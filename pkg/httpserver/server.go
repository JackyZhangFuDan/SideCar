package httpserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jackyzhangfudan/sidecar/pkg/ca"
)

const (
	port int32 = 8111
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
		Addr:    fmt.Sprintf(":%v", port),
		Handler: mux,
	}

	running = true
	fmt.Printf("server listening at %v, http", server.Addr)
	if server.ListenAndServe() != nil {
		running = false
		log.Printf("can't start http server at %v", server.Addr)
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
		SubjectCountry:            []string{"China"},
		SubjectOrganization:       []string{"Qinghua"},
		SubjectOrganizationalUnit: []string{"ComputerScience"},
		SubjectProvince:           []string{"Beijing"},
		SubjectLocality:           []string{"北京"},

		SubjectCommonName: "www.tsinghua.edu.cn",
		EmailAddresses:    []string{"ex@example.com"},
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

	var sync chan int = make(chan int, 1)
	go signCsrRoutine(w, csr, sync)

	<-sync
}

func signCsrRoutine(w http.ResponseWriter, csr *ca.CertificateSigningRequest, sync chan<- int) {
	defer close(sync)
	theCert, err := ca.CA.SignX509(csr)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error happen: %v", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Add("Content-Type", "application/json")
	jsonByte, _ := json.Marshal(theCert)
	w.Write(jsonByte)

}
