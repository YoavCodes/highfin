/*
	Get and Set config is geared toward command line interactions, specifically used in fry-box by Guppy
	all other functionality requires these parameters to be defined, when calling them directly via the api ie: by shark
	you'll have the config available as you're already dealing with a given project by the time you need Guppy
*/

package main

import "os"
import "io"
import "fmt"
import "encoding/json"

const (
	SHARK_PATH   string = "/shark"
	SHARK_CONFIG string = SHARK_PATH + "/config.json"
)

type Config struct {
	Port string `json:"port"`
}

func GetConfig() Config {
	var config Config

	configFile, err := os.Open(SHARK_CONFIG)

	if err != nil {
		// there is no config file
		SetConfig(config)
		return GetConfig()
	}

	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err == io.EOF || err == nil {
	} else {
		fmt.Println(err)
	}

	if config.Port == "" {
		SetConfig(config)
		return GetConfig()
	}

	return config
}

func SetConfig(config Config) {
	_ = os.MkdirAll(SHARK_PATH, 777)
	configFile, err := os.Create(SHARK_CONFIG)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}

	defer configFile.Close()

	_ = configFile // this seems like half a thought

	if config.Port == "" {
		config.Port = "80"
	}

	// json is encoded as a []byte
	configJson, err := json.MarshalIndent(config, "", "   ")
	fmt.Println(err)

	configFile.Write(configJson)

	if err != nil {
		fmt.Println("couldn't create Guppy config\n" + err.Error())
	}

}
