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
		//todo: this is either triggered by highf.in admin panel, guppy directly, or from a post-receive hook
		// it should: checkout the latest of the requested branch of the project from coral
		// read the -.json file, run npm install or go install, run tests, delete the .git folder and create a tar
		// reading the -.json file should consult it's list of sharks for available resources
		// upload the tar file to that shark or spin up a new one

		// basic wiring: checkout dev-next branch, read -.json file, tar, upload to shark.
		project.Deploy(r)

	case "/salmon/update":
		// todo: this would be triggered by a git postreceive hook in the salmon repo
		// it needs to do a shallow clone of salmon to /octopus/salmon,
		// delete the .git folder, run git init, git add ., git commit -m "new salmon repo"
		// should also create devnext and devcurrent branches
		// then git create() will simply copy this new salmon repo's .git folder to a user's project git folder code.git
		// git symbolic-ref HEAD refs/heads/dev-next make the dev-next branch default, git branch -D master
		/*
			git branch dev-next
			git branch -D master
			git symbolic-ref HEAD refs/heads/dev-next
			// other branches are delete and created during deploys to various environments
		*/
	default:
		r.Response.Meta.Status = 404
		r.AddError("Path: " + r.Req.URL.Path + " is not valid")
		r.Kill(200)
	}
}
