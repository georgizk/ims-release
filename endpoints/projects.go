package endpoints

import (
	"./config"
	"./models"
	"database/sql"
	"net/http"
)

// RegisterProjectHandlers attaches the closures generated by each function defined below
// to handle incoming requests to the appropriate endpoint using a subrouter with an
// appropriate prefix, specified in main.
func RegisterProjectHandlers(r *mux.Router, db *sql.DB, cfg *config.Config) {
	r.HandleFunc("/projects", ListProjects(db, cfg)).Methods("GET")
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
	return func(w http.ResponseWriter, r *http.Response) {
		w.Header().Set("Content-Type", "application/json")
		request := listProjectsRequest{}
		ordering, found := r.URL.Query()["ordering"]
		if !found {
			ordering = "newest"
		}
		request.Ordering = ordering

		encoder := json.NewEncoder(w)
		// TODO - List projects from the database
		encoder.Encode(listProjectsResponse{nil, []models.Project{}})
	}
}
