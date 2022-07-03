package httpserver

import (
	"log"
	"net/http"
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

}
