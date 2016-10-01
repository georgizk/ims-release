package main

import (
	"./endpoints"

	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// TODO - Initialize the database and load a configuration.

	router := mux.NewRouter()
	projectsRouter := router.PathPrefix("/projects").Subrouter()
	endpoints.RegisterProjectHandlers(projectsRouter, nil, nil)
	releasesRouter := router.PathPrefix("/projects/{projectId}/releases").Subrouter()
	endpoints.RegisterReleaseHandlers(releasesRouter, nil, nil)
	pagesRouter := router.PathPrefix("/projects/{projectId}/releases/{releaseId}/pages").Subrouter()
	endpoints.RegisterPageHandlers(pagesRouter, nil, nil)

	address := "0.0.0.0:3000"
	fmt.Printf("Listening on %s\n", address)
	http.ListenAndServe(address, router)
}
