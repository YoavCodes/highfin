package project

import "os"
import "os/exec"
import "code.highf.in/chalkhq/highfin/types"
import "strconv"
import "fmt"

//import "code.highf.in/guppy/util"

func Create(r types.Response) {
	fmt.Println("creating...")
	account := r.Req.FormValue("account")
	project := r.Req.FormValue("project")
	key := r.Req.FormValue("key")
	force, err := strconv.ParseBool(r.Req.FormValue("force"))

	// todo: a) better parsing/validation using regexes
	// 		 b) abstract away into shared library

	if account == "" || project == "" || key == "" || err != nil {
		r.AddError("missing variables. Expected [account, project, key, force], got [" + r.Req.FormValue("account") + ", " + r.Req.FormValue("project") + ", " + r.Req.FormValue("key") + ", " + r.Req.FormValue("force") + "]")
		r.Kill(422)
		return
	}

	project_user := account + "_" + project
	project_path := "/coral/" + account + "/" + project + "/"

	// if the project folder exists
	if _, err := os.Stat(project_path); err == nil {
		if force == true {
			_ = exec.Command("rm", "-rf", project_path).Run()
			_ = exec.Command("userdel", project_user).Run()
		} else {
			// kill the api query
			r.AddError("Project already exists") //todo: should be AddMessage
			r.Kill(200)
			return
		}
	}

	// add user
	_ = exec.Command("useradd", project_user, "-d", project_path, "-m").Run()

	_ = exec.Command("mkdir", "-p", project_path+".ssh").Run()

	_ = exec.Command("touch", project_path+".ssh/authorized_keys").Run()

	//_ = exec.Command("mkdir", "-p", project_path+"code.git").Run()

	// todo: guppy should have responsibility for creating a salmon app. the repo cloned here should
	//		only contain a -.json file and a readme or maybe default to salmon is the way to go
	//cmd := exec.Command("git", "clone", "--depth", "1", "https://github.com/YoavGivati/salmon", "-b", "dev-next", "/octopus/salmon/") //project_path+"code.git")
	_ = exec.Command("cp", "-r", "/octopus/salmon/.git", project_path+"code.git").Run()
	//_ = exec.Command("cp", "-R", "/octopus/salmon/.git", project_path+"code.git").Run()

	//cmd := exec.Command("git", "clone", "--bare", "https://github.com/YoavGivati/salmon", project_path+"code.git")
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	//_ = cmd.Run()

	//_ = exec.Command("rm", "-Rd", project_path+"code.git").Run()

	_ = exec.Command("chown", "-R", project_user+":"+project_user, project_path).Run()

	AddKey(account, project, key)

	r.AddError("Project succesfully created")
	r.Kill(200)

}
