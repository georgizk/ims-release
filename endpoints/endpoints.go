package endpoints

import (
	"ims-release/config"
	"ims-release/database"
	"ims-release/storage_provider"

	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func NewHttpHandler(cfg *config.Config) http.Handler {
	db, err := database.NewDbHandle(cfg)
	if err != nil {
		panic(err)
	}
	db.Migrate(os.Getenv("GOPATH") + "/src/ims-release/migrations")
	router := mux.NewRouter()
	sp := storage_provider.File{Root: cfg.ImageDirectory}
	registerHandlers(router, db, cfg, &sp)

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	corsRouter := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}))(loggedRouter)

	return corsRouter
}

func registerHandlers(r *mux.Router, db database.DB, cfg *config.Config, sp storage_provider.Binary) {
	r.StrictSlash(true)
	RegisterProjectHandlers(r, db, cfg, sp)
	RegisterReleaseHandlers(r, db, cfg, sp)
	RegisterPageHandlers(r, db, cfg, sp)
}
