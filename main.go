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
	cfgPath := ""
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	} else {
		fmt.Fprintf(os.Stderr, "Usage: ims-release <configPath>\n")
		fmt.Fprintf(os.Stderr, "You must specify the path to the json configuration file.\n")
		os.Exit(1)
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
	router.StrictSlash(true)
	endpoints.RegisterProjectHandlers(router, db, &cfg)
	endpoints.RegisterReleaseHandlers(router, db, &cfg)
	endpoints.RegisterPageHandlers(router, db, &cfg)
	endpoints.RegisterDownloadHandlers(router, db, &cfg)

	address := cfg.BindAddress
	fmt.Printf("Listening on %s\n", address)
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	http.ListenAndServe(address, loggedRouter)
}
