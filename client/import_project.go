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

func getOneshotDirs(projectsFolder string, projectDirs []os.FileInfo) (map[string]bool, error) {
	oneshotDirs := map[string]bool{} // key is folder path, true if it's a oneshot

	for _, d := range projectDirs {
		if !d.IsDir() {
			continue
		}
		projectFolder := path.Join(projectsFolder, d.Name())
		releaseDirs, err := ioutil.ReadDir(projectFolder)
		if err != nil {
			return oneshotDirs, err
		}
		isOneshotFolder := true
		isOneshotSuperfolder := (d.Name() == "Oneshot")
		for _, d := range releaseDirs {
			if !d.IsDir() {
				continue
			}

			if isOneshotSuperfolder {
				oneshotDir := path.Join(projectFolder, d.Name())
				oneshotDirs[oneshotDir] = true
			}
			isOneshotFolder = false
		}
		if isOneshotFolder {
			oneshotDirs[projectFolder] = true
		}
	}

	return oneshotDirs, nil
}

func importRelease(apiRoute, authToken, releaseFolder string, projectId uint32, releasesResponse endpoints.ReleaseResponse, oneshotDirs map[string]bool, usedIdentifiers map[string]string) (map[string]string, error) {
	d, err := os.Stat(releaseFolder)
	if err != nil {
		return usedIdentifiers, nil
	}
	if !d.IsDir() {
		return usedIdentifiers, nil
	}
	tokens := strings.Split(d.Name(), "-")
	identifier := strings.Trim(tokens[0], " ")
	if oneshotDirs[releaseFolder] {
		identifier = "" // use empty identifier for oneshots
	} else {
		if len(identifier) == 0 {
			return usedIdentifiers, nil
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
			log.Println("skipping release (identifier too long):", projectId, releaseFolder)
			return usedIdentifiers, nil
		}
	}

	identifierKey := fmt.Sprintf("%d-%s", projectId, identifier)
	existingPath := usedIdentifiers[identifierKey]
	if existingPath != "" {
		log.Printf("skipping release (identifier already used by %s): %d %s\n", existingPath, projectId, releaseFolder)
		return usedIdentifiers, nil
	}

	usedIdentifiers[identifierKey] = releaseFolder

	releaseId, err := addRelease(apiRoute, authToken, identifier, releasesResponse, projectId)
	if err != nil {
		return usedIdentifiers, err
	}
	err = addPages(apiRoute, authToken, releaseFolder, projectId, releaseId)
	if err != nil {
		return usedIdentifiers, err
	}
	return usedIdentifiers, nil
}

func importProject(apiRoute, authToken, projectsFolder string) error {
	// imports projects from the imangascans reader
	// assumes the following path: <projects folder>/<prjoect folder>/<chapter folder>/<images>
	projectDirs, err := ioutil.ReadDir(projectsFolder)
	if err != nil {
		return err
	}

	oneshotDirs, err := getOneshotDirs(projectsFolder, projectDirs)
	if err != nil {
		return err
	}

	str, err := getProjects(apiRoute)
	if err != nil {
		return err
	}

	projectsResponse, err := parseProjectResponse(str)
	if err != nil {
		return err
	}

	usedIdentifiers := map[string]string{} // a map of projectId-identifier to path

	// first add the oneshots
	for oneshotDir, _ := range oneshotDirs {
		usedIdentifiers, err = addOneshot(apiRoute, authToken, oneshotDir, projectsResponse, oneshotDirs, usedIdentifiers)
		if err != nil {
			return err
		}
	}

	for _, d := range projectDirs {
		if !d.IsDir() {
			continue
		}

		if d.Name() == "Oneshot" {
			continue
		}

		projectFolder := path.Join(projectsFolder, d.Name())
		if oneshotDirs[projectFolder] {
			continue
		}

		releaseDirs, err := ioutil.ReadDir(projectFolder)
		if err != nil {
			return err
		}

		name := path.Base(projectFolder)
		projectId, err := addProject(apiRoute, authToken, name, projectsResponse)

		if err != nil {
			return err
		}

		str, err = getReleases(apiRoute, projectId)
		if err != nil {
			return err
		}

		releasesResponse, err := parseReleaseResponse(str)
		if err != nil {
			return err
		}

		for _, d := range releaseDirs {
			releaseFolder := path.Join(projectFolder, d.Name())
			if oneshotDirs[releaseFolder] {
				continue
			}
			usedIdentifiers, err = importRelease(apiRoute, authToken,
				releaseFolder, projectId, releasesResponse,
				oneshotDirs, usedIdentifiers)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func addOneshot(apiRoute, authToken, oneshotDir string, projectsResponse endpoints.ProjectResponse, oneshotDirs map[string]bool, usedIdentifiers map[string]string) (map[string]string, error) {
	name := path.Base(oneshotDir)
	projectId, err := addProject(apiRoute, authToken, name, projectsResponse)

	str, err := getReleases(apiRoute, projectId)
	if err != nil {
		return usedIdentifiers, err
	}

	releasesResponse, err := parseReleaseResponse(str)
	if err != nil {
		return usedIdentifiers, err
	}

	return importRelease(apiRoute, authToken,
		oneshotDir, projectId, releasesResponse,
		oneshotDirs, usedIdentifiers)
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

func addProject(apiRoute, authToken, name string, prResp endpoints.ProjectResponse) (uint32, error) {
	shorthand := name
	if len(shorthand) > 30 {
		shorthand = shorthand[0:30]
	}

	for _, pr := range prResp.Result {
		if pr.Shorthand == shorthand {
			if pr.Name != name {
				pr.Name = name
				uri := fmt.Sprintf("%s/projects/%d", apiRoute, pr.Id)
				str, err := makeJsonRequest("PUT", uri, authToken, pr)
				if err != nil {
					log.Println("failed to update project", pr.Id, err)
				} else {
					resp, err := parseProjectResponse(str)
					if err != nil {
						log.Println("failed to update project", pr.Id, err)
					}
					if resp.Error != nil {
						log.Println("failed to update project", pr.Id, *resp.Error)
					}
				}
			}
			return pr.Id, nil
		}
	}

	project := models.Project{Shorthand: shorthand, Name: name, Status: "active"}
	uri := fmt.Sprintf("%s/projects", apiRoute)
	str, err := makeJsonRequest("POST", uri, authToken, project)
	if err != nil {
		return 0, err
	}
	resp, err := parseProjectResponse(str)
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

func makeJsonRequest(method, uri, authToken string, entity interface{}) (string, error) {
	buffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buffer)
	err := encoder.Encode(entity)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequest(method, uri, buffer)
	if err != nil {
		return "", err
	}
	httpReq.Header.Add("Auth-Token", authToken)
	return makeRequest(httpReq)
}

func addRelease(apiRoute, authToken, identifier string, resp endpoints.ReleaseResponse, projectId uint32) (uint32, error) {
	uri := fmt.Sprintf("%s/projects/%d/releases", apiRoute, projectId)

	for _, rel := range resp.Result {
		if rel.Identifier == identifier && rel.Version == uint32(0) {
			return rel.Id, nil
		}
	}

	release := models.Release{Identifier: identifier, Version: uint32(0)}
	str, err := makeJsonRequest("POST", uri, authToken, release)
	if err != nil {
		return 0, err
	}
	resp, err = parseReleaseResponse(str)
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

func parseReleaseResponse(str string) (endpoints.ReleaseResponse, error) {
	var resp endpoints.ReleaseResponse
	err := parseResponse(str, &resp)
	return resp, err
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

func parsePageResponse(str string) (endpoints.PageResponse, error) {
	var resp endpoints.PageResponse
	err := parseResponse(str, &resp)
	return resp, err
}

func getProjects(apiRoute string) (string, error) {
	uri := fmt.Sprintf("%s/projects", apiRoute)
	httpReq, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return makeRequest(httpReq)
}

func parseProjectResponse(str string) (endpoints.ProjectResponse, error) {
	var resp endpoints.ProjectResponse
	err := parseResponse(str, &resp)
	return resp, err
}

func parseResponse(str string, resp interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(str))
	err := decoder.Decode(resp)
	if err != nil {
		return err
	}
	return nil
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

	pagesResponse, err := parsePageResponse(pagesStr)
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
		log.Printf("Adding page '%s' (projectId %d, releaseId %d)", req.Name, projectId, releaseId)
		str, err := makeJsonRequest("POST", uri, authToken, req)
		if err != nil {
			return err
		}
		resp, err := parsePageResponse(str)
		if err != nil {
			return err
		}
		if resp.Error != nil {
			log.Println("error:", *resp.Error)
		}
	}

	return nil
}
