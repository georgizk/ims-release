package main

import (
	"./endpoints"

	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	projectsRouter := router.PathPrefix("/projects").Subrouter()
	endpoints.RegisterProjectHandlers(projectsRouter, nil, nil)

	address := "0.0.0.0:3000"
	fmt.Printf("Listening on %s\n", address)
	http.ListenAndServe(address, router)
}
