package config

import "os"
import "io"
import "encoding/json"
import "path/filepath"
import (
	"fmt"
)

type DashConfig struct {
	App struct {
		Lang    string   `json:"lang"`
		Version string   `json:"version"`
		Main    string   `json:"main"`
		Watch   []string `json:"watch"`
	} `json:"app"`
	BasePath string // absolute path to app's root directory. ie: where the -.json file is
}

func GetDashConfig(...string) DashConfig {
	var dashConfig DashConfig
	parent_search := ""
	dashConfigFile, err := os.Open("-.json")
	for err != nil {
		parent_search += "../"
		var abs string
		abs, err = filepath.Abs(parent_search + "-.json")
		if abs == "/-.json" {
			fmt.Println("Could not find -.json file.")
			return dashConfig
		}
		dashConfigFile, err = os.Open(parent_search + "-.json")
	}

	defer dashConfigFile.Close()

	jsonParser := json.NewDecoder(dashConfigFile)
	if err = jsonParser.Decode(&dashConfig); err == io.EOF || err == nil {
	} else {
		fmt.Println(err)
	}

	dashConfig.BasePath, _ = filepath.Abs(parent_search)

	for i := 0; i < len(dashConfig.App.Watch); i++ {
		dashConfig.App.Watch[i], _ = filepath.Abs(parent_search + dashConfig.App.Watch[i])
	}

	//dashConfig.App.Main, _ = filepath.Abs(parent_search + dashConfig.App.Main)

	return dashConfig
}
