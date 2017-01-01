package endpoints

import (
	"encoding/base64"
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

func TestListPages(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	var resp PageResponse

	// test fetch release error
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
	r, _ := http.NewRequest("GET", "/projects/12/releases/70/pages", nil)
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test list pages error
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id}, nil
	}

	mListPages = func(db database.DB, release models.Release) ([]models.Page, error) {
		return []models.Page{}, errors.New("some error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/12/releases/70/pages", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgListPages, resp.getError().Error())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test success case
	mListPages = func(db database.DB, release models.Release) ([]models.Page, error) {
		return []models.Page{models.Page{Id: uint32(71)}}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/12/releases/70/pages", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, uint32(71), resp.Result[0].Id)
}

type SpTest struct {
	Bytes             []byte
	Error             error
	IsExists          bool
	Testing           *testing.T
	ExpectedKey       string
	ExpectedBytesBenc string
}

func (sp SpTest) Set(key string, data []byte) error {
	assert.Equal(sp.Testing, sp.ExpectedKey, key)
	assert.Equal(sp.Testing, sp.ExpectedBytesBenc, base64.StdEncoding.EncodeToString(data))
	return sp.Error
}

func (sp SpTest) Get(key string) ([]byte, error) {
	assert.Equal(sp.Testing, sp.ExpectedKey, key)
	return sp.Bytes, sp.Error
}

func (sp SpTest) Unset(key string) error {
	assert.Equal(sp.Testing, sp.ExpectedKey, key)
	return sp.Error
}

func (sp SpTest) Exists(key string) bool {
	assert.Equal(sp.Testing, sp.ExpectedKey, key)
	return sp.IsExists
}

func TestCreate(t *testing.T) {
	router := mux.NewRouter()
	var sp SpTest
	registerHandlers(router, nil, sp)
	var resp PageResponse

	// test fetch release error
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
	r, _ := http.NewRequest("POST", "/projects/12/releases/70/pages", nil)
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test decode error
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/12/releases/70/pages", strings.NewReader(""))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgJsonDecode, resp.getError().Error())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test bad image data
	const badImageData = `{"name":"pageName.jpg", "data":"asd"}`

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/12/releases/70/pages", strings.NewReader(badImageData))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgBadImageData, resp.getError().Error())
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test bad extension
	const badExtension = `{"name":"noExtension", "data":""}`

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/12/releases/70/pages", strings.NewReader(badExtension))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgWrongType, resp.getError().Error())
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test extension-data mismatch
	const bencJpg = "/9j/4AAQSkZJRgABAQEAYABgAAD/4QBaRXhpZgAATU0AKgAAAAgABQMBAAUAAAABAAAASgMDAAEAAAABAAAAAFEQAAEAAAABAQAAAFERAAQAAAABAAAOxFESAAQAAAABAAAOxAAAAAAAAYagAACxj//bAEMAAgEBAgEBAgICAgICAgIDBQMDAwMDBgQEAwUHBgcHBwYHBwgJCwkICAoIBwcKDQoKCwwMDAwHCQ4PDQwOCwwMDP/bAEMBAgICAwMDBgMDBgwIBwgMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDP/AABEIAAEAAQMBIgACEQEDEQH/xAAfAAABBQEBAQEBAQAAAAAAAAAAAQIDBAUGBwgJCgv/xAC1EAACAQMDAgQDBQUEBAAAAX0BAgMABBEFEiExQQYTUWEHInEUMoGRoQgjQrHBFVLR8CQzYnKCCQoWFxgZGiUmJygpKjQ1Njc4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4eLj5OXm5+jp6vHy8/T19vf4+fr/xAAfAQADAQEBAQEBAQEBAAAAAAAAAQIDBAUGBwgJCgv/xAC1EQACAQIEBAMEBwUEBAABAncAAQIDEQQFITEGEkFRB2FxEyIygQgUQpGhscEJIzNS8BVictEKFiQ04SXxFxgZGiYnKCkqNTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqCg4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2dri4+Tl5ufo6ery8/T19vf4+fr/2gAMAwEAAhEDEQA/AP34ooorMD//2Q=="
	const bencPng = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAANSURBVBhXY/j3/+9/AAnzA/pJMr8HAAAAAElFTkSuQmCC"

	const extMismatch1 = `{"name":"fileName.png", "data":"` + bencJpg + `"}`
	const extMismatch2 = `{"name":"fileName.jpg", "data":"` + bencPng + `"}`
	const extMismatch3 = `{"name":"fileName.jpg", "data":""}`

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/12/releases/70/pages", strings.NewReader(extMismatch1))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgWrongType, resp.getError().Error())
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/12/releases/70/pages", strings.NewReader(extMismatch2))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgWrongType, resp.getError().Error())
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/12/releases/70/pages", strings.NewReader(extMismatch3))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgWrongType, resp.getError().Error())
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test save to disk error
	const dataJpg = `{"name":"fileName.jpg", "data":"` + bencJpg + `"}`
	const dataPng = `{"name":"fileName.png", "data":"` + bencPng + `"}`
	sp.Error = errors.New("some error")
	sp.Testing = t
	sp.ExpectedKey = "12/70/fileName.jpg"
	sp.ExpectedBytesBenc = bencJpg
	router = mux.NewRouter()
	registerHandlers(router, nil, sp)

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/12/releases/70/pages", strings.NewReader(dataJpg))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgCreatePage, resp.getError().Error())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test save to db error
	mSavePage = func(db database.DB, page models.Page) (models.Page, error) {
		return models.Page{}, errors.New("some error")
	}
	sp.Error = nil
	sp.Testing = t
	sp.ExpectedKey = "12/70/fileName.png"
	sp.ExpectedBytesBenc = bencPng
	router = mux.NewRouter()
	registerHandlers(router, nil, sp)

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/12/releases/70/pages", strings.NewReader(dataPng))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgCreatePage, resp.getError().Error())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test success
	mSavePage = func(db database.DB, page models.Page) (models.Page, error) {
		assert.Equal(t, uint32(70), page.ReleaseID)
		return models.Page{Id: uint32(100)}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects/12/releases/70/pages", strings.NewReader(dataPng))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, uint32(100), resp.Result[0].Id)
}

func TestGetPage(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	var resp PageResponse

	// test release not found
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
	r, _ := http.NewRequest("GET", "/projects/12/releases/70/pages/thePage.png", nil)
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// test page not found
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id}, nil
	}

	mFindPageByName = func(db database.DB, release models.Release, name string) (models.Page, error) {
		return models.Page{}, errors.New("some error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/12/releases/70/pages/thePage.png", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// test page data not found
	var sp SpTest
	sp.Error = errors.New("some error")
	sp.Testing = t
	sp.ExpectedKey = "12/70/thePage.png"
	const bencPng = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAANSURBVBhXY/j3/+9/AAnzA/pJMr8HAAAAAElFTkSuQmCC"
	sp.Bytes, _ = base64.StdEncoding.DecodeString(bencPng)
	router = mux.NewRouter()
	registerHandlers(router, nil, sp)

	mFindPageByName = func(db database.DB, release models.Release, name string) (models.Page, error) {
		assert.Equal(t, uint32(70), release.Id)
		assert.Equal(t, "thePage.png", name)
		return models.Page{Name: name, Id: uint32(100), ReleaseID: release.Id, MimeType: models.MimeTypeFromFilename(name)}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/12/releases/70/pages/thePage.png", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// test success
	sp.Error = nil
	router = mux.NewRouter()
	registerHandlers(router, nil, sp)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/12/releases/70/pages/thePage.png", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header()["Content-Type"][0])
}

func TestBadDeletePageRequest(t *testing.T) {
	// test bad request
	router := mux.NewRouter()
	// this is done so that the route will still match even w/ invalid request
	router.HandleFunc("/projects/{projectId}/releases/{releaseId}/pages/{pageId}", deletePage(nil, nil)).Methods("DELETE")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("DELETE", "/projects/12/releases/70/pages/g", nil)
	var resp ReleaseResponse

	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(12), id)
		return models.Project{Id: id}, nil
	}

	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id}, nil
	}

	router.ServeHTTP(w, r)

	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgBadRequest, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePage(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	var resp PageResponse

	// test fetch release error
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
	r, _ := http.NewRequest("DELETE", "/projects/12/releases/70/pages/100", nil)
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test find page error
	mFindRelease = func(db database.DB, p models.Project, id uint32) (models.Release, error) {
		return models.Release{Id: id, ProjectID: p.Id}, nil
	}

	mFindPage = func(db database.DB, release models.Release, pageId uint32) (models.Page, error) {
		return models.Page{}, errors.New("some error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/12/releases/70/pages/100", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test delete page error
	mFindPage = func(db database.DB, release models.Release, pageId uint32) (models.Page, error) {
		assert.Equal(t, uint32(100), pageId)
		assert.Equal(t, uint32(70), release.Id)
		return models.Page{Name: "somePage.png", Id: pageId, ReleaseID: release.Id}, nil
	}

	mDeletePage = func(db database.DB, page models.Page) (models.Page, error) {
		return page, errors.New("some error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/12/releases/70/pages/100", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgDeletePage, resp.getError().Error())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, len(resp.Result))

	// test success
	var sp SpTest
	sp.Testing = t
	sp.ExpectedKey = "12/70/somePage.png"
	router = mux.NewRouter()
	registerHandlers(router, nil, sp)

	mDeletePage = func(db database.DB, page models.Page) (models.Page, error) {
		assert.Equal(t, uint32(100), page.Id)
		assert.Equal(t, uint32(70), page.ReleaseID)
		return page, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/12/releases/70/pages/100", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, uint32(100), resp.Result[0].Id)
}
