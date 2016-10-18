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

// GET /{projectName}-{chapter}{groupName}{checksum}.{version}.zip
type getArchiveRequest struct {
	ProjectName string
	Chapter     string
	GroupName   string
	Checksum    string
	Version     int
}

// parseDownloadArchiveRequest attempts to parse all of the parameters out of a DownloadArchive
// request from the URL requested to download an archive.
func parseDownloadArchiveRequest(path string) (getArchiveRequest, error) {
	req := getArchiveRequest{}

	// Expect the url to be formatted {projectName} - {chapter}[{version}][{groupName}].zip
	parts := strings.Split(path, "-")
	if len(parts) != 2 {
		return getArchiveRequest{}, ErrInvalidURLFormat
	}
	req.ProjectName = strings.Trim(parts[0], " ")
	parts = strings.Split(parts[1], "[")
	if len(parts) != 3 {
		return getArchiveRequest{}, ErrInvalidURLFormat
	}
	req.Chapter = strings.Trim(parts[0], " ")
	version, parseErr := strconv.Atoi(strings.Trim(parts[1], "]"))
	if parseErr != nil {
		return getArchiveRequest{}, parseErr
	}
	req.Version = version
	parts = strings.Split(parts[2], ".")
	if len(parts) != 2 {
		return getArchiveRequest{}, ErrInvalidURLFormat
	}
	req.GroupName = strings.Trim(parts[0], "]")

	return req, nil
}

// DownloadArchive prepares and downloads the latest version of an archive for a particular release.
func DownloadArchive(db *sql.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, parseErr := parseDownloadArchiveRequest(mux.Vars(r)["path"])
		if parseErr != nil {
			fmt.Println("[---] Parse error:", parseErr)
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "Could not parse all of the required parameters from the URL."
			w.Write([]byte(errMsg))
			return
		}
		release, lookupErr := models.LookupRelease(request.Chapter, request.Version, request.Checksum, request.ProjectName, db)
		if lookupErr != nil {
			fmt.Println("[---] Lookup error:", lookupErr)
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "Could not lookup requested archive. Please check that the file format is correct or try again later."
			w.Write([]byte(errMsg))
			return
		}
		archive, buildErr := release.CreateArchive(db)
		if buildErr != nil {
			fmt.Println("[---] Build error:", buildErr)
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := "Could not produce an archive for the release requested. Please try again later."
			w.Write([]byte(errMsg))
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Write(archive)
	}
}

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

// DownloadImage retrieves the contents of a page from disk.
func DownloadImage(db *sql.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		projectAndChapter := vars["pc"]
		pageNumber := vars["page"]
		request, parseErr := parseDownloadImageRequest(projectAndChapter, pageNumber)

		if parseErr != nil {
			fmt.Println("[---] Parse error: %v\n", parseErr)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Could not parse all of the parameters required from the URL."))
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
