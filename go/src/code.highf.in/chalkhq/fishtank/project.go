package main

import (
	"bufio"
	"code.highf.in/chalkhq/shared/command"
	"code.highf.in/chalkhq/shared/persistence"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

// note: when removing a project, we only need to kill execs if there's a running fishtank server
func Remove() {
	// removes a project

	var project string
	if len(os.Args) > 2 {
		project = os.Args[2]
	} else {
		fmt.Println("Enter a project name, should be the domain name without protocol eg: domain.com or staging.domain.com :")
		fmt.Scanf("%s", &project)
	}

	_ = command.E("rm -R /srv/logs/" + project + ".log").Run()
	_ = command.E("rm -R /srv/www/" + project).Run()
	_ = command.E("rm -R /etc/nginx/sites-enabled/" + project + ".conf").Run()
	_ = command.E("rm -R /srv/coral/" + project).Run()

	_ = command.E("nginx -s reload").Run()

	_ = command.E("userdel -f " + project).Run()

	// if fishtank server is running ping it to take down the running execs related to the project
	_, err := http.Get("http://127.0.0.1:" + config.Port + "/remove-project?project=" + project)
	if err == nil {
		RemoveProject(project)
	}
}

// web api handler
func AddProject(project string, branch string) {
	fmt.Println("adding project")
	if len(data.Projects) == 0 {
		data.Projects = make(map[string]Project)
	}
	// todo: branch, current_revision, is not being set! why? wrong syntax probably
	data.Projects[project] = Project{
		Branch:           branch,
		Current_revision: "default",
		Revisions:        make(map[string]Revision),
	}

	persistence.SaveData(data, DATA_JSON)
}

// web api handler
func RemoveProject(project string) {
	fmt.Println("remove project")

	for i := range data.Projects[project].Revisions {
		KillOldRev(project, i)
	}

	delete(data.Projects, project)

	persistence.SaveData(data, DATA_JSON)
}

// command line usage. will try pinging the command-line helper to update the project list
// note: when adding a project it configures a default project that's just an html file with no execs
func Add() {

	// note: if the project is test then the git url will be
	// ssh://test@10.10.10.200/~/code.git
	fmt.Println("Adding a new project...")

	var project string
	if len(os.Args) > 2 {
		project = os.Args[2]
	} else {
		fmt.Println("Enter a project name, should be the domain name without protocol eg: domain.com or staging.domain.com :")
		fmt.Scanf("%s", &project)
	}

	var branch string
	fmt.Println("Enter the branch name, you'd like deployed eg: prod")
	fmt.Scanf("%s", &branch)

	fmt.Println("Enter your public key:")
	in := bufio.NewReader(os.Stdin)
	key, _ := in.ReadString('\n')

	if project == "" || key == "" {
		fmt.Println("missing variables. Expected [project, key]")
		return
	}

	if project == "default" {
		fmt.Println("Please choose a valid project name")
		return
	}

	project_user := project
	coral_path := "/srv/coral/" + project + "/"

	// if the project folder exists
	if _, ok := data.Projects[project]; ok == true {
		// kill the api query
		fmt.Println("Project already exists") //todo: should be AddMessage
		return
	}

	// add user
	_ = exec.Command("useradd", project_user, "-d", coral_path, "-m").Run()

	_ = exec.Command("mkdir", "-p", coral_path+".ssh").Run()

	_ = exec.Command("touch", "/srv/logs/"+project+".log").Run()

	_ = exec.Command("touch", coral_path+".ssh/authorized_keys").Run()

	_ = exec.Command("mkdir", "-p", coral_path+"/code.git").Run()

	_ = exec.Command("mkdir", "-p", "/srv/www/"+project).Run()

	_ = exec.Command("mkdir", "-p", coral_path+"/data").Run()
	_ = exec.Command("chown", project+":"+project, coral_path+"/data").Run()

	_ = exec.Command("touch", "/etc/nginx/sites-enabled/"+project+".conf").Run()

	_ = exec.Command("chown", project_user+":"+project_user, "/etc/nginx/sites-enabled/"+project+".conf").Run()

	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = coral_path + "/code.git"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	CreatePostRecieveHook(project, branch)
	// create default revision
	CreateDefaultRevision(project, branch)
	default_rev_path := "/srv/www/" + project + "/default"
	// create branch
	cmd = exec.Command("git", "--work-tree="+default_rev_path, "checkout", "-b", branch, "-f")
	cmd.Dir = coral_path + "code.git"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	// add all files in default revision
	cmd = exec.Command("git", "--work-tree="+default_rev_path, "add", "-A")
	cmd.Dir = coral_path + "code.git"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	// do first commit
	cmd = exec.Command("git", "--work-tree="+default_rev_path, "commit", "-m", "First deploy")
	cmd.Dir = coral_path + "code.git"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	CreateNginxSiteConf(project)

	_ = exec.Command("chmod", "ug+x", coral_path+"/code.git/hooks/post-receive").Run()

	AddKey(project, key)

	_ = exec.Command("chown", "-R", project_user+":"+project_user, coral_path).Run()
	_ = exec.Command("chown", "-R", project_user+":"+project_user, "/srv/www/"+project).Run()

	// if fishtank is running in the background somewhere, try update it's project list
	_, err := http.Get("http://127.0.0.1:" + config.Port + "/add-project?project=" + project + "&branch=" + branch)
	if err != nil {
		AddProject(project, branch)
	}

	_ = exec.Command("nginx", "-s", "reload").Run()

	fmt.Println("Project succesfully created")
}

func AddKey(project string, key string) {
	coral_path := "/srv/coral/" + project + "/"

	authorized_keys, err := os.Create(coral_path + ".ssh/authorized_keys")
	defer authorized_keys.Close()

	if err != nil {
		fmt.Println("error: " + err.Error())
		return
	}

	authorized_keys.WriteString(key + "\n\n")
}
