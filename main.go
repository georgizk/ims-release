package main

import (
	"ims-release/config"
	"ims-release/endpoints"

	"errors"
	"log"
	"net/http"
	"os"
)

const MissingConf = "You must specify the path to the json configuration file.\n"
const InvalidConf = "Unable to parse config file. Make sure the file is valid.\n"

func getArg(argIdx int) (string, error) {
	if len(os.Args) > argIdx {
		return os.Args[argIdx], nil
	} else {
		return "", errors.New("Argument missing")
	}
}

func main() {
	const usage = "Usage: ims-release <configPath>\n"
	log.SetFlags(log.LstdFlags | log.Llongfile)
	cfgPath, err := getArg(1)
	if err != nil {
		log.Print(usage)
		log.Fatal(MissingConf)
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Print(usage)
		log.Fatal(InvalidConf)
	}

	handler := endpoints.NewHttpHandler(cfg)
	address := cfg.BindAddress
	log.Printf("Listening on %s\n", address)
	http.ListenAndServe(address, handler)
}
