package endpoints

import (
	"ims-release/database"
	"ims-release/models"
	"ims-release/storage_provider"

	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// RegisterPageHandlers attaches the closures generated by each function defined below
// to handle incoming requests to the appropriate endpoint using a subrouter with an
// appropriate prefix, specified in main.
func RegisterPageHandlers(r *mux.Router, db database.DB, sp storage_provider.Binary) {
	root := "/projects/{projectId:[0-9]+}/releases/{releaseId:[0-9]+}/pages"
	sr := r.PathPrefix(root).Subrouter()
	r.HandleFunc(root, listPages(db)).Methods("GET")
	r.HandleFunc(root, createPage(db, sp)).Methods("POST")
	sr.HandleFunc("/{pageId:[0-9]+}", deletePage(db, sp)).Methods("DELETE")
	sr.HandleFunc("/{name}", getPage(db, sp)).Methods("GET")
}

// GET /projects/{projectId}/releases/{releaseId}/pages

type listPagesRequest struct {
	ProjectID uint32
	ReleaseID uint32
}

type listPagesResponse struct {
	Error *string       `json:"error"`
	Pages []models.Page `json:"pages"`
}

// listPages lists descriptive information about
func listPages(db database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := listPagesRequest{}
		vars := mux.Vars(r)
		_, parseErr1 := fmt.Sscanf(vars["projectId"], "%d", &request.ProjectID)
		_, parseErr2 := fmt.Sscanf(vars["releaseId"], "%d", &request.ReleaseID)

		encoder := json.NewEncoder(w)
		if parseErr1 != nil || parseErr2 != nil {
			log.Printf("[---] Parse error: %v || %v\n", parseErr1, parseErr2)
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "projectId and releaseId must be integer IDs."
			encoder.Encode(listPagesResponse{&errMsg, []models.Page{}})
			return
		}

		project, err := models.FindProject(db, request.ProjectID)
		if err != nil {
			log.Println("unable to find project")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		release, findErr := models.FindRelease(db, project, request.ReleaseID)
		if findErr != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		pages, listErr := models.ListPages(db, release)
		if listErr != nil {
			log.Println("[---] List error:", listErr)
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := "Could not list pages. Please check that the projectId is correct or try again later."
			encoder.Encode(listPagesResponse{&errMsg, []models.Page{}})
			return
		}
		encoder.Encode(listPagesResponse{nil, pages})
	}
}

// POST /projects/{projectId}/releases/{releaseId}/pages

type createPageRequest struct {
	ProjectID uint32 // Pulled from the URL parameters
	ReleaseID uint32 // Pulled from the URL parameters
	Name      string `json:"name"`
	ImageData string `json:"data"`
}

type createPageResponse struct {
	Error   *string `json:"error"`
	Success bool    `json:"success"`
	Id      uint32  `json:"id"`
}

// createPage inserts a new page into the DB and saves page data to a file.
func createPage(db database.DB, sp storage_provider.Binary) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		request := createPageRequest{}
		vars := mux.Vars(r)
		_, parseErr1 := fmt.Sscanf(vars["projectId"], "%d", &request.ProjectID)
		_, parseErr2 := fmt.Sscanf(vars["releaseId"], "%d", &request.ReleaseID)

		const errMsgWrongType = "The uploaded image is neither a valid JPG/JPEG or PNG image."
		const errMsgMustBeInt = "projectId and releaseId must be integer IDs."
		const errMsgNoSuchRelease = "No such release."
		const errMsgJsonFormat = "JSON format error or missing field detected."
		const errMsgImageData = "The supplied image data is not base64 encoded."
		const errMsgSaveFailed = "Failed to save image file. Please try again later."

		if parseErr1 != nil || parseErr2 != nil {
			log.Printf("[---] Parse error: %v || %v\n", parseErr1, parseErr2)
			w.WriteHeader(http.StatusBadRequest)
			errStr := errMsgMustBeInt
			encoder.Encode(createPageResponse{&errStr, false, 0})
			return
		}

		project, err := models.FindProject(db, request.ProjectID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			errStr := errMsgNoSuchRelease
			encoder.Encode(createPageResponse{&errStr, false, 0})
			return
		}

		release, err := models.FindRelease(db, project, request.ReleaseID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			errStr := errMsgNoSuchRelease
			encoder.Encode(createPageResponse{&errStr, false, 0})
			return
		}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		decodeErr := decoder.Decode(&request)
		log.Printf("[+++] Project ID = %d, Release ID = %d\n", request.ProjectID, request.ReleaseID)

		if decodeErr != nil {
			log.Println("[---] Decode error:", decodeErr)
			w.WriteHeader(http.StatusBadRequest)
			errStr := errMsgJsonFormat
			encoder.Encode(createPageResponse{&errStr, false, 0})
			return
		}

		imageData, decodeErr := base64.StdEncoding.DecodeString(request.ImageData)
		if decodeErr != nil {
			log.Println("[---] Image decode error:", decodeErr)
			w.WriteHeader(http.StatusBadRequest)
			errStr := errMsgImageData
			encoder.Encode(createPageResponse{&errStr, false, 0})
			return
		}
		log.Println("[+++] Successfully decoded image data")

		page := models.NewPage(release, request.Name, time.Now())
		mimeType := page.MimeType
		var imgParseErr error
		switch mimeType {
		case models.MimeTypePng:
			_, imgParseErr = png.Decode(bytes.NewReader(imageData))
			break
		case models.MimeTypeJpg:
			_, imgParseErr = jpeg.Decode(bytes.NewReader(imageData))
			break
		case models.MimeTypeUnknown:
		default:
			imgParseErr = errors.New("bad extension")
			break
		}

		if imgParseErr != nil {
			// The image is neither a valid JPG/JPEG nor a valid PNG image.
			log.Printf("[---] Uploaded error: %v\n", imgParseErr)
			w.WriteHeader(http.StatusBadRequest)
			errStr := errMsgWrongType
			encoder.Encode(createPageResponse{&errStr, false, 0})
			return
		}

		filePath := models.GeneratePagePath(project, release, page.Name)

		log.Printf("[+++] Computed filename %s\n", filePath)
		saveErr := sp.Set(filePath, imageData)
		if saveErr != nil {
			log.Println("[---] Save error:", saveErr)
			w.WriteHeader(http.StatusInternalServerError)
			errStr := errMsgSaveFailed
			encoder.Encode(createPageResponse{&errStr, false, 0})
			return
		}
		log.Println("[+++] Successfully saved image to disk")

		saveErr = page.Save(db)
		if saveErr != nil {
			log.Println("[---] Insert error:", saveErr)
			w.WriteHeader(http.StatusInternalServerError)
			errStr := errMsgSaveFailed
			encoder.Encode(createPageResponse{&errStr, false, 0})
			sp.Unset(filePath)
			return
		}
		encoder.Encode(createPageResponse{nil, true, page.Id})
	}
}

func getPage(db database.DB, sp storage_provider.Binary) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		var projectId uint32
		var releaseId uint32
		_, parseErr1 := fmt.Sscanf(vars["projectId"], "%d", &projectId)
		_, parseErr2 := fmt.Sscanf(vars["releaseId"], "%d", &releaseId)

		if parseErr1 != nil || parseErr2 != nil {
			log.Println("[---] Parse error")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		project, findErr := models.FindProject(db, projectId)
		if findErr != nil {
			log.Println("[---] Find error:", findErr)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		release, findErr := models.FindRelease(db, project, releaseId)
		if findErr != nil {
			log.Println("[---] Find error:", findErr)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		page, findErr := models.FindPageByName(db, release, vars["name"])
		if findErr != nil {
			log.Println("[---] Find error:", findErr)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		path := models.GeneratePagePath(project, release, vars["name"])
		imageBytes, err := sp.Get(path)
		if err != nil {
			log.Println("[---] error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not read the page file. Please try again later."))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", page.MimeType.String())
		w.Write(imageBytes)
	}
}

// DELETE /projects/{projectId}/releases/{releaseId}/pages/{pageId}

type deletePageRequest struct {
	ProjectID uint32
	ReleaseID uint32
	PageID    uint32
}

type deletePageResponse struct {
	Error   *string `json:"error"`
	Success bool    `json:"success"`
}

// deletePage removes a page from the DB and deletes the file containing the image.
func deletePage(db database.DB, sp storage_provider.Binary) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := deletePageRequest{}
		vars := mux.Vars(r)
		_, parseErr1 := fmt.Sscanf(vars["projectId"], "%d", &request.ProjectID)
		_, parseErr2 := fmt.Sscanf(vars["releaseId"], "%d", &request.ReleaseID)
		_, parseErr3 := fmt.Sscanf(vars["pageId"], "%d", &request.PageID)

		encoder := json.NewEncoder(w)
		if parseErr1 != nil || parseErr2 != nil || parseErr3 != nil {
			log.Println("[---] Parse error: %v || %v || %v\n", parseErr1, parseErr2, parseErr3)
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "projectId, releaseId, and pageId must all be integer IDs."
			encoder.Encode(deletePageResponse{&errMsg, false})
			return
		}

		project, findErr := models.FindProject(db, request.ProjectID)
		if findErr != nil {
			log.Println("[---] Find error:", findErr)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		release, findErr := models.FindRelease(db, project, request.ReleaseID)
		if findErr != nil {
			log.Println("[---] Find error:", findErr)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		page, findErr := models.FindPage(db, release, request.PageID)
		if findErr != nil {
			log.Println("[---] Find error:", findErr)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		filePath := models.GeneratePagePath(project, release, page.Name)

		log.Println("[+++] Attempting to delete page", page)
		deleteErr := page.Delete(db)
		if deleteErr != nil {
			log.Println("[---] Delete error:", findErr)
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := "Could not delete the requested page. Please try again later."
			encoder.Encode(deletePageResponse{&errMsg, false})
			return
		}
		sp.Unset(filePath)
		encoder.Encode(deletePageResponse{nil, true})
	}
}
