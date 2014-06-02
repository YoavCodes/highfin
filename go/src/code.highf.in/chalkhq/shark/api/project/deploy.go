package project

/*
	We want Octopus to keep a docker repository running.
	ideally implemented in golang integrated into the octopus source.
	In it we store every version of node + jellyfish.
	- deploying a new version of jellyfish will mean, replacing the jellyfish binary
	in each of those images on all the octopi.
	- the optimization here being uploading context, since node and jellyfish will already be an image
	- however note: you're still downloading the image with node and jellyfish from
	the octopus private docker repo. so it adds lots of overhead and code for managing the
	repo and images and rebuilding the images every time jellyfish is upgraded.
	but it saves a few seconds on deploys by minimizing the context to uplaod

	note: we fundamentally want all building to occur on sharks and not octopus. we want octopus
	ram, cpu, and disk space to be used as close to zero as possible. when scaling to thousands
	of sharks, we can't put the burden on octopus, it's fundamentally not scalable, AND
	the sharks would sit there with unused resources.

	note: docker will re-run any containers on reboot. so shark doesn't have to micromanage.

	Shark: keep track of how many free resources it has, octopus can keep track of which
	containers are deployed on which sharks.
*/

import (
	"code.highf.in/chalkhq/highfin/config"
	//"code.highf.in/chalkhq/highfin/nodejs"
	"code.highf.in/chalkhq/highfin/types"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func Deploy(r types.Response) {
	fmt.Println("deploying..")
	fmt.Println(r.Req)

	r.Req.ParseMultipartForm(64)

	app_name := r.Req.MultipartForm.Value["app_name"][0]
	//fmt.Println("app name: " + app_name)

	file, err := r.Req.MultipartForm.File["tar"][0].Open()
	if err != nil {
		r.AddError("error opening uploaded file")
		r.Kill(500)
		return
	}

	/*fmt.Println("here pointer to file: ", r.Req.MultipartForm.File["tar"])
	// note: will automatically replace the existing parts
	_ = exec.Command(`mkdir`, `-p`, "/shark/tmp/chalkhq/highfin/dev-next").Run()

	fmt.Println("attempting untar")
	// directly untar into drive
	cmd := exec.Command(`tar`, `-x`, `-C`, `/shark/tmp/chalkhq/highfin/dev-next`)
	in_pipe, err := cmd.StdinPipe()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	io.Copy(in_pipe, file)
	_ = cmd.Run()

	fmt.Println("untar success")*/

	// todo: stream into untar command
	// save file to disk

	_ = os.RemoveAll("/shark/tmp/chalkhq/highfin/dev-next")
	_ = os.RemoveAll("/shark/tmp/chalkhq/highfin/dev-next.tar")
	_ = os.MkdirAll("/shark/tmp/chalkhq/highfin/dev-next/salmon", 777)

	dst, err := os.Create("/shark/tmp/chalkhq/highfin/dev-next/salmon.tar")

	defer dst.Close()

	if err != nil {
		r.AddError("error creating tmp tar file")
		r.Kill(500)
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		r.AddError("error populating tmp tar file")
		r.Kill(500)
		return
	}

	cmd := exec.Command(`tar`, `-x`, `-C`, `/shark/tmp/chalkhq/highfin/dev-next/salmon`, `-f`, `/shark/tmp/chalkhq/highfin/dev-next/salmon.tar`)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		r.AddError("error extracting tar file")
		r.Kill(500)
		return
	}

	// get config
	dashConfig := config.GetDashConfig(`/shark/tmp/chalkhq/highfin/dev-next/salmon`)
	app := dashConfig.Apps[app_name]
	fmt.Println(app)

	// switch app.Lang {
	// case "nodejs":
	// 	nodejs.InstallNode(app.Version)
	// 	_ = exec.Command(`cp`, nodejs.BinFolder, `/shark/tmp/chalkhq/highfin/dev-next/salmon`).Run()
	// }

	/// build the image
	cmd = exec.Command(`docker`, `build`, `-t`, `chalkhq_highfin_dev-next_salmon`, `/shark/tmp/chalkhq/highfin/dev-next/salmon`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		r.AddError("failed to build image")
		r.Kill(500)
		return
	}

	fmt.Println("running container")
	//cmd = exec.Command(`docker`, `-d`, `run`, `chalkhq_highfin_dev-next`)
	// todo: this should have more advnaced unique naming, and remove the previous image after
	err = exec.Command(`docker`, `rm`, `-f`, `chalkhq_highfin_dev-next_salmon`).Run()

	fmt.Println(err)

	cmd = exec.Command(`docker`, `run`, `-p`, `50000:8081`, `--name=chalkhq_highfin_dev-next_salmon`, `chalkhq_highfin_dev-next_salmon`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		fmt.Println(err)
		r.AddError("failed to run container")
		r.Kill(500)
		return
	}

	r.Kill(200)

	// run the container
}
