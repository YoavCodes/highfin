package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Endpoint struct {
	Path string `json:"path"`
	Port string `json:"port"`
}

type Static struct {
	Path string `json:"path"`
	Dir  string `json:"dir"`
}

type Exec struct {
	Lang      string     `json:"lang"`
	Version   string     `json:"version"`
	Main      string     `json:"main"`
	Watch     []string   `json:"watch"`
	Exclude   []string   `json:"exclude"`
	Npm       []string   `json:"npm"`
	Endpoints []Endpoint `json:"endpoints"`
}

type App struct {
	Execs      []Exec            `json:"exec"`
	Statics    []Static          `json:"static"`
	Sharkports map[string]string `json:"sharkports"` // map[jellyport]sharkport
	Domains    map[string]string `json:"domains"`    // map[domain]sharkport
	Instances  []string          `json:"instances"`  //jellyports
	Balances   map[string]string `json:"balances"`   // map[port]instance_index_range
	Deploys    map[string]string `json:"deploys"`    //map[instanceID]sharkport
}

type DashConfig struct {
	Apps map[string]App

	BasePath string // absolute path to app's root directory. ie: where the -.json file is
}

func GetDashConfig(path string) DashConfig {
	var dashConfig DashConfig
	parent_search := ""
	dashConfigFile, err := os.Open(path + "-.json")
	for err != nil {
		parent_search += "../"
		var abs string
		abs, err = filepath.Abs(path + parent_search + "-.json")
		if abs == "/-.json" {
			fmt.Println("Could not find -.json file.")
			return dashConfig
		}
		dashConfigFile, err = os.Open(parent_search + "-.json")
	}

	defer dashConfigFile.Close()

	jsonParser := json.NewDecoder(dashConfigFile)
	if err = jsonParser.Decode(&dashConfig.Apps); err == io.EOF || err == nil {
	} else {
		fmt.Println("Could not parse -.json files")
		fmt.Println(err.Error())
	}

	dashConfig.BasePath, _ = filepath.Abs(parent_search)

	for j := range dashConfig.Apps {
		app := dashConfig.Apps[j]
		for k := 0; k < len(app.Execs); k++ {
			appPart := app.Execs[k]
			for i := 0; i < len(appPart.Watch); i++ {
				appPart.Watch[i], _ = filepath.Abs(parent_search + appPart.Watch[i])
			}
		}
	}

	//dashConfig.App.Main, _ = filepath.Abs(parent_search + dashConfig.App.Main)

	return dashConfig
}
