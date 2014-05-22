package project

import (
	"bytes"
	"code.highf.in/chalkhq/highfin/config"
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
	// todo: maybe change the tmp path for easier cleanup

	_ = exec.Command(`mkdir`, `-p`, checkout_folder).Run()
	//defer exec.Command(`rm`, `-R`, checkout_folder).Run()

	// get the latest dev-next branch (git exports it as a tar file) and untar it to the temporary directory
	cmd := exec.Command(`git`, `archive`, `--remote`, repo_folder, branch)

	cmd2 := exec.Command(`tar`, `-x`, `-C`, checkout_folder)

	cmd2.Stdin, _ = cmd.StdoutPipe()

	_ = cmd2.Start()
	_ = cmd.Start()
	_ = cmd2.Wait()

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
	dashConfig := config.GetDashConfig()
	fmt.Println(dashConfig)

	// copy in DockerFile and todo: jellyfish
	// todo: move Dockerfile to /shark/docker/Dockerfile

	_ = exec.Command(`cp`, `/vagrant/docker/Dockerfile`, checkout_folder+"/Dockerfile").Run()
	_ = exec.Command(`cp`, `/vagrant/go/bin/jellyfish`, checkout_folder+"/jellyfish").Run()
	// todo: remove the following line, this merely simulates an app
	_ = exec.Command(`cp`, `/vagrant/go/bin/test`, checkout_folder+"/test").Run()

	// if err != nil {
	// 	r.AddError("failed to copy docker file")
	// 	r.Kill(500)
	// }

	tar_file := `/coral/` + account + `/` + project + `/` + branch + ".tar"

	// todo: run tests and build
	// todo: tests should only be run in the final docker container setup, so dev-next test that uses a database can actually access the database container

	// create docker file

	// assign deploy to sharks
	assigned_shark := "http://10.10.10.10"

	// tar branch for re-deploys
	exec.Command(`rm`, `-R`, tar_file).Run()
	_ = exec.Command(`tar`, `-c`, `-f`, tar_file, `-C`, checkout_folder, `.`).Run()

	// upload
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	f, err := os.Open(tar_file)
	if err != nil {
		r.Kill(500)
		return
	}
	fw, err := w.CreateFormFile("tar", branch+`.tar`)
	if err != nil {
		r.Kill(500)
		return
	}

	if _, err = io.Copy(fw, f); err != nil {
		r.Kill(500)
		return
	}

	w.Close()

	req, err := http.NewRequest("POST", assigned_shark+`/project/deploy`, &b)

	req.Header.Set("Content-Type", w.FormDataContentType())
	fmt.Println("attempting request")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("failed")

		r.Kill(500)
		return
	}

	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
	fmt.Println("success")

	r.Kill(200)

}
