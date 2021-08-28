package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"

	"BenchmarkSitesForScraping/internal/benchmark"
)

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/sites", handleURL).Methods("GET")
	log.Printf("main: starting HTTP server")
	log.Fatal(http.ListenAndServe(":8000", router))

}

func handleURL(w http.ResponseWriter, r *http.Request) {

	url2Process := r.FormValue("search")

	if url2Process != "" {
		bm := benchmark.NewBenchmark(url2Process)
		bm.Process()
	}



}