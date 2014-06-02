package project

import (
	"bytes"
	"code.highf.in/chalkhq/highfin/config"
	"code.highf.in/chalkhq/highfin/nodejs"
	"code.highf.in/chalkhq/highfin/types"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"

	//"strconv"
	//"time"
)

func Deploy(r types.Response) {
	account := r.Req.FormValue("account")
	project := r.Req.FormValue("project")
	branch := r.Req.FormValue("branch")

	if account == "" || project == "" || branch == "" {
		r.AddError("missing variables. Expected [account, project, branch], got [" + r.Req.FormValue("account") + ", " + r.Req.FormValue("project") + ", " + r.Req.FormValue("branch") + "]")
		r.Kill(422)
		return
	}
	fmt.Println(branch)

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
	fmt.Println(checkout_folder + "/")
	dashConfig := config.GetDashConfig(checkout_folder + "/")
	if len(dashConfig.Apps) == 0 {
		r.AddError("-.json file is empty or not found")
		r.Kill(500)
		//todo: cleanup on 500, maybe defers
		return
	}
	fmt.Println(dashConfig)

	for app_name := range dashConfig.Apps {
		app := dashConfig.Apps[app_name]
		app_folder := checkout_folder + "/" + app_name

		/*c := exec.Command(`ls`, app_folder+"/..")
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		_ = c.Run()*/

		_ = exec.Command(`cp`, checkout_folder+`/-.json`, app_folder+`/-.json`).Run()

		//_ = exec.Command(`mkdir`, `-p`, app_folder).Run() // this should already exist
		// the rest should probably be a goroutine:

		// copy in DockerFile and todo: jellyfish
		// todo: move Dockerfile to /shark/docker/Dockerfile

		//_ = exec.Command(`cp`, `/vagrant/octopus/docker/Dockerfile`, app_folder+"/Dockerfile").Run()
		// create docker file
		//newline :=
		docker_instructions := "FROM google/debian:wheezy\n" //note: must use "" instead of `` for \n to resolve to newline and not literally \n
		docker_instructions += "ADD . /code\n"

		for i := 0; i < len(app.Execs); i++ {
			appPart := app.Execs[i]

			if appPart.Lang == "nodejs" {

				// the docker image is actually built on shark, octopus doesn't need node.js versions installed
				nodejs.InstallNode(appPart.Version)
				_ = exec.Command(`mkdir`, `-p`, app_folder+"/__dep/n/").Run()
				_ = exec.Command(`cp`, `-r`, nodejs.BinFolder(appPart.Version), app_folder+`/__dep/n/`).Run()
				//docker_instructions += "ADD /usr/local/n/" + app.Version + " /usr/local/n/" + app.Version + "\n"
				fmt.Println("node folder:")
				/*c := exec.Command(`ls`, app_folder+`/__dep/n/`+app.Version+"/")
				c.Stderr = os.Stderr
				c.Stdout = os.Stdout
				_ = c.Run()
				fmt.Println("======")*/

			}

		}

		docker_instructions += "ENTRYPOINT /code/jellyfish " + app_name

		err := ioutil.WriteFile(app_folder+`/Dockerfile`, []byte(docker_instructions), 777)

		_ = exec.Command(`cp`, `/vagrant/go/bin/jellyfish`, app_folder+"/jellyfish").Run()
		// todo: remove the following line, this merely simulates an app
		//_ = exec.Command(`cp`, `/vagrant/go/bin/test`, app_folder+"/test").Run()

		// if err != nil {
		// 	r.AddError("failed to copy docker file")
		// 	r.Kill(500)
		// }

		tar_file := `/coral/` + account + `/` + project + `/` + branch + `/` + app_name + ".tar"

		// todo: run tests and build
		// todo: tests should only be run in the final docker container setup, so dev-next test that uses a database can actually access the database container

		// create docker file

		// assign deploy to sharks
		assigned_shark := "http://10.10.10.11"

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
		fmt.Println("attempting request")
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("failed to upload file to shark")

			r.Kill(500)
			return
		}

		body, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(body))
		fmt.Println("success")

		r.Kill(200)
	}

}
