package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"ims-release/endpoints"
	"ims-release/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func getArg(argIdx int) (string, error) {
	if len(os.Args) > argIdx {
		return os.Args[argIdx], nil
	} else {
		return "", errors.New("Argument missing")
	}
}

func main() {
	const usage = "Usage: import_project <apiRoute> <authToken> <projectFolder>\n"
	const missingApiRoute = "apiRoute missing.\n"
	const missingAuthToken = "authToken missing.\n"
	const missingProjectFolder = "projectFolder missing.\n"

	apiRoute, err := getArg(1)
	if err != nil {
		log.Print(usage)
		log.Fatal(missingApiRoute)
	}

	authToken, err := getArg(2)
	if err != nil {
		log.Print(usage)
		log.Fatal(missingAuthToken)
	}

	projectFolder, err := getArg(3)
	if err != nil {
		log.Print(usage)
		log.Fatal(missingProjectFolder)
	}

	importProject(apiRoute, authToken, projectFolder)
}

type pageAddRequest struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func importProject(apiRoute, authToken, projectsFolder string) error {
	// imports projects from the imangascans reader
	// assumes the following path: <projects folder>/<prjoect folder>/<chapter folder>/<images>
	projectDirs, err := ioutil.ReadDir(projectsFolder)
	if err != nil {
		return err
	}

	str, err := getProjects(apiRoute)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(strings.NewReader(str))
	var projectsResponse endpoints.ProjectResponse
	err = decoder.Decode(&projectsResponse)
	if err != nil {
		return err
	}

	for _, d := range projectDirs {
		if !d.IsDir() {
			continue
		}
		projectFolder := path.Join(projectsFolder, d.Name())
		releaseDirs, err := ioutil.ReadDir(projectFolder)
		if err != nil {
			return err
		}

		shorhtand := path.Base(projectFolder)
		projectId, err := addProject(apiRoute, authToken, shorhtand, projectsResponse)

		if err != nil {
			return err
		}

		str, err = getReleases(apiRoute, projectId)
		if err != nil {
			return err
		}

		decoder = json.NewDecoder(strings.NewReader(str))
		var releasesResponse endpoints.ReleaseResponse
		err = decoder.Decode(&releasesResponse)
		if err != nil {
			return err
		}

		for _, d := range releaseDirs {
			if !d.IsDir() {
				continue
			}
			tokens := strings.Split(d.Name(), "-")
			identifier := strings.Trim(tokens[0], " ")
			if len(identifier) == 0 {
				continue
			}
			if identifier[0] >= '0' && identifier[0] <= '9' {
				identifier = fmt.Sprintf("Ch%s", identifier)
			}
			identifier = strings.Replace(identifier, "Volume", "v", 1)
			identifier = strings.Replace(identifier, "(", "", 1)
			identifier = strings.Replace(identifier, ")", "", 1)
			identifier = strings.Replace(identifier, "Extra", "e", 1)
			identifier = strings.Replace(identifier, "Prologue part", "p", 1)
			if len(identifier) > 10 {
				log.Println("skipping release (identifier too long):", projectId, d.Name())
				continue
			}

			releaseId, err := addRelease(apiRoute, authToken, identifier, releasesResponse, projectId)
			if err != nil {
				return err
			}
			imageFolder := path.Join(projectFolder, d.Name())
			err = addPages(apiRoute, authToken, imageFolder, projectId, releaseId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func makeRequest(req *http.Request) (string, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return string(data), err
}

func addProject(apiRoute, authToken, shorthand string, prResp endpoints.ProjectResponse) (uint32, error) {
	if len(shorthand) > 30 {
		shorthand = shorthand[0:30]
	}

	for _, pr := range prResp.Result {
		if pr.Shorthand == shorthand {
			return pr.Id, nil
		}
	}

	project := models.Project{Shorthand: shorthand, Name: shorthand, Status: "active"}
	buffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buffer)
	err := encoder.Encode(project)
	if err != nil {
		return 0, err
	}
	httpReq, err := http.NewRequest("POST", apiRoute+"/projects", buffer)
	if err != nil {
		return 0, err
	}
	httpReq.Header.Add("Auth-Token", authToken)
	str, err := makeRequest(httpReq)
	if err != nil {
		return 0, err
	}
	decoder := json.NewDecoder(strings.NewReader(str))
	prResp = endpoints.ProjectResponse{}
	err = decoder.Decode(&prResp)
	if err != nil {
		return 0, err
	}

	if prResp.Error != nil {
		return 0, errors.New(*prResp.Error)
	}

	if 1 != len(prResp.Result) {
		return 0, errors.New("no result found")
	}

	return prResp.Result[0].Id, nil
}

func addRelease(apiRoute, authToken, identifier string, resp endpoints.ReleaseResponse, projectId uint32) (uint32, error) {
	uri := fmt.Sprintf("%s/projects/%d/releases", apiRoute, projectId)

	for _, rel := range resp.Result {
		if rel.Identifier == identifier && rel.Version == uint32(0) {
			return rel.Id, nil
		}
	}

	release := models.Release{Identifier: identifier, Version: uint32(0)}
	buffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buffer)
	err := encoder.Encode(release)
	if err != nil {
		return 0, err
	}
	httpReq, err := http.NewRequest("POST", uri, buffer)
	if err != nil {
		return 0, err
	}
	httpReq.Header.Add("Auth-Token", authToken)
	str, err := makeRequest(httpReq)
	if err != nil {
		return 0, err
	}
	decoder := json.NewDecoder(strings.NewReader(str))
	resp = endpoints.ReleaseResponse{}
	err = decoder.Decode(&resp)
	if err != nil {
		return 0, err
	}

	if resp.Error != nil {
		return 0, errors.New(*resp.Error)
	}

	if 1 != len(resp.Result) {
		return 0, errors.New("no result found")
	}

	return resp.Result[0].Id, nil
}

func getReleases(apiRoute string, projectId uint32) (string, error) {
	uri := fmt.Sprintf("%s/projects/%d/releases", apiRoute, projectId)
	httpReq, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return makeRequest(httpReq)
}

func getPages(apiRoute string, projectId, releaseId uint32) (string, error) {
	uri := fmt.Sprintf("%s/projects/%d/releases/%d/pages", apiRoute, projectId, releaseId)
	httpReq, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return makeRequest(httpReq)
}

func getProjects(apiRoute string) (string, error) {
	httpReq, err := http.NewRequest("GET", apiRoute+"/projects", nil)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return makeRequest(httpReq)
}

func addPages(apiRoute, authToken, imageFolder string, projectId, releaseId uint32) error {
	uri := fmt.Sprintf("%s/projects/%d/releases/%d/pages", apiRoute, projectId, releaseId)
	files, err := ioutil.ReadDir(imageFolder)
	if err != nil {
		return err
	}

	pagesStr, err := getPages(apiRoute, projectId, releaseId)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(strings.NewReader(pagesStr))
	var pagesResponse endpoints.PageResponse
	err = decoder.Decode(&pagesResponse)
	if err != nil {
		return err
	}

	if pagesResponse.Error != nil {
		return errors.New(*pagesResponse.Error)
	}

	addRequests := make([]pageAddRequest, 0, len(files))

	for _, file := range files {
		fn := file.Name()
		name := fn
		if file.IsDir() {
			continue
		}
		ext := path.Ext(fn)
		if strings.EqualFold(ext, ".png") {
			name = fn[0:len(fn)-len(ext)] + ".png"
		} else if strings.EqualFold(ext, ".jpg") || strings.EqualFold(ext, ".jpeg") {
			name = fn[0:len(fn)-len(ext)] + ".jpg"
		} else {
			continue
		}

		pageExists := false
		for _, pg := range pagesResponse.Result {
			if pg.Name == name {
				pageExists = true
				break
			}
		}

		if pageExists {
			continue
		}

		fn = filepath.Join(imageFolder, fn)
		f, err := ioutil.ReadFile(fn)
		if err != nil {
			return err
		}
		data := base64.StdEncoding.EncodeToString(f)
		addRequests = append(addRequests, pageAddRequest{Name: name, Data: data})
	}

	for _, req := range addRequests {
		buffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(buffer)
		err = encoder.Encode(req)
		if err != nil {
			return err
		}
		httpReq, err := http.NewRequest("POST", uri, buffer)
		if err != nil {
			return err
		}
		httpReq.Header.Add("Auth-Token", authToken)
		log.Printf("Adding page '%s' (projectId %d, releaseId %d)", req.Name, projectId, releaseId)
		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Println("response Status:", resp.Status)
			data, _ := ioutil.ReadAll(resp.Body)
			log.Println(string(data))
		}
	}

	return nil
}
