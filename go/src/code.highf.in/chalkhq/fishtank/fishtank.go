package main

import (
	dConfig "code.highf.in/chalkhq/shared/config"
	Log "code.highf.in/chalkhq/shared/log"
	"code.highf.in/chalkhq/shared/persistence"
	"code.highf.in/chalkhq/shared/types"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	FISHTANK_PATH   string = "/etc/fishtank"
	FISHTANK_CONFIG string = FISHTANK_PATH + "/config.json"
	DEFAULT_PORT    string = "6000"
	DATA_JSON       string = "/etc/fishtank/data.json"
	INSTALL_PATH    string = "/usr/bin"
	VERSION         string = "0.1.0"
)

type Project struct {
	Current_revision string `json:"current_revision"`
	Revisions        map[string]Revision
	Branch           string `json:"branch"`
	sync.Mutex
}

type Revision struct {
	cmds []*exec.Cmd
}

type Data struct {
	Projects map[string]Project `json:"projects"`
}

//var projects map[string]Project
var data Data

var config Config

// Fishtank is a single-server CD manager
func main() {

	go gracefulShutdown()
	checkInstalledVersion()

	persistence.GetData(&data, DATA_JSON)
	//go persistence.PersistData(&data, DATA_JSON, &projectsChanged)

	var command string
	// get the command from the commandline
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {

	case "install":
		Install()
	case "uninstall":
		Uninstall()

	case "add":
		Add()
	case "remove":
		Remove()
	case "post-receive":
		if len(os.Args) != 7 {
			Log.Log("not enough args: guppy post-receive project oldrev newrev branch refname")
			return
		}
		PostReceive(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	case "-h", "--help":
		Log.Log("Try: fishtank [install, add, post-receive]")
	case "-v":
		fmt.Println(VERSION)
	default:
		// by default run itself as a server, unless it's already running
		_, err := http.Get("http://127.0.0.1:" + config.Port)
		if err != nil {
			StartServer()
		} else {
			Log.Log("not starting server. port " + config.Port + " already in use. fishtank may already be running")
		}

	}
}

func gracefulShutdown() {
	sig_chan := make(chan os.Signal, 1)
	// signal will be caught when killing from top or manually from guppy on filechange
	signal.Notify(sig_chan, syscall.SIGTERM) // listen for TERM signal
	go func() {
		fmt.Println("waiting for kill signal...")

		// wait for signal
		//signal.Notify(sig_chan, syscall.SIGKILL)
		sig := <-sig_chan
		fmt.Println("Got signal:", sig)
		// kill all project execs
		for i := range data.Projects {
			for j := range data.Projects[i].Revisions {
				revision := data.Projects[i].Revisions[j]
				for k := range revision.cmds {
					cmd := revision.cmds[k]

					cmd.Process.Signal(syscall.SIGTERM)

					go func(cmd *exec.Cmd) {
						time.Sleep(300 * time.Millisecond)
						cmd.Process.Kill()
					}(cmd)
				}
			}
		}
		time.Sleep(300 * time.Millisecond)
		fmt.Println("finished cleanin up")

		os.Exit(0)
	}()
}

func HealthCheck() {
	// projects[project].Lock()
	// check if processes are still alive every second
	tick := time.Tick(1 * time.Second)
	for _ = range tick {
		for j := range data.Projects {
			var project = data.Projects[j]
			project.Lock()

			current_rev := data.Projects[j].Revisions[data.Projects[j].Current_revision]
			for i := 0; i < len(current_rev.cmds); i++ {

				cmd := current_rev.cmds[i]

				if cmd.ProcessState != nil && cmd.ProcessState.Exited() == true {
					// re-run
					fmt.Println("exec exited " + cmd.ProcessState.String() + " attempting to restart")
					current_rev.cmds[i] = RestartExec(cmd)
				}

			}

			project.Unlock()
		}

	}

}

func RestartExec(cmd *exec.Cmd) *exec.Cmd {
	newCmd := exec.Command(cmd.Path)
	newCmd.Args = cmd.Args
	newCmd.Dir = cmd.Dir
	newCmd.Stdout = cmd.Stdout
	newCmd.Stderr = cmd.Stderr

	go newCmd.Run()
	return newCmd

}

func StartServer() {
	// start the fishtank server
	// todo: detach

	// projects = make(map[string]Project)
	for j := range data.Projects {
		current_rev := data.Projects[j].Current_revision
		fmt.Println("starting " + j + " at " + "/srv/www/" + j + "/" + current_rev)
		dashConfig := dConfig.GetDashConfig("/srv/www/" + j + "/" + current_rev + "/")
		RunNewRev(j, current_rev, dashConfig)
	}

	// start health check
	go HealthCheck()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
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
	})

	Log.Log("Fishtank running at 127.0.0.1:" + config.Port)

	log.Fatal(http.ListenAndServe(":"+config.Port, nil))

}

func router(r types.Response) {
	//fmt.Println("router" + r.Req.URL.Path)

	switch r.Req.URL.Path {

	case "/post-receive":
		if len(r.Req.Form["project"]) == 0 || len(r.Req.Form["new_rev"]) == 0 || len(r.Req.Form["old_rev"]) == 0 || len(r.Req.Form["branch"]) == 0 {
			r.AddError("not enough params, expecting 127.0.0.1:6000/post-receive?project=project&old_rev=old_rev&new_rev=new_rev&branch=branch")
			r.Kill(422)
			return
		}

		project := r.Req.Form["project"][0]
		old_rev := r.Req.Form["old_rev"][0]
		new_rev := r.Req.Form["new_rev"][0]
		branch := r.Req.Form["branch"][0]

		// todo: create standardized go api for passing in r to functions and streaming messages back to the client before closing the connection.
		// note: the post-receive hook that curled this function streams its stdout to git
		PostReceive(project, old_rev, new_rev, branch)

		r.AddMessage("Deploying rev:" + new_rev + " for branch:" + branch)
		r.Kill(200)
	case "/add-project":
		if len(r.Req.Form["project"]) == 0 || len(r.Req.Form["branch"]) == 0 {
			r.AddError("not enough params, expecting 127.0.0.1:6000/add-project?project=project&branch=branch")
			r.Kill(422)
			return
		}
		project := r.Req.Form["project"][0]
		branch := r.Req.Form["branch"][0]
		AddProject(project, branch)
		r.AddMessage("Adding projects")
		r.Kill(200)
	case "/remove-project":
		if len(r.Req.Form["project"]) == 0 {
			r.AddError("not enough params, expecting 127.0.0.1:6000/remove-project?project=project")
			r.Kill(422)
			return
		}
		project := r.Req.Form["project"][0]
		RemoveProject(project)
		r.AddMessage("Removing project")
		r.Kill(200)
	case "/kill-project":
		// this api endpoint is deprecated
		if len(r.Req.Form["project"]) == 0 {
			r.AddError("not enough params, expecting 127.0.0.1:6000/kill-project?project=project")
			r.Kill(422)
			return
		}
		project := r.Req.Form["project"][0]
		KillProject(project)
		r.AddMessage("Killing project execs")
		r.Kill(200)
	default:
		r.AddError("Path: " + r.Req.URL.Path + " is not valid")
		r.Kill(404)
	}
}
