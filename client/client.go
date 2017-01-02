package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	const usage = "Usage: client <apiRoute> <authToken> <imageFolder>\n"
	const missingApiRoute = "apiRoute missing.\n"
	const missingAuthToken = "authToken missing.\n"
	const missingImageFolder = "imageFolder missing.\n"

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

	imageFolder, err := getArg(3)
	if err != nil {
		log.Print(usage)
		log.Fatal(missingImageFolder)
	}

	err = addImages(apiRoute, authToken, imageFolder)
	if err != nil {
		log.Print(usage)
		log.Fatal(err)
	}
}

type pageAddRequest struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func addImages(apiRoute, authToken, imageFolder string) error {
	files, err := ioutil.ReadDir(imageFolder)
	if err != nil {
		return err
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
		httpReq, err := http.NewRequest("POST", apiRoute, buffer)
		if err != nil {
			return err
		}
		httpReq.Header.Add("Auth-Token", authToken)

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		log.Println("response Status:", resp.Status)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("response Body:", string(body))
	}

	return nil
}
