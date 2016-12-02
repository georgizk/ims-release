package main

import (
	"./config"
	"./endpoints"
	"./models"

	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	cfgPath := os.Getenv("CONFIG_FILE")
	if cfgPath == "" {
		cfgPath = "./config/config.json"
	}
	cfg := config.MustLoad(cfgPath)

	db, err := sql.Open("mysql", cfg.Database)
	if err != nil {
		panic(err)
	}

	initErr := models.InitDB(db)
	if initErr != nil {
		panic(initErr)
	}

	router := mux.NewRouter()
	projectsRouter := router.PathPrefix("/projects").Subrouter()
	endpoints.RegisterProjectHandlers(projectsRouter, db, &cfg)
	releasesRouter := router.PathPrefix("/projects/{projectId}/releases").Subrouter()
	endpoints.RegisterReleaseHandlers(releasesRouter, db, &cfg)
	pagesRouter := router.PathPrefix("/projects/{projectId}/releases/{releaseId}/pages").Subrouter()
	endpoints.RegisterPageHandlers(pagesRouter, db, &cfg)

	// Should match /{projectName} - {chapter}[{version}]/{page}.{ext}
	router.HandleFunc("/{pc:\\w+\\s-\\s\\w+\\[\\d+\\]}/{page:\\w+\\.\\w+}", endpoints.DownloadImage(db, &cfg)).Methods("GET")
	// Should match /{projectName} - {chapter}[{version}][{groupName}].zip
	router.HandleFunc("/{path:\\w+\\s-\\s\\w+\\[\\d+\\]\\[\\w+\\]\\.zip}", endpoints.DownloadArchive(db, &cfg)).Methods("GET")

	address := cfg.BindAddress
	fmt.Printf("Listening on %s\n", address)
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	http.ListenAndServe(address, loggedRouter)
}
