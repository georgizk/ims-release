package endpoints

import (
	"../config"
	"../models"

	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterProjectHandlers attaches the closures generated by each function defined below
// to handle incoming requests to the appropriate endpoint using a subrouter with an
// appropriate prefix, specified in main.
func RegisterProjectHandlers(r *mux.Router, db *sql.DB, cfg *config.Config) {
	r.HandleFunc("/", ListProjects(db, cfg)).Methods("GET")
}

// GET /projects

type listProjectsRequest struct {
	Ordering string
}

type listProjectsResponse struct {
	Error    *string          `json:"error"`
	Projects []models.Project `json:"projects"`
}

// ListProjects produces a list of all projects. It accepts
func ListProjects(db *sql.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := listProjectsRequest{}
		orderings, found := r.URL.Query()["ordering"]
		if !found {
			orderings = []string{"newest"}
		}
		request.Ordering = orderings[0]

		encoder := json.NewEncoder(w)
		// TODO - List projects from the database
		encoder.Encode(listProjectsResponse{nil, []models.Project{}})
	}
}
