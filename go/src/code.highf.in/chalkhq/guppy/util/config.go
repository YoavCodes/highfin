/*
	Get and Set config is geared toward command line interactions, specifically used in fry-box by Guppy
	all other functionality requires these parameters to be defined, when calling them directly via the api ie: by shark
	you'll have the config available as you're already dealing with a given project by the time you need Guppy
*/

package util

import "os"
import "io"
import "fmt"
import "os/exec"
import "encoding/json"
import "code.highf.in/chalkhq/shared/log"

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
		log.LogE(err)
	}

	if config.Account == "" || config.Project == "" || config.Email == "" {
		SetConfig(config)
		return GetConfig()
	}

	return config
}

func SetConfig(config Config) {
	_ = os.MkdirAll(GUPPY_PATH, 777)
	configFile, err := os.Create(GUPPY_CONFIG)

	log.LogE(err)
	if err != nil {
		return
	}

	defer configFile.Close()

	_ = configFile // this seems like half a thought

	if config.Account == "" {
		log.Log("Please enter your account name:")
		fmt.Scanf("%s", &config.Account)
	}

	if config.Project == "" {
		log.Log("Please enter your project name:")
		fmt.Scanf("%s", &config.Project)
	}

	if config.Email == "" {
		log.Log("Please enter your email address:")
		fmt.Scanf("%s", &config.Email)
	}

	// todo: add name to configuration
	//_ = exec.Command("git", "config", "--global", "user.name", config.Name).Run()
	_ = exec.Command("git", "config", "--global", "user.email", config.Email).Run()

	if config.Server == "" {
		config.Server = "10.10.10.50" // the default, public would be octopus.highf.in which will hit the squid and be routed to octopus
	}

	// json is encoded as a []byte
	configJson, err := json.MarshalIndent(config, "", "   ")
	log.LogE(err)

	configFile.Write(configJson)

	if err != nil {
		log.Log("couldn't create Guppy config\n" + err.Error())
	}

}
