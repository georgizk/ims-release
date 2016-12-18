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
	"time"
)

func TestListProjects(t *testing.T) {
	mListProjects = func(db database.DB) ([]models.Project, error) {
		return []models.Project{}, nil
	}

	listFn := listProjects(nil)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/projects", nil)
	listFn.ServeHTTP(w, r)

	decoder := json.NewDecoder(w.Body)
	var resp ProjectResponse
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusOK, w.Code)

	projects := []models.Project{models.Project{
		Id:        5,
		CreatedAt: time.Now(),
		Name:      "name",
		Shorthand: "short",
		Status:    "unknown",
	}}
	mListProjects = func(db database.DB) ([]models.Project, error) {
		return projects, nil
	}

	w = httptest.NewRecorder()
	listFn.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, projects[0], resp.Result[0])

	expErr := errors.New("error")
	mListProjects = func(db database.DB) ([]models.Project, error) {
		return []models.Project{}, expErr
	}

	w = httptest.NewRecorder()
	listFn.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)
	assert.Equal(t, ErrRspListProjects.getError().Error(), resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateProject(t *testing.T) {
	// test bad request
	fn := createProject(nil)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/projects", strings.NewReader(""))
	fn.ServeHTTP(w, r)

	decoder := json.NewDecoder(w.Body)
	var resp ProjectResponse
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgJsonDecode, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// test save error
	const createReq = `{"name":"Georgi is coolish","shorthand":"geocool","description":"yeah","status":"completed"}`
	mSaveProject = func(db database.DB, p models.Project) (models.Project, error) {
		return p, errors.New("save error")
	}
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects", strings.NewReader(createReq))
	fn.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgCreateProject, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// test success case
	mSaveProject = func(db database.DB, p models.Project) (models.Project, error) {
		p.Id = 7
		return p, nil
	}
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/projects", strings.NewReader(createReq))
	fn.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint32(7), resp.Result[0].Id)
}

func TestGetProject(t *testing.T) {
	// test bad request
	fn := getProject(nil)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/projects/g", strings.NewReader(""))
	fn.ServeHTTP(w, r)

	decoder := json.NewDecoder(w.Body)
	var resp ProjectResponse
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgBadRequest, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// test not found
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{}, errors.New("not found")
	}
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/5", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusNotFound, w.Code)

	// test success
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(5), id)
		return models.Project{Id: id}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/projects/5", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, uint32(5), resp.Result[0].Id)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateProject(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	const updateReq = `{"name":"Georgi is coolish","shorthand":"geocool","description":"yeah","status":"completed"}`
	var resp ProjectResponse

	// test not found
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(7), id)
		return models.Project{}, errors.New("not found")
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/projects/7", nil)
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusNotFound, w.Code)

	// test decode error
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(7), id)
		return models.Project{Id: id}, nil
	}
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/projects/7", strings.NewReader(""))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgJsonDecode, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// test update error
	mUpdateProject = func(db database.DB, p models.Project) (models.Project, error) {
		assert.Equal(t, uint32(7), p.Id)
		return p, errors.New("update error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/projects/7", strings.NewReader(updateReq))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgProjectUpdate, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// test success case
	mUpdateProject = func(db database.DB, p models.Project) (models.Project, error) {
		assert.Equal(t, uint32(7), p.Id)
		assert.Equal(t, "Georgi is coolish", p.Name)
		assert.Equal(t, "geocool", p.Shorthand)
		assert.Equal(t, "yeah", p.Description)
		assert.Equal(t, "completed", p.Status)
		return p, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/projects/7", strings.NewReader(updateReq))
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, uint32(7), resp.Result[0].Id)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteProject(t *testing.T) {
	router := mux.NewRouter()
	registerHandlers(router, nil, nil)
	var resp ProjectResponse

	// test not found
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(7), id)
		return models.Project{}, errors.New("not found")
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("DELETE", "/projects/7", nil)
	router.ServeHTTP(w, r)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgNotFound, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusNotFound, w.Code)

	// test error fetching releases
	mFindProject = func(db database.DB, id uint32) (models.Project, error) {
		assert.Equal(t, uint32(7), id)
		return models.Project{Id: id}, nil
	}

	mListReleases = func(db database.DB, p models.Project) ([]models.Release, error) {
		return []models.Release{}, errors.New("list releases error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/7", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgUnexpected, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// test releases non-zero
	mListReleases = func(db database.DB, p models.Project) ([]models.Release, error) {
		return []models.Release{models.Release{}}, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/7", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgReleasesNotEmpty, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusExpectationFailed, w.Code)

	// test deletion error
	mListReleases = func(db database.DB, p models.Project) ([]models.Release, error) {
		return []models.Release{}, nil
	}

	mDeleteProject = func(db database.DB, p models.Project) (models.Project, error) {
		assert.Equal(t, uint32(7), p.Id)
		return p, errors.New("project delete error")
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/7", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, ErrMsgUnexpected, resp.getError().Error())
	assert.Equal(t, 0, len(resp.Result))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// success case
	mDeleteProject = func(db database.DB, p models.Project) (models.Project, error) {
		assert.Equal(t, uint32(7), p.Id)
		return p, nil
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/projects/7", nil)
	router.ServeHTTP(w, r)
	decoder = json.NewDecoder(w.Body)
	decoder.Decode(&resp)

	assert.Equal(t, nil, resp.getError())
	assert.Equal(t, 1, len(resp.Result))
	assert.Equal(t, uint32(7), resp.Result[0].Id)
	assert.Equal(t, http.StatusOK, w.Code)
}
