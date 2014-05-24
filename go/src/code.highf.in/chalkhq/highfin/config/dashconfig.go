package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type App struct {
	Lang    string   `json:"lang"`
	Version string   `json:"version"`
	Main    string   `json:"main"`
	Watch   []string `json:"watch"`
	Exclude []string `json:"exclude"`
	Npm     []string `json:"npm"`
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
		for i := 0; i < len(app.Watch); i++ {
			app.Watch[i], _ = filepath.Abs(parent_search + app.Watch[i])
		}
	}

	//dashConfig.App.Main, _ = filepath.Abs(parent_search + dashConfig.App.Main)

	return dashConfig
}
