package endpoints

import (
	"ims-release/config"
	"ims-release/database"
	"ims-release/models"
	"ims-release/storage_provider"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// RegisterProjectHandlers attaches the closures generated by each function defined below
// to handle incoming requests to the appropriate endpoint using a subrouter with an
// appropriate prefix, specified in main.
func RegisterProjectHandlers(r *mux.Router, db database.DB, cfg *config.Config, sp storage_provider.Binary) {
	root := "/projects"
	sr := r.PathPrefix(root).Subrouter()
	r.HandleFunc(root, listProjects(db, cfg)).Methods("GET")
	r.HandleFunc(root, createProject(db, cfg)).Methods("POST")
	sr.HandleFunc("/{projectId:[0-9]+}", getProject(db, cfg)).Methods("GET")
	sr.HandleFunc("/{projectId:[0-9]+}", updateProject(db, cfg)).Methods("PUT")
	sr.HandleFunc("/{projectId:[0-9]+}", deleteProject(db, cfg)).Methods("DELETE")
}

// GET /projects
type listProjectsResponse struct {
	Error    *string          `json:"error"`
	Projects []models.Project `json:"projects"`
}

// listProjects produces a list of all projects. It accepts an "ordering"
// parameter, which can be either "newest" or "oldest" to specify which should
// come first.
func listProjects(db database.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		projects, listErr := models.ListProjects(db)
		if listErr != nil {
			log.Println("[---] Listing error:", listErr)
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := "Could not obtain a list of projects. Please try again later."
			encoder.Encode(listProjectsResponse{&errMsg, []models.Project{}})
			return
		}
		encoder.Encode(listProjectsResponse{nil, projects})
	}
}

// POST /projects

type createProjectRequest struct {
	Name        string `json:"name"`
	ProjectName string `json:"projectName"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type createProjectResponse struct {
	Error   *string `json:"error"`
	Success bool    `json:"success"`
	Id      uint32  `json:"id"`
}

// createProject creates a new project.
func createProject(db database.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := createProjectRequest{}
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		decodeErr := decoder.Decode(&request)

		encoder := json.NewEncoder(w)
		if decodeErr != nil {
			log.Println("[---] Decode error:", decodeErr)
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "JSON format error or missing field detected."
			encoder.Encode(createProjectResponse{&errMsg, false, 0})
			return
		}
		project := models.NewProject(request.Name, request.ProjectName, request.Description, request.Status, time.Now())
		insertErr := project.Save(db)
		if insertErr != nil {
			log.Println("[---] Insert error:", insertErr)
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := "Could not create project. Try again later, or with a different projectName."
			encoder.Encode(createProjectResponse{&errMsg, false, 0})
			return
		}
		encoder.Encode(createProjectResponse{nil, true, project.Id})
	}
}

// GET /projects/{projectId}

type getProjectRequest struct {
	Id uint32
}

type getProjectResponse struct {
	Error       *string   `json:"error"`
	Name        string    `json:"name"`
	ProjectName string    `json:"projectName"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

// getProject obtains information about a specific project.
func getProject(db database.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := getProjectRequest{}
		vars := mux.Vars(r)
		_, parseErr := fmt.Sscanf(vars["projectId"], "%d", &request.Id)
		encoder := json.NewEncoder(w)
		if parseErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "projectId must be an integer ID."
			encoder.Encode(getProjectResponse{&errMsg, "", "", models.PStatusUnknownStr, "", time.Now()})
			return
		}
		project, findErr := models.FindProject(db, request.Id)
		if findErr != nil {
			log.Println("[---] Find error:", findErr)
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := "Could not find the requested project."
			encoder.Encode(getProjectResponse{&errMsg, "", "", models.PStatusUnknownStr, "", time.Now()})
			return
		}
		encoder.Encode(getProjectResponse{
			nil,
			project.Name,
			project.Shorthand,
			project.Status,
			project.Description,
			project.CreatedAt,
		})
	}
}

// PUT /projects/{projectId}

type updateProjectRequest struct {
	Id          uint32 // Provided from a URL path parameter
	Name        string `json:"name"`
	Shorthand   string `json:"projectName"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type updateProjectResponse struct {
	Error   *string `json:"error"`
	Success bool    `json:"success"`
}

// updateProject updates every field of an existing project with some supplied data.
func updateProject(db database.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := updateProjectRequest{}
		vars := mux.Vars(r)
		_, parseErr := fmt.Sscanf(vars["projectId"], "%d", &request.Id)
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		decodeErr := decoder.Decode(&request)

		encoder := json.NewEncoder(w)
		if parseErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "projectId must be an integer ID."
			encoder.Encode(updateProjectResponse{&errMsg, false})
			return
		}
		if decodeErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "JSON format error or missing field detected."
			encoder.Encode(updateProjectResponse{&errMsg, false})
			return
		}

		project, findErr := models.FindProject(db, request.Id)
		if findErr != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		project.Name = request.Name
		project.Status = request.Status
		project.Shorthand = request.Shorthand
		project.Description = request.Description

		updateErr := project.Update(db)
		if updateErr != nil {
			log.Println("[---] Update error:", updateErr)
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "Could not update specified project. Please ensure the ID and status are correct."
			encoder.Encode(updateProjectResponse{&errMsg, false})
			return
		}
		encoder.Encode(updateProjectResponse{nil, true})
	}
}

// DELETE /projects/{projectId}

type deleteProjectRequest struct {
	Id uint32
}

type deleteProjectResponse struct {
	Error   *string `json:"error"`
	Success bool    `json:"success"`
}

// deleteProject removes an entire project from the database, along with
// all of the releases under that project.
func deleteProject(db database.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := deleteProjectRequest{}
		vars := mux.Vars(r)
		_, parseErr := fmt.Sscanf(vars["projectId"], "%d", &request.Id)
		encoder := json.NewEncoder(w)
		if parseErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "projectId must be an integer ID."
			encoder.Encode(deleteProjectResponse{&errMsg, false})
			return
		}

		project, err := models.FindProject(db, request.Id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		releases, err := models.ListReleases(db, project)
		if err != nil {
			log.Println("[---] Delete error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := "Unexpected error."
			encoder.Encode(deleteProjectResponse{&errMsg, false})
			return
		}

		if len(releases) > 0 {
			log.Println("[---] Delete error: releases not empty")
			w.WriteHeader(http.StatusExpectationFailed)
			errMsg := "All releases must be deleted before deleting a project."
			encoder.Encode(deleteProjectResponse{&errMsg, false})
			return
		}

		deleteErr := project.Delete(db)
		if deleteErr != nil {
			log.Println("[---] Delete error:", deleteErr)
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := "Could not delete the specified project. Please check that the ID is valid."
			encoder.Encode(deleteProjectResponse{&errMsg, false})
			return
		}
		encoder.Encode(deleteProjectResponse{nil, true})
	}
}
