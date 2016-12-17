package main

import (
	"./config"
	"./endpoints"
	"./storage_provider"

	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/DavidHuie/gomigrate"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
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

	Migrate(db)
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

func Migrate(db *sql.DB) error {
	migrator, _ := gomigrate.NewMigrator(db, gomigrate.Mysql{}, "./migrations")
	err := migrator.Migrate()
	return err
}
