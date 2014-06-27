package project

import (
	"bytes"
	"code.highf.in/chalkhq/shared/config"
	"code.highf.in/chalkhq/shared/nodejs"
	"code.highf.in/chalkhq/shared/types"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"

	"strings"
	"time"
)

/*
	TODO: pass mesh in from main(). we want a global Mesh object as it's responsible for managing shark health as well.
*/

func Deploy(r types.Response, mesh *types.Mesh) {
	account := r.Req.FormValue("account")
	project := r.Req.FormValue("project")
	branch := r.Req.FormValue("branch")

	project_name := account + "_" + project

	//mesh.Projects = make(map[string]*Project)
	//mesh.Projects[account+"_"+project] = Project

	// todo: persist to db or disk. until then we just build the object and then keep it in memory on each deploy.
	// TODO:: Move this to the main() function and pass mesh into Deploy fu

	if mesh.Projects[project_name] == nil {
		//mesh.Projects = make(map[string]*types.Project)
		mesh.Projects[project_name] = &types.Project{}
		mesh.Projects[project_name].Info.GITrepo = "/coral/" + account + "/" + project + "/code.git"
	}
	//mesh.Sharks["10.10.10.11"].Info.ports = make([]string)

	//mesh.Projects[project_name].Info = &ProjectInfo{}

	// fmt.Println(mesh.Projects[account+"_"+project].Info.GITrepo)

	if account == "" || project == "" || branch == "" {
		r.AddError("missing variables. Expected [account, project, branch], got [" + r.Req.FormValue("account") + ", " + r.Req.FormValue("project") + ", " + r.Req.FormValue("branch") + "]")
		r.Kill(422)
		return
	}
	// fmt.Println(branch)

	// randomly generate a folder
	checkout_folder := `/octopus/tmp/` + account + `/` + project + `/` + branch
	repo_folder := `/coral/` + account + `/` + project + `/code.git`
	repo_branch_folder := `/coral/` + account + `/` + project + `/` + branch
	_ = exec.Command(`mkdir`, `-p`, repo_branch_folder).Run()

	// todo: maybe change the tmp path for easier cleanup

	// check if the repo exists
	if _, err := os.Stat(repo_folder); os.IsNotExist(err) {
		r.AddError("repo does not exist, create the project first")
		r.Kill(500)
		return
	}
	_ = exec.Command(`mkdir`, `-p`, checkout_folder).Run()
	//defer exec.Command(`rm`, `-R`, checkout_folder).Run()

	// get the latest dev-next branch (git exports it as a tar file) and untar it to the temporary directory
	cmd := exec.Command(`git`, `archive`, `--remote`, repo_folder, branch)

	cmd2 := exec.Command(`tar`, `-x`, `-C`, checkout_folder)

	cmd2.Stdin, _ = cmd.StdoutPipe()

	_ = cmd2.Start()
	err := cmd.Start()
	_ = cmd2.Wait()

	if err != nil {
		r.AddError("failed to archive repo " + branch)
		r.Kill(500)
	}

	/*
		todo: you ideally want to build the container on the shark it's being deployed on that way if you have an app and a database you can build an image with only the relevant subfolders of your git repo
		ie: you have a salmon folder and an mongo folder, then in -.json you can have apps named by the folder name.
		the mongo folder only has a fixtures subfolder
		the salmon folder has the app
		a memcached -.json entry doesn’t need a folder.
		so then you get a tar file for every piece of the deploy
		you need to add a Dockerfile to every tar
	*/
	// at this point there's an untarred export of the latest of specified branch

	// get the dashconfig
	// fmt.Println(checkout_folder + "/")
	mesh.Projects[account+"_"+project].Temp = config.GetDashConfig(checkout_folder + "/")
	// fmt.Println(mesh.Projects[project_name].Temp.Apps)
	if len(mesh.Projects[project_name].Temp.Apps) == 0 {
		r.AddError("-.json file is empty or not found")
		r.Kill(500)
		mesh.Projects[project_name].Temp = config.DashConfig{}
		//todo: cleanup on 500, maybe defers
		return
	}
	// fmt.Println(mesh.Projects[project_name].Temp)

	// todo: at this point it should find a shark, and assign the sharkports and read the domains /ports from json object
	// preceding the deploy-loop below, squid should be notified with new routing information and activate private routing

	// sharkports for the whole deploy
	sharkports := make(map[string]string)

	for app_name := range mesh.Projects[project_name].Temp.Apps {

		app := mesh.Projects[project_name].Temp.Apps[app_name]
		app_folder := checkout_folder + "/" + app_name
		// todo: should loop for each instance of the app, but only loop the upload request, the other build stuff should not be repeted for each instance.

		//sharkport_port := strconv.FormatInt(int64(len(mesh.Sharks["10.10.10.11"].Info.Ports))+int64(5000), 10)

		/*c := exec.Command(`ls`, app_folder+"/..")
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		_ = c.Run()*/

		_ = exec.Command(`mkdir`, `-p`, app_folder).Run()

		_ = exec.Command(`cp`, checkout_folder+`/-.json`, app_folder+`/-.json`).Run()

		//_ = exec.Command(`mkdir`, `-p`, app_folder).Run() // this should already exist
		// the rest should probably be a goroutine:

		// copy in DockerFile and todo: jellyfish
		// todo: move Dockerfile to /shark/docker/Dockerfile

		//_ = exec.Command(`cp`, `/vagrant/octopus/docker/Dockerfile`, app_folder+"/Dockerfile").Run()
		// create docker file
		//newline :=
		// todo: we need to create a proper docker basebox
		docker_instructions := ""

		for i := 0; i < len(app.Execs); i++ {
			appPart := app.Execs[i]

			if appPart.Lang == "nodejs" {

				docker_instructions += "FROM debian:7.4\n" //note: must use "" instead of `` for \n to resolve to newline and not literally \n
				docker_instructions += "ADD . /code\n"
				docker_instructions += "ENTRYPOINT /code/jellyfish " + app_name

				// the docker image is actually built on shark, octopus doesn't need node.js versions installed
				nodejs.InstallNode(appPart.Version)
				_ = exec.Command(`mkdir`, `-p`, app_folder+"/__dep/n/").Run()
				_ = exec.Command(`cp`, `-r`, nodejs.BinFolder(appPart.Version), app_folder+`/__dep/n/`).Run()
				//docker_instructions += "ADD /usr/local/n/" + app.Version + " /usr/local/n/" + app.Version + "\n"
				// fmt.Println("node folder:")
				/*c := exec.Command(`ls`, app_folder+`/__dep/n/`+app.Version+"/")
				c.Stderr = os.Stderr
				c.Stdout = os.Stdout
				_ = c.Run()
				fmt.Println("======")*/

			} else if appPart.Lang == "golang" {
				docker_instructions += "FROM debian:7.4\n" //note: must use "" instead of `` for \n to resolve to newline and not literally \n
				docker_instructions += "ADD . /code\n"
				docker_instructions += "ENTRYPOINT /code/jellyfish " + app_name
			} else if appPart.Lang == "mongodb" {
				//docker_instructions += "FROM debian:7.4\n"
				docker_instructions += "FROM ubuntu:12.04\n"
				docker_instructions += "ADD . /code\n"
				docker_instructions += "RUN echo 'deb http://archive.ubuntu.com/ubuntu precise main universe' > /etc/apt/sources.list\n"
				docker_instructions += "RUN apt-get -y update\n"
				docker_instructions += "RUN apt-key adv --keyserver keyserver.ubuntu.com --recv 7F0CEB10\n"
				docker_instructions += "RUN echo 'deb http://downloads-distro.mongodb.org/repo/ubuntu-upstart dist 10gen' | tee -a /etc/apt/sources.list.d/10gen.list\n"
				docker_instructions += "RUN apt-get update\n"
				docker_instructions += "RUN apt-get -y install apt-utils\n"
				docker_instructions += "RUN apt-get -y install mongodb-10gen\n"
				// todo: this should be an attached volume
				docker_instructions += "RUN mkdir -p /data/db\n"
				//docker_instructions += "RUN \n"
				//docker_instructions += "ENTRYPOINT /code/jellyfish " + app_name
				docker_instructions += "ENTRYPOINT /usr/bin/mongod --smallfiles\n"
			} else {
				// todo: errors should cleanup steps that have already happened
				r.AddError("-.json file misconfiguration")
				r.AddError("app " + app_name + " has Lang " + appPart.Lang + ", expecting one of [nodejs, golang, mongodb]")
				r.Kill(422)
				return
			}

		}

		// todo: rewrite the docker file from docker_instructions + instance name for each instance. to pass in the instance id to each one

		err := ioutil.WriteFile(app_folder+`/Dockerfile`, []byte(docker_instructions), 777)

		_ = exec.Command(`cp`, `/vagrant/go/bin/jellyfish`, app_folder+"/jellyfish").Run()
		// todo: remove the following line, this merely simulates an app
		//_ = exec.Command(`cp`, `/vagrant/go/bin/test`, app_folder+"/test").Run()

		// if err != nil {
		// 	r.AddError("failed to copy docker file")
		// 	r.Kill(500)
		// }

		tar_file := `/coral/` + account + `/` + project + `/` + branch + `/` + app_name + ".tar"

		// tar branch for re-deploys
		_ = exec.Command(`rm`, `-R`, tar_file).Run() // todo: only remove this on new code, a plain deploy should use the existing tar files

		cmd = exec.Command(`tar`, `-c`, `-f`, tar_file, `-C`, app_folder, `.`)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Run()
		if err != nil {
			r.AddError("failed to create tar file")
			r.Kill(500)
			return
		}

		// todo: run tests and build
		// todo: tests should only be run in the final docker container setup, so dev-next test that uses a database can actually access the database container

		// create docker file
		app.Sharkports = make(map[string]string)
		app.Deploys = make(map[string]string)

		// todo: should loop through all apps preparing them, then start another app loop for instances // interacting with shark via goroutines

		for i := 0; i < len(app.Instances); i++ {
			jellyport := app.Instances[i]

			//fmt.Println(mesh.Sharks["10.10.10.11"].Info.Ports)
			// assign deploy to sharks
			assigned_shark := "http://" + mesh.Sharks["10.10.10.11"].Info.Ip

			// upload
			var b bytes.Buffer
			w := multipart.NewWriter(&b)

			// todo: this should be a json struct, related to mesh
			// app_name_field, err := w.CreateFormField("app_name")
			// if err != nil {
			// 	r.AddError("failed to create app_name field")
			// 	r.Kill(500)
			// 	return
			// }
			//app_name_field.Write([]byte(app_name))
			instanceID := strconv.FormatInt(time.Now().UnixNano(), 10)
			_ = w.WriteField("instanceID", instanceID)
			//_ = w.WriteField("sharkport_port", sharkport_port)
			_ = w.WriteField("account_name", account)
			_ = w.WriteField("project_name", project)
			_ = w.WriteField("env_name", branch) // todo: disambiguate branch vs. env, they mean different things
			_ = w.WriteField("app_name", app_name)

			// write the tar file to the upload request body
			f, err := os.Open(tar_file)
			if err != nil {
				r.AddError("failed to open tar file")
				r.Kill(500)
				return
			}

			fw, err := w.CreateFormFile("tar", branch+`.tar`)
			if err != nil {
				r.AddError("failed to create form file")
				r.Kill(500)
				return
			}

			if _, err = io.Copy(fw, f); err != nil {
				r.AddError("failed to populate form file")
				r.Kill(500)
				return
			}

			// todo: this should probably happen after the request. re: ensure we're streaming bytes and not copying the whole tar file into ram
			w.Close()

			// make the request
			req, err := http.NewRequest("POST", assigned_shark+`/project/deploy`, &b)

			req.Header.Set("Content-Type", w.FormDataContentType())
			// fmt.Println("attempting request")
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				fmt.Println("failed to upload file to shark")

				r.Kill(500)
				mesh.Projects[project_name].Temp = config.DashConfig{}
				return
			}

			body, _ := ioutil.ReadAll(res.Body)

			// fmt.Println(string(body))

			apiRes := ApiRes{}

			err = json.Unmarshal(body, &apiRes)

			if err != nil {
				fmt.Println("failed to unmarshal response")
				fmt.Println(err)
			}

			// fmt.Println(apiRes.Meta.Status)

			// fmt.Println(apiRes)
			if apiRes.Data["sharkport_port"] == nil {
				r.AddError("error: shark failed")
				r.Kill(500)
				mesh.Projects[project_name].Temp = config.DashConfig{}
				return
			}

			var sharkport_port string = apiRes.Data["sharkport_port"].(string)
			sharkport := mesh.Sharks["10.10.10.11"].Info.Ip + ":" + sharkport_port

			// fmt.Println("sharkport: " + sharkport)
			// fmt.Println("instanceID: " + instanceID)
			// fmt.Println("jellyport: " + jellyport)
			// if success

			app.Sharkports[jellyport] = sharkport
			sharkports[jellyport] = sharkport

			app.Deploys[instanceID] = sharkport

			// fmt.Println("SETTING")
			// fmt.Println(app.Deploys)
			// fmt.Println(len(app.Deploys))
			// fmt.Println(mesh.Projects[project_name].DEVnext.Apps[app_name].Deploys)

			for domain_key := range app.Domains {
				if len(app.Domains[domain_key]) == 0 {
					//app.Domains[domain_key] = make([]string, len(app.Instances))
				}
				app.Domains[domain_key] = append(app.Domains[domain_key], sharkport)
				// fmt.Println("--")
				// fmt.Println(app.Domains[domain_key])
			}

			// todo: not sure we need this now that the shark is selecting available ports automatically
			//fmt.Println(mesh.Sharks["10.10.10.11"].Info.Ports)
			//mesh.Sharks["10.10.10.11"].Info.Ports = append(mesh.Sharks["10.10.10.11"].Info.Ports, sharkport_port)
			//fmt.Println(mesh.Sharks["10.10.10.11"].Info.Ports)
		}

		mesh.Projects[project_name].Temp.Apps[app_name] = app

	}
	// todo: this part should wait for all success signals from jellyfish, for now we assume the deploy was successfull

	// ping each jellyfish with sharkports of others in the deploy
	sharkportJSON, _ := json.Marshal(sharkports)

	time.Sleep(30 * time.Second)

	for jellyport := range sharkports {
		//sharkport_ip := strings.SplitAfter(sharkports[jellyport], ":")[0]
		sharkport := sharkports[jellyport]

		res, err := http.PostForm("http://"+sharkport+"/___api", url.Values{"sharkports": {string(sharkportJSON)}})
		if err != nil {
			fmt.Println("request error: ", err.Error())
			return
		}

		body, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(body))
		res.Body.Close()
	}

	// todo: send signal to squid to enable domain switch
	// iptables -A FORWARD -p tcp -d 127.0.0.1 --dport 6101 -m state --state NEW,ESTABLISHED,RELATED -j ACCEPT
	// iptables -t nat -A PREROUTING -p tcp -d 9.9.9.9 --dport 6105 -j DNAT --to 10.10.10.11:49190
	// iptables -t nat -A POSTROUTING -p tcp --dport 6106 -o eth0 -j SNAT --to-source 10.10.10.11:49190

	var domainMap map[string][]string

	domainMap = make(map[string][]string)

	for app_name := range mesh.Projects[project_name].Temp.Apps {
		// fmt.Println("=====" + app_name)
		app := mesh.Projects[project_name].Temp.Apps[app_name]
		for domain_name := range app.Domains {
			// fmt.Println("=======" + domain_name)
			domainMap[domain_name] = app.Domains[domain_name]
		}
	}

	domainMapJSON, _ := json.Marshal(domainMap)
	// fmt.Println("domainMapJSON_+_+_+")
	// fmt.Println(string(domainMapJSON))

	res, err := http.PostForm("http://10.10.10.5:8282/route/update", url.Values{"account": {account}, "project": {project}, "branch": {branch}, "domains": {string(domainMapJSON)}})
	if err != nil {
		fmt.Println("request error: ", err.Error())
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
	// todo: convert dev-next branch variable to DEVnext Project env name so it works cross env
	// fmt.Println("TEETS")
	// fmt.Println(mesh.Projects[project_name].DEVnext)
	// fmt.Println(mesh.Projects[project_name].Temp)
	// issue take down to sharks where devnext apps are hosted.
	for app_name := range mesh.Projects[project_name].DEVnext.Apps {
		// fmt.Println("))))removing")
		// fmt.Println(app_name)
		app := mesh.Projects[project_name].DEVnext.Apps[app_name]
		// fmt.Println("GETTING")
		// fmt.Println(app.Deploys)
		// fmt.Println(len(app.Deploys))
		for instanceID := range app.Deploys {
			// fmt.Println(instanceID)
			sharkport := app.Deploys[instanceID]

			sharkport_ip := strings.SplitAfter(sharkport, ":")[0]
			// todo: send request to shark to take down
			res, err := http.PostForm("http://"+sharkport_ip+"80/project/remove", url.Values{"instanceID": {instanceID}})
			if err != nil {
				fmt.Println("failed to remove instance error: ", err.Error())
				return
			}
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			fmt.Println(string(body))
		}
	}

	// clean up Temp env.
	mesh.Projects[project_name].DEVnext = mesh.Projects[project_name].Temp

	// fmt.Println("TEST")
	// fmt.Println(mesh.Projects[project_name].DEVnext)
	// fmt.Println(mesh.Projects[project_name].Temp)

	mesh.Projects[project_name].Temp = config.DashConfig{}
	// fmt.Println("TOAST")
	// fmt.Println(mesh.Projects[project_name].DEVnext)
	// fmt.Println(mesh.Projects[project_name].Temp)

	fmt.Println("success")
	// todo: remove old docker images

	r.Kill(200)

}

type ApiRes struct {
	Meta struct {
		Status int      `json:"status"`
		Errors []string `json:"errors"`
	} `json:"meta"`
	Data map[string]interface{} `json:"data"`
}

// type Env struct {
// 	AppParts map[string]AppPart
// }

// type AppPart struct {
// 	Execs      []config.Exec     // could differ between envs.
// 	Sharkports map[string]string // map[instance]sharkport
// 	Domains    map[string]string // map[domain]sharkport
// }

/*

	Deploys[chalkhq_highfin] {
		info stuct {
			git_repo string // /coral/chalkhq/highfin/code.git

		}

		dev-next struct {
			env string // dev-next
			appParts struct { // web
				lang string //nodejs
			}
			domains map[domain_name][["private_ip", "private_ip2"]string] // public facing domains to internal ips
			ports map[domain_name][["private_ip", "private_ip2"]string] // private mapping for jellyfish proxy -.json specified private ports to internal ips:ports that may exist on other
		}
		qa-next etc..
		can later purchase
		custom-feature {}

	}


*/
