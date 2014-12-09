package config

import (
	"code.highf.in/chalkhq/shared/log"
	"encoding/json"
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

type Less struct {
	From string `json:"from"`
	To   string `json:"to"`
	Min  bool   `json:"min"`
}

type Jasmine struct {
	Frontend string `json:"frontend"`
	Backend  string `json:"backend"`
}

type Exec struct {
	Lang           string     `json:"lang"`
	Version        string     `json:"version"`
	Main           string     `json:"main"`
	Watch          []string   `json:"watch"`
	Exclude        []string   `json:"exclude"`
	Cachecontrols  []string   `json:"cachecontrol"`
	Less           []Less     `json:"less"`
	Jasmine        Jasmine    `json:"jasmine"`
	GruntDirectory string     `json:"gruntdirectory"`
	Npm            []string   `json:"npm"`
	Endpoints      []Endpoint `json:"endpoints"`
}

type App struct {
	Type       string              `json:"type"`
	Execs      []Exec              `json:"exec"`
	Statics    []Static            `json:"static"`
	Sharkports map[string]string   `json:"sharkports"` // map[jellyport]sharkport
	Domains    map[string][]string `json:"domains"`    // map[domain][]sharkports
	Instances  []string            `json:"instances"`  //jellyports
	Balances   map[string]string   `json:"balances"`   // map[port]instance_index_range
	Deploys    map[string]string   `json:"deploys"`    //map[instanceID]sharkport
}

type DashConfig struct {
	Apps map[string]App

	BasePath string // absolute path to app's root directory. ie: where the -.json file is
}

func GetDashConfig(path string) DashConfig {
	var dashConfig DashConfig
	parent_search := "./"
	dashConfigFile, err := os.Open(path + "-.json")
	for err != nil {
		parent_search += "../"
		var abs string
		abs, err = filepath.Abs(path + parent_search + "-.json")
		if abs == "/-.json" {
			log.Log("Could not find -.json file.")
			return dashConfig
		}
		dashConfigFile, err = os.Open(parent_search + "-.json")
	}

	defer dashConfigFile.Close()

	jsonParser := json.NewDecoder(dashConfigFile)
	if err = jsonParser.Decode(&dashConfig.Apps); err == io.EOF || err == nil {
	} else {
		log.Log("Could not parse -.json files")
		log.Log(err.Error())
	}

	dashConfig.BasePath, _ = filepath.Abs(parent_search)
	// convert paths to absolute
	for j := range dashConfig.Apps {
		app := dashConfig.Apps[j]
		for k := 0; k < len(app.Execs); k++ {
			appPart := app.Execs[k]
			// get abs paths
			appPart.Main, _ = filepath.Abs(parent_search + appPart.Main)

			// if appPart.Jasmine.Backend != "" {
			// 	appPart.Jasmine.Backend, _ = filepath.Abs(parent_search + appPart.Jasmine.Backend)
			// }
			// if appPart.Jasmine.Frontend != "" {
			// 	log.Log("good")
			// 	appPart.Jasmine.Frontend = dashConfig.BasePath + appPart.Jasmine.Frontend
			// }
			// watch folders
			for i := 0; i < len(appPart.Watch); i++ {
				appPart.Watch[i], _ = filepath.Abs(parent_search + appPart.Watch[i])
			}
			// exclude folders
			for i := 0; i < len(appPart.Exclude); i++ {
				appPart.Exclude[i], _ = filepath.Abs(parent_search + appPart.Exclude[i])
			}
			// less/css folders
			for i := 0; i < len(appPart.Less); i++ {
				appPart.Less[i].From, _ = filepath.Abs(parent_search + appPart.Less[i].From)
				appPart.Less[i].To, _ = filepath.Abs(parent_search + appPart.Less[i].To)
			}
		}
	}

	//dashConfig.App.Main, _ = filepath.Abs(parent_search + dashConfig.App.Main)

	return dashConfig
}
