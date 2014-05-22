/*
	Get and Set config is geared toward command line interactions, specifically used in fry-box by Guppy
	all other functionality requires these parameters to be defined, when calling them directly via the api ie: by shark
	you'll have the config available as you're already dealing with a given project by the time you need Guppy
*/

package util

import "os"
import "io"
import "fmt"
import "encoding/json"

const (
	GUPPY_PATH   string = "/guppy"
	GUPPY_CONFIG string = GUPPY_PATH + "/config.json"
)

type Config struct {
	Account string `json:"account"`
	Project string `json:"project"`
	Email   string `json:"email"`
	Server  string `json:"server"`
}

func GetConfig() Config {
	var config Config

	configFile, err := os.Open(GUPPY_CONFIG)

	if err != nil {
		// there is no config file
		SetConfig(config)
		return GetConfig()
	}

	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err == io.EOF || err == nil {
	} else {
		LogE(err)
	}

	if config.Account == "" || config.Project == "" || config.Email == "" {
		SetConfig(config)
		return GetConfig()
	}

	return config
}

func SetConfig(config Config) {
	configFile, err := os.Create(GUPPY_CONFIG)

	LogE(err)
	if err != nil {
		return
	}

	defer configFile.Close()

	_ = configFile // this seems like half a thought

	if config.Account == "" {
		Log("Please enter your account name:")
		fmt.Scanf("%s", &config.Account)
	}

	if config.Project == "" {
		Log("Please enter your project name:")
		fmt.Scanf("%s", &config.Project)
	}

	if config.Email == "" {
		Log("Please enter your email address:")
		fmt.Scanf("%s", &config.Email)
	}

	if config.Server == "" {
		config.Server = "http://highf.in" // the default
	}

	// json is encoded as a []byte
	configJson, err := json.MarshalIndent(config, "", "   ")
	LogE(err)

	configFile.Write(configJson)

	if err != nil {
		Log("couldn't create Guppy config\n" + err.Error())
	}

}
