package endpoints

import (
	"../config"
	"../models"

	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// TODO
// - If loading images into memory to serve them becomes too much of a burden (and it may well)
//   then we should switch to a means of streaming the contents of the file into the HTTP response.

// Error types pertaining to download requests.
var (
	ErrInvalidURLFormat = errors.New("The URL you requested is not formatted correctly and appears to be missing data.")
)

// GET /{projectName}-{chapter}.{version}/{page}.{ext}

type getPageRequest struct {
	ProjectName string
	Chapter     string
	Version     int
	Page        string
}

// Attempts to parse all of the parameters out of a DownloadImage request from the
// url requested to download a page.
func parseDownloadImageRequest(pac, pnum string) (getPageRequest, error) {
	req := getPageRequest{}

	// Expect pac (page and chapter section) to be formatted {projectName}-{chapter}.{version}
	parts := strings.Split(pac, ".")
	if len(parts) != 2 {
		return getPageRequest{}, ErrInvalidURLFormat
	}
	version, parseErr := strconv.Atoi(parts[1])
	if parseErr != nil {
		return getPageRequest{}, ErrInvalidURLFormat
	}
	req.Version = version
	parts = strings.Split(parts[0], "-")
	if len(parts) != 2 {
		return getPageRequest{}, ErrInvalidURLFormat
	}
	req.ProjectName = parts[0]
	req.Chapter = parts[1]

	// Expect pnum (page number) to be formatted {pageNumber}.{ext}
	// We will ignore the extension.
	parts = strings.Split(pnum, ".")
	if len(parts) != 2 {
		return getPageRequest{}, ErrInvalidURLFormat
	}
	req.Page = parts[0]

	return req, nil
}

// getPage retrieves the contents of a page from disk.
func DownloadImage(db *sql.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		projectAndChapter := vars["pc"]
		pageNumber := vars["page"]
		request, parseErr := parseDownloadImageRequest(projectAndChapter, pageNumber)

		if parseErr != nil {
			fmt.Println("[---] Parse error: %v\n", parseErr)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Please ensure that each of the projectId, releaseId, and pageId parameters are valid integers."))
			return
		}
		page, findErr := models.LookupPage(request.Page, request.Chapter, request.Version, request.ProjectName, db)
		if findErr != nil {
			fmt.Println("[---] Find error:", findErr)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Could not find the requested page. Please ensure that the pageId is correct."))
			return
		}
		f, openErr := os.Open(page.Location)
		if openErr != nil {
			fmt.Println("[---] Open error:", openErr)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not read the page file. Please try again later."))
			return
		}
		imageBytes, readErr := ioutil.ReadAll(f)
		defer f.Close()
		if readErr != nil {
			fmt.Println("[---] Open error:", openErr)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not read the page file. Please try again later."))
			return
		}
		w.WriteHeader(http.StatusOK)
		if strings.HasSuffix(page.Location, "png") {
			w.Header().Set("Content-Type", "image/png")
		} else {
			w.Header().Set("Content-Type", "image/jpeg")
		}
		w.Write(imageBytes)
	}
}
