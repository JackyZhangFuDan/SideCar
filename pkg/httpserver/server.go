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
	mux.HandleFunc("/csr", signCsr)
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
	//w.Write([]byte("hello from Jacky"))

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("error when reading body"))
		return
	}
	w.Write(body)
}

func signCsr(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "received")
}
