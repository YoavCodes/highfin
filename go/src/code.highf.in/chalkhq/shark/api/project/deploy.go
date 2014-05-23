package project

import (
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
	_ = os.MkdirAll("/shark/tmp/chalkhq/highfin/dev-next", 777)

	dst, err := os.Create("/shark/tmp/chalkhq/highfin/dev-next.tar")

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

	cmd := exec.Command(`tar`, `-x`, `-C`, `/shark/tmp/chalkhq/highfin/dev-next`, `-f`, `/shark/tmp/chalkhq/highfin/dev-next.tar`)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		r.AddError("error extracting tar file")
		r.Kill(500)
		return
	}

	/// build the image
	cmd = exec.Command(`docker`, `build`, `-t`, `chalkhq_highfin_dev-next`, `/shark/tmp/chalkhq/highfin/dev-next`)
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
	err = exec.Command(`docker`, `rm`, `-f`, `chalkhq_highfin_dev-next`).Run()

	fmt.Println(err)

	cmd = exec.Command(`docker`, `run`, `-p`, `50000:8081`, `--name=chalkhq_highfin_dev-next`, `chalkhq_highfin_dev-next`)
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
