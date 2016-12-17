package config

import (
	"encoding/json"
	"os"
)

// Config contains information that the API service generally needs to run.
// It includes the address (formatted like <ip>:<port>) to bind the HTTP server to,
// as well as the path to the directory to which image files are to be written.
type Config struct {
	BindAddress    string `json:"bindAddress"`
	ImageDirectory string `json:"imageDirectory"`
	DbProtocol     string `json:"dbProtocol"`
	DbAddress      string `json:"dbAddress"`
	DbName         string `json:"dbName"`
	DbUser         string `json:"dbUser"`
	DbPassword     string `json:"dbPassword"`
}

// MustLoad attempts to load a Config from a specified path and panics if it
// cannot successfully read or decode the contents of the file at that path.
func LoadConfig(path string) (*Config, error) {
	file, openErr := os.Open(path)
	if openErr != nil {
		return nil, openErr
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	decodeErr := decoder.Decode(&config)
	return &config, decodeErr
}
