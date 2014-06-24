package types

import (
	"code.highf.in/chalkhq/shared/config"
)

/*
	Sharks is a map of Shark servers and their available resources
	todo: find most efficient mechanism for finding available space for a given set of requirements.
	Mesh.Sharks[ip_address]struct{}

	Deploys is a map of project/env deploys, which shark they're on
	Mesh.Projects[deploy_id].Shark
*/
type Mesh struct {
	Sharks   map[string]*Shark
	Projects map[string]*Project
}

type Project struct {
	Info struct {
		GITrepo   string // /coral/chalhkq/highfin/code.git.. may support github git hosting
		Deploying string // a string ie: devnext, devcurrent, etc. of the currently deploying env.
	}
	DEVnext    config.DashConfig // map[appPart]
	QAnext     config.DashConfig
	DEVcurrent config.DashConfig
	QAcurrent  config.DashConfig
	PROD       config.DashConfig
	Temp       config.DashConfig // temporary. a project can deploy one env at a time. if successful it'll replace that env. otherwise Temp will just be cleared unapologetically
}

type Shark struct {
	Info struct {
		Ip          string // the shark's private ip for octopus to connect to
		Num_deploys int    // num deploys on the shark
		//Ports       []string // exposed ports
	}
	Cpu struct {
		Total     int
		Available int
	}
	Ram struct {
		Total     int
		Available int
	}
}
