package main

import (
	"ims-release/config"
	"ims-release/endpoints"
	"ims-release/storage_provider"
	"ims-release/database"

	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	const usage = "Usage: ims-release <configPath>\n"
	const missingConf = "You must specify the path to the json configuration file.\n"
	log.SetFlags(log.LstdFlags | log.Llongfile)
	cfgPath := ""
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	} else {
		fmt.Fprintf(os.Stderr, usage)
		fmt.Fprintf(os.Stderr, missingConf)
		os.Exit(1)
	}
	cfg := config.MustLoad(cfgPath)

	db, err := database.NewDbHandle(&cfg)
	if err != nil {
		panic(err)
	}
	migrationsPath := os.Getenv("GOPATH") + "/src/ims-release/migrations"
	db.Migrate(migrationsPath)
	sp := storage_provider.File{Root: cfg.ImageDirectory}

	router := mux.NewRouter()
	router.StrictSlash(true)
	endpoints.RegisterProjectHandlers(router, db, &cfg, &sp)
	endpoints.RegisterReleaseHandlers(router, db, &cfg, &sp)
	endpoints.RegisterPageHandlers(router, db, &cfg, &sp)

	address := cfg.BindAddress
	log.Printf("Listening on %s\n", address)
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	corsRouter := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}))(loggedRouter)

	http.ListenAndServe(address, corsRouter)
}
