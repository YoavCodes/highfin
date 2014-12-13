package main

import (
	"code.highf.in/chalkhq/shared/command"
	dConfig "code.highf.in/chalkhq/shared/config"
	"code.highf.in/chalkhq/shared/nodejs"
	"code.highf.in/chalkhq/shared/persistence"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func RunNewRev(project string, newrev string, dashConfig dConfig.DashConfig) {

	newrevision := data.Projects[project].Revisions[newrev]

	for j := range dashConfig.Apps {
		app := dashConfig.Apps[j]
		newrevision.cmds = make([]*exec.Cmd, len(app.Execs))

		for i := 0; i < len(app.Execs); i++ {
			appPart := app.Execs[i]

			switch appPart.Lang {

			// todo: golang and binary apps require changes to -.json
			case "nodejs":
				fmt.Println("running node")
				path := `/srv/www/` + project + `/` + newrev + "/"
				nodejs.InstallNode(appPart.Version)
				mainpath, _ := filepath.Abs(path + appPart.Main)
				node := command.E(nodejs.BinPath(appPart.Version) + ` ` + mainpath)
				node.Dir = `/srv/www/` + project + `/` + newrev + "/"
				newrevision.cmds[i] = node
				go node.Run()
			}
		}
	}
	data.Projects[project].Revisions[newrev] = newrevision

}

func KillProject(project string) {
	for i := range data.Projects[project].Revisions {
		KillOldRev(project, i)
	}
	delete(data.Projects, project)
	persistence.SaveData(data, DATA_JSON)
}

func KillOldRev(project string, oldrev string) {
	cmds := data.Projects[project].Revisions[oldrev].cmds
	for i := 0; i < len(cmds); i++ {
		// todo: should check to see if process is still running first
		if cmds[i] != nil {
			//_, process_name := filepath.Split(cmds[i].Path)
			cmds[i].Process.Kill()
		}
	}

	delete(data.Projects[project].Revisions, oldrev)

}

func PostReceive(project string, old_rev string, new_rev string, branch string) {
	/*
		if production
			mkdir /srv/www/project/newrev
			git --work-tree=$new_rev_path checkout production -f
			if forever
				find an available port
				sed /etc/nginx/sites-enabled/project.conf updating the site root to newrev
				sed /etc/nginx/sites-enabled/project.conf updating the port
				sed /srv/www/project/newref/config/config.js updating the port to run on
				forever start --uid "newref" /srv/www/project/newref/salmon.js
				sleep 5
				forever stop --uid "oldref"
			if docker
				read -.json file for port mapping
				copy/fetch required node or golang lib
				build docker container attaching the data folder /srv/data/project
				run nginx or some other static fileserver inside docker container
			sudo nginx -s reload
			rm -R /srv/www/project/oldrev
	*/

	www_path := "/srv/www/" + project + "/"
	new_rev_path := www_path + new_rev
	//log_path := "/srv/logs/" + project + ".log"
	git_path := "/srv/coral/" + project + "/code.git"

	old_rev_path := www_path + old_rev

	// checkout latest changes to prod from repo to newrev folder
	_ = command.E("mkdir -p " + new_rev_path).Run()
	fmt.Println("git" + " --work-tree=" + new_rev_path + " checkout " + branch + " -f")
	cmd_git := exec.Command("git", "--work-tree="+new_rev_path, "checkout", branch, "-f")
	cmd_git.Dir = git_path
	cmd_git.Stdout = os.Stdout
	cmd_git.Stderr = os.Stderr
	cmd_git.Run()

	dashConfig := dConfig.GetDashConfig(new_rev_path + "/")

	_ = command.E("mkdir -p " + new_rev_path).Run()
	// have os give us a free port
	find_free, _ := net.Listen("tcp", ":0")
	free := strings.Split(find_free.Addr().String(), ":")
	find_free.Close()

	port := free[len(free)-1]

	nginx_conf_path := "/etc/nginx/sites-enabled/" + project + ".conf"
	tmp_sed_path := "/srv/coral/" + project + "/sed.tmp"
	app_conf_path := "/srv/www/" + project + "/" + new_rev + "/config/config.js"

	// update nginx configuration

	cmd := exec.Command(`sed`, `s:root.*;:root /srv/www/`+project+"/"+new_rev+`/public/;:`, nginx_conf_path)
	cmd2 := exec.Command(`tee`, tmp_sed_path)

	cmd2.Stdin, _ = cmd.StdoutPipe()

	_ = cmd2.Start()
	_ = cmd.Start()
	_ = cmd2.Wait()
	_ = cmd.Wait()

	_ = command.E("cp " + tmp_sed_path + " " + nginx_conf_path).Run()

	cmd3 := exec.Command(`sed`, `s+proxy_pass http://.*;+proxy_pass http://127.0.0.1:`+port+`;+`, nginx_conf_path)
	cmd4 := exec.Command(`tee`, tmp_sed_path)

	cmd4.Stdin, _ = cmd3.StdoutPipe()

	_ = cmd4.Start()
	_ = cmd3.Start()
	_ = cmd4.Wait()
	_ = cmd3.Wait()

	_ = command.E("cp " + tmp_sed_path + " " + nginx_conf_path).Run()

	// update app configuration
	// we have permissions for this folder, so just use sed inline
	_ = command.E("sed -i s/port:.*/port:" + port + "/ " + app_conf_path).Run()

	//forever_path := nodejs.ForeverPath(node_version)
	// run app with forever
	// todo: run each node app/binary like guppy, if it fails, respawn. don't use forever
	// also keep track of the running projects somehow and re-run them when fishtank loads.
	proj := data.Projects[project]
	proj.Lock()

	RunNewRev(project, new_rev, dashConfig)

	time.Sleep(3 * time.Second) // give the app a chance to boot up

	_ = command.E("nginx -s reload").Run()

	time.Sleep(1 * time.Second) // give nginx a chance to reload

	// kill the old app
	KillOldRev(project, old_rev)

	proj.Unlock()
	// cleanup old rev
	_ = command.E("rm -R " + old_rev_path)

	// go restricts map property assignments. workaround
	p := data.Projects[project]
	p.Current_revision = new_rev
	data.Projects[project] = p

	persistence.SaveData(data, DATA_JSON)

}
