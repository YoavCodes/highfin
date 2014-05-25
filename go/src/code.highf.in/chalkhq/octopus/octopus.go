package main

/*
Octopus is a Shark & Sqid co-ordinator.
- spins up new instances
- aggregates and respondds to health checks
- maintains open connection to Squids and Sharks at all times
- allocates Sharks to deploy containers
- manages Coral and builds containers before sending tar to a given Shark to deploy
*/

import (
	"bytes"
	"code.highf.in/chalkhq/highfin/types"
	"code.highf.in/chalkhq/octopus/api/project"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Paths struct {
	Base      string
	Config    string
	Templates string
	Public    string
	Routes    string
}

var (
	paths Paths
)

// todo: why are we compiling less on the server and not the client? for javascript-less clients? but then our whole app is dead

func init() {
	var whoami bytes.Buffer
	cmd := exec.Command("whoami")
	cmd.Stdout = &whoami
	_ = cmd.Run()
	if strings.Index(whoami.String(), "root") == -1 {
		fmt.Println("octopus must be run as root")
		os.Exit(1)
		return
	}
}

func main() {
	fmt.Println("running octopus")
	config := GetConfig()

	UpdateSalmon()

	fmt.Println("listening on " + config.Port)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("/")
		r := types.Response{}

		r.Response.Meta.Status = 200

		r.Response.Meta.Errors = make([]string, 0)

		r.Response.Data = make(map[string]interface{})

		r.W = w

		r.Req = *req
		r.Req.ParseForm()
		//r.Req.ParseMultipartForm(64)

		r.Fields = r.Req.Form
		if len(r.Hashbang) > 0 {
			r.Segments = strings.Split(r.Hashbang, "/")
		}

		router(r)

		//fmt.Println("/", r.Req.PostForm, r.Req.MultipartForm, r.Req.PostFormValue("project"))

	})

	//http.Handle("/", http.FileServer(http.Dir(paths.Config)))

	// Simple static webserver:
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
	//http.FileServer(http.Dir("./public")

}

func router(r types.Response) {
	fmt.Println("router")
	switch r.Req.URL.Path {
	case "/project/create":
		fmt.Println("create called")
		project.Create(r)

	case "/project/deploy":
		// todo:reading the -.json file should consult it's list of sharks for available resources
		// basic wiring: checkout dev-next branch, read -.json file, tar, upload to shark.
		project.Deploy(r)

	case "/salmon/update":
		UpdateSalmon()

	default:
		r.Response.Meta.Status = 404
		r.AddError("Path: " + r.Req.URL.Path + " is not valid")
		r.Kill(200)
	}
}
