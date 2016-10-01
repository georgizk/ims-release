package endpoints

import (
	"../config"
	"../models"

	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// RegisterReleaseHandlers attaches the closures generated by each function defined below
// to handle incoming requests to the appropriate endpoint using a subrouter with an
// appropriate prefix, specified in main.
func RegisterReleaseHandlers(r *mux.Router, db *sql.DB, cfg *config.Config) {
	r.HandleFunc("/", listReleases(db, cfg)).Methods("GET")
	r.HandleFunc("/", createRelease(db, cfg)).Methods("POST")
	r.HandleFunc("/{releaseId}", getRelease(db, cfg)).Methods("GET")
	r.HandleFunc("/{releaseId}", updateRelease(db, cfg)).Methods("PUT")
	r.HandleFunc("/{releaseId}", deleteRelease(db, cfg)).Methods("DELETE")
}

// GET /projects/{projectId}/releases

type listReleasesRequest struct {
	ProjectID int
	Ordering  string
}

type listReleasesResponse struct {
	Error    *string          `json:"error"`
	Releases []models.Release `json:"releases"`
}

// listReleases produces a list of all releases under a given project.
// It accepts an "ordering" parameter, which can be either "newest" or "oldest"
// to specify which should come first.
func listReleases(db *sql.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		pid := mux.Vars(r)["projectId"]
		projectId, parseErr := strconv.Atoi(pid)
		request := listReleasesRequest{0, "newest"}
		orderings, found := r.URL.Query()["ordering"]
		if found {
			request.Ordering = orderings[0]
		}

		encoder := json.NewEncoder(w)
		if parseErr != nil {
			errMsg := "projectId must be an integer ID."
			encoder.Encode(listReleasesResponse{&errMsg, []models.Release{}})
			return
		}
		request.ProjectID = projectId
		// TODO - List all releases under the project from the database.
		encoder.Encode(listReleasesResponse{nil, []models.Release{}})
	}
}

// POST /projects/{projectId}/releases

type createReleaseRequest struct {
	Chapter string               `json:"chapter"`
	Version int                  `json:"version"`
	Status  models.ReleaseStatus `json:"status"`
}

type createReleaseResponse struct {
	Error   *string `json:"error"`
	Success bool    `json:"success"`
	Id      int     `json:"id"`
}

// createRelease inserts a new release into the database.
func createRelease(db *sql.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := createReleaseRequest{}
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		decodeErr := decoder.Decode(&request)

		encoder := json.NewEncoder(w)
		if decodeErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "JSON format error or missing field detected."
			encoder.Encode(createReleaseResponse{&errMsg, false, 0})
			return
		}
		// TODO - Insert the release into the DB and update its ID etc.
		encoder.Encode(createReleaseResponse{nil, true, 1})
	}
}

// GET /projects/{projectId}/releases/{releaseId}

type getReleaseRequest struct {
	ProjectID int
	ReleaseID int
}

type getReleaseResponse struct {
	Error       *string              `json:"error"`
	ProjectName string               `json:"projectName"`
	Chapter     string               `json:"chapter"`
	GroupName   string               `json:"groupName"`
	Checksum    string               `json:"checksum"`
	Version     int                  `json:"version"`
	Status      models.ReleaseStatus `json:"status"`
	ReleasedOn  time.Time            `json:"releasedOn"`
}

// getRelease obtains information about a specific release.
func getRelease(db *sql.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := getReleaseRequest{}
		vars := mux.Vars(r)
		pid := vars["projectId"]
		rid := vars["releaseId"]
		projectId, parseErr1 := strconv.Atoi(pid)
		releaseId, parseErr2 := strconv.Atoi(rid)

		encoder := json.NewEncoder(w)
		if parseErr1 != nil || parseErr2 != nil {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "projectId and releaseId must be integer IDs."
			encoder.Encode(getReleaseResponse{&errMsg, "", "", "", "", 0, "", time.Now()})
			return
		}
		request.ProjectID = projectId
		request.ReleaseID = releaseId
		// TODO - Retrieve the release from the DB.
		encoder.Encode(getReleaseResponse{
			nil,
			"project",
			"ch001",
			"ims",
			"abc123",
			1,
			models.RStatusReleased,
			time.Now(),
		})
	}
}

// PUT /projects/{projectId}/releases/{releaseId}

type updateReleaseRequest struct {
	ProjectID int                  // Pulled from the URL params
	ReleaseID int                  // Pulled from the URL params
	Chapter   string               `json:"chapter"`
	Version   int                  `json:"version"`
	Status    models.ReleaseStatus `json:"status"`
}

type updateReleaseResponse struct {
	Error   *string `json:"error"`
	Success bool    `json:"success"`
}

// updateRelease updates the chapter, version, and status of a release.
func updateRelease(db *sql.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := updateReleaseRequest{}
		vars := mux.Vars(r)
		pid := vars["projectId"]
		rid := vars["releaseId"]
		projectId, parseErr1 := strconv.Atoi(pid)
		releaseId, parseErr2 := strconv.Atoi(rid)
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		decodeErr := decoder.Decode(&request)

		encoder := json.NewEncoder(w)
		if parseErr1 != nil || parseErr2 != nil {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "projectId and releaseId must both be integer IDs."
			encoder.Encode(updateReleaseResponse{&errMsg, false})
			return
		}
		request.ProjectID = projectId
		request.ReleaseID = releaseId
		if decodeErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "Failed to decode JSON in the request body."
			encoder.Encode(updateReleaseResponse{&errMsg, false})
			return
		}
		// TODO - Update the release in the DB.
		encoder.Encode(updateReleaseResponse{nil, true})
	}
}

// DELETE /projects/{projectId}/releases/{releaseId}

type deleteReleaseRequest struct {
	ProjectID int
	ReleaseID int
}

type deleteReleaseResponse struct {
	Error   *string `json:"error"`
	Success bool    `json:"success"`
}

// deleteRelease deletes a release from the DB and also all associated pages.
func deleteRelease(db *sql.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := deleteReleaseRequest{}
		vars := mux.Vars(r)
		pid := vars["projectId"]
		rid := vars["releaseId"]
		projectId, parseErr1 := strconv.Atoi(pid)
		releaseId, parseErr2 := strconv.Atoi(rid)

		encoder := json.NewEncoder(w)
		if parseErr1 != nil || parseErr2 != nil {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "projectId and releaseId must both be integer IDs."
			encoder.Encode(deleteReleaseResponse{&errMsg, false})
			return
		}
		request.ProjectID = projectId
		request.ReleaseID = releaseId
		// TODO - Delete the release from the DB.
		encoder.Encode(deleteReleaseResponse{nil, true})
	}
}
