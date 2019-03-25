package main

import (
	"fmt"
	"log"
	"net/http"
)

const PublicPort string = "80"

var publicServer *http.ServeMux

func startListeningPublic(exitCallback chan bool) {
	publicServer := http.NewServeMux()

	publicServer.Handle("/", http.FileServer(http.Dir("./www/public")))

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%s", PublicPort), publicServer); err != nil {
			panic(err)
		}
		exitCallback <- true
	}()

	log.Printf("Public server started on port %s", PublicPort)
}
