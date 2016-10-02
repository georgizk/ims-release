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

	address := "0.0.0.0:3000"
	fmt.Printf("Listening on %s\n", address)
	http.ListenAndServe(address, router)
}
