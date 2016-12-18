package endpoints

import (
	"ims-release/config"
	"ims-release/database"
	"ims-release/storage_provider"

	"encoding/json"
	"errors"
	"log"
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
	registerHandlers(router, db, &sp)

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	corsRouter := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}))(loggedRouter)

	return corsRouter
}

func registerHandlers(r *mux.Router, db database.DB, sp storage_provider.Binary) {
	r.StrictSlash(true)
	RegisterProjectHandlers(r, db)
	RegisterReleaseHandlers(r, db, sp)
	RegisterPageHandlers(r, db, sp)
}

var (
	ErrMsgJsonDecode = "JSON format error or missing field detected."
	ErrRspJsonDecode = NewApiResponse(http.StatusBadRequest, &ErrMsgJsonDecode)
	ErrMsgBadRequest = "Bad request."
	ErrRspBadRequest = NewApiResponse(http.StatusBadRequest, &ErrMsgBadRequest)
	ErrMsgNotFound   = "Not found."
	ErrRspNotFound   = NewApiResponse(http.StatusNotFound, &ErrMsgNotFound)
	ErrMsgUnexpected = "Unexpected error."
	ErrRspUnexpected = NewApiResponse(http.StatusInternalServerError, &ErrMsgUnexpected)
)

var NoErr = NewApiResponse(http.StatusOK, nil)

type ApiResponseIf interface {
	getCode() int
	getError() error
}
type ApiResponse struct {
	Code  int     `json:"-"`
	Error *string `json:"error"`
}

func NewApiResponse(code int, e *string) ApiResponse {
	return ApiResponse{Code: code, Error: e}
}

func (r ApiResponse) getCode() int {
	return r.Code
}

func (r ApiResponse) getError() error {
	if nil != r.Error {
		return errors.New(*r.Error)
	} else {
		return nil
	}
}

func decodeHelper(r *http.Request, s interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(s)
	if err != nil {
		log.Println("[---] Decode error:", err)
		return ErrRspJsonDecode.getError()
	}
	return nil
}

func encodeHelper(w http.ResponseWriter, s ApiResponseIf) {
	encoder := json.NewEncoder(w)
	w.WriteHeader(s.getCode())
	w.Header().Set("Content-Type", "application/json")
	encoder.Encode(s)
}
