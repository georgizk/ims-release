package endpoints

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"image/jpeg"
	"image/png"
	"ims-release/database"
	"ims-release/models"
	"ims-release/storage_provider"
	"log"
	"net/http"
)

func RegisterThumbnailHandlers(r *mux.Router, db database.DB, sp storage_provider.Binary) {
	root := "/projects/{projectId:[0-9]+}/releases/{releaseId:[0-9]+}/thumbnails"
	sr := r.PathPrefix(root).Subrouter()
	sr.HandleFunc("/{name}", getThumbnail(db, sp)).Methods("GET")
}

func getThumbnail(db database.DB, sp storage_provider.Binary) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		project, release, err := fetchReleaseUsingRequestArgs(db, w, r, false)
		if err != nil {
			log.Println("[---] Release fetch error:", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		vars := mux.Vars(r)
		page, err := mFindPageByName(db, release, vars["name"])
		if err != nil {
			log.Println("[---] Find error:", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		path := mGeneratePagePath(project, release, page.Name)
		imageBytes, err := sp.Get(path)
		if err != nil {
			log.Println("[---] error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		mimeType := page.MimeType
		buffer := bytes.NewBuffer([]byte{})
		const maxHeight = 300
		const maxWidth = 200
		switch mimeType {
		case models.MimeTypePng:
			image, err := png.Decode(bytes.NewReader(imageBytes))
			if err != nil {
				log.Println("[---] error:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			image = resize.Thumbnail(maxHeight, maxWidth, image, resize.Bilinear)
			err = png.Encode(buffer, image)
		case models.MimeTypeJpg:
			image, err := jpeg.Decode(bytes.NewReader(imageBytes))
			if err != nil {
				log.Println("[---] error:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			image = resize.Thumbnail(maxHeight, maxWidth, image, resize.Bilinear)
			err = jpeg.Encode(buffer, image, nil)
		case models.MimeTypeUnknown:
			fallthrough
		default:
			log.Println("[---] error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", page.MimeType.String())
		w.Write(buffer.Bytes())
	}
}
