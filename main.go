package main

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Println("booting server...")

	r := mux.NewRouter()

	r.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		log.Println("GET /")

		rw.WriteHeader(http.StatusOK)
		return
	})

	http.ListenAndServe(":9001", r)
}
