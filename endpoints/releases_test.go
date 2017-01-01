package endpoints

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"ims-release/assert"
	"ims-release/database"
	"ims-release/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListReleases(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/projects/5/releases", nil)
	var resp ReleaseResponse

	// test project not found
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{}, errors.New("some error")
	}
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test releases error
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{Id: id}, nil
	}

	mListReleases = func(db database.DB, p models.Project) ([]models.Release, error) {
		assert.Equal(t, uint32(5), p.Id)
		return []models.Release{}, errors.New("some error")
	}
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/5/releases", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgListReleases, resp.getError().Error())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test success case
	mListReleases = func(db database.DB, p models.Project) ([]models.Release, error) {
		assert.Equal(t, uint32(5), p.Id)
		return []models.Release{models.Release{Id: 6, ProjectID: p.Id}}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/5/releases", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint32(6), resp.Result[0].Id)
}

func TestCreateRelease(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/projects/5/releases", nil)
	var resp ReleaseResponse

	// test project not found
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{}, errors.New("some error")
	}
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test decoding error
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{Id: id}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/5/releases", strings.NewReader(""))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgJsonDecode, resp.getError().Error())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test save error
	mSaveRelease = func(db database.DB, release models.Release) (models.Release, error) {
		return models.Release{}, errors.New("some error")
	}

	const createReq = `{"identifier":"c1","version":1,"status":"draft"}`
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/5/releases", strings.NewReader(createReq))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgCreateRelease, resp.getError().Error())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test success case
	mSaveRelease = func(db database.DB, release models.Release) (models.Release, error) {
		release.Id = uint32(7)
		return release, nil
	}
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/5/releases", strings.NewReader(createReq))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, uint32(7), resp.Result[0].Id)
	assert.Equal(t, "c1", resp.Result[0].Identifier)
	assert.Equal(t, uint32(1), resp.Result[0].Version)
	assert.Equal(t, "draft", resp.Result[0].Status)
}

func TestBadGetReleaseRequest(t *testing.T) {
	// test bad request
	router := mux.NewRouter()
	// this is done so that the route will still match even w/ invalid request
	router.HandleFunc("/projects/{projectId}/releases/{releaseId}", getRelease(nil)).Methods("GET")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/projects/5/releases/g", nil)
	var resp ReleaseResponse

	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{Id: id}, nil
	}

	router.ServeHTTP(w, r)

	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgBadRequest, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetRelease(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/projects/5/releases/7", nil)
	var resp ReleaseResponse

	// test project not found
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{}, errors.New("some error")
	}

	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test release not found
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{Id: id}, nil
	}

	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		assert.Equal(t, uint32(5), p.Id)
		assert.Equal(t, uint32(7), id)
		return models.Release{}, errors.New("some error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/5/releases/7", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test success case
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		assert.Equal(t, uint32(5), p.Id)
		assert.Equal(t, uint32(7), id)
		return models.Release{Id: id, ProjectID: p.Id}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/5/releases/7", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, uint32(7), resp.Result[0].Id)
}

func TestUpdateRelease(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	var resp ReleaseResponse

	// test release not found
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{Id: id}, nil
	}

	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		assert.Equal(t, uint32(5), p.Id)
		assert.Equal(t, uint32(7), id)
		return models.Release{}, errors.New("some error")
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/projects/5/releases/7", nil)
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test invalid json
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		assert.Equal(t, uint32(5), p.Id)
		assert.Equal(t, uint32(7), id)
		return models.Release{Id: id, ProjectID: p.Id}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/projects/5/releases/7", strings.NewReader(""))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgJsonDecode, resp.getError().Error())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test editing published release error
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id, Version: uint32(1), Status: "released"}, nil
	}
	const UpdateReq = `{"identifier":"c1","version":2,"status":"released"}`
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/projects/5/releases/7", strings.NewReader(UpdateReq))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgMustDraft, resp.getError().Error())
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test downversion error
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id, Version: uint32(5)}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/projects/5/releases/7", strings.NewReader(UpdateReq))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgDownversioning, resp.getError().Error())
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test upversion required error
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id, Version: uint32(2), Status: "draft"}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/projects/5/releases/7", strings.NewReader(UpdateReq))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgMustUpversion, resp.getError().Error())
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test save error
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id, Version: uint32(1), Status: "draft"}, nil
	}

	mUpdateRelease = func(db database.DB, release models.Release) (models.Release, error) {
		assert.Equal(t, "released", release.Status)
		assert.Equal(t, "c1", release.Identifier)
		assert.Equal(t, uint32(2), release.Version)
		return release, errors.New("some error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/projects/5/releases/7", strings.NewReader(UpdateReq))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgReleaseUpdate, resp.getError().Error())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test success case
	mUpdateRelease = func(db database.DB, release models.Release) (models.Release, error) {
		return release, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/projects/5/releases/7", strings.NewReader(UpdateReq))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, "c1", resp.Result[0].Identifier)
}

func TestDeleteRelease(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	var resp ReleaseResponse

	// test release not found
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{Id: id}, nil
	}

	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		assert.Equal(t, uint32(5), p.Id)
		assert.Equal(t, uint32(7), id)
		return models.Release{}, errors.New("some error")
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("DELETE", "/projects/5/releases/7", nil)
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test list pages error
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id, Version: uint32(2), Status: "draft"}, nil
	}

	mListPages = func(db database.DB, release models.Release) ([]models.Page, error) {
		return []models.Page{}, errors.New("some error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/5/releases/7", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgUnexpected, resp.getError().Error())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test pages not empty
	mListPages = func(db database.DB, release models.Release) ([]models.Page, error) {
		return []models.Page{models.Page{}}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/5/releases/7", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgPagesNotEmpty, resp.getError().Error())
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test delete error
	mListPages = func(db database.DB, release models.Release) ([]models.Page, error) {
		assert.Equal(t, uint32(7), release.Id)
		return []models.Page{}, nil
	}

	mDeleteRelease = func(db database.DB, release models.Release) (models.Release, error) {
		return release, errors.New("some error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/5/releases/7", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgReleaseDelete, resp.getError().Error())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test success case
	mDeleteRelease = func(db database.DB, release models.Release) (models.Release, error) {
		assert.Equal(t, uint32(7), release.Id)
		return release, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/5/releases/7", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, uint32(7), resp.Result[0].Id)
}

func TestDownloadRelease(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	var resp ReleaseResponse

	// test no release found
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(12), id)
		return models.Project{Id: id}, nil
	}

	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		assert.Equal(t, uint32(12), p.Id)
		assert.Equal(t, uint32(70), id)
		return models.Release{}, errors.New("some error")
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/projects/12/releases/70/download/someName.zip", nil)
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// test release found but not in released state
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id, Status: "draft"}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/12/releases/70/download/someName.zip", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// test archive name mismatch
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id, Status: "released"}, nil
	}

	mGenerateArchiveName = func(project models.Project, release models.Release) string {
		assert.Equal(t, uint32(12), project.Id)
		assert.Equal(t, uint32(70), release.Id)
		return "someOtherName.zip"
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/12/releases/70/download/someName.zip", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// test list pages error
	mListPages = func(db database.DB, release models.Release) ([]models.Page, error) {
		return []models.Page{}, errors.New("some error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/12/releases/70/download/someOtherName.zip", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// test data get error
	mListPages = func(db database.DB, release models.Release) ([]models.Page, error) {
		return []models.Page{models.Page{Name: "p1.png"}}, nil
	}

	router = mux.NewRouter()
	var sp SpTest
	sp.Error = errors.New("some error")
	sp.Testing = t
	sp.ExpectedKey = "12/70/p1.png"
	sp.Bytes = []byte{1, 2}
	registerHandlers(router, nil, sp)

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/12/releases/70/download/someOtherName.zip", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// test success
	sp.Error = nil
	router = mux.NewRouter()
	registerHandlers(router, nil, sp)

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/12/releases/70/download/someOtherName.zip", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/zip", w.Header()["Content-Type"][0])
}
