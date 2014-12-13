package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func checkInstalledVersion() {
	// check if there's already a configfile on this system
	_, err := os.Open(FISHTANK_CONFIG)
	if err != nil {
		// no config, so install
		Install()
	}

	// note: this will also create the config file if it didn't already exist
	config = GetConfig()

	if config.Version != VERSION {
		Install()
		config.Version = VERSION
		SetConfig(config)
	}
}

func Install() {
	// todo: make run on boot. add line fishtank, to /etc/rc.local. make sure you only add it once.
	fmt.Println("Installing...")
	_ = exec.Command("mkdir", "-p", "/srv/coral/").Run()
	_ = exec.Command("mkdir", "-p", "/srv/www/").Run()
	_ = exec.Command("mkdir", "-p", "/srv/logs/").Run()

	_ = exec.Command("mkdir", "-p", "/usr/local/n").Run()

	_ = exec.Command("mkdir", "-p", "/etc/fishtank/").Run()

	_ = exec.Command("touch", "/etc/fishtank/data.json").Run()
	//_ = exec.Command("chmod", "ug+x", "-R", "/usr/local/n").Run()

	// install nginx
	_ = exec.Command("apt-get", "install", "nginx", "-y").Run()

	_ = exec.Command("nginx").Run()

	CreateStartupScript()
	_ = exec.Command("chmod", "ugo+x", "/etc/init.d/fishtank").Run()
	_ = exec.Command("update-rc.d", "fishtank", "defaults").Run()

	// copy fishtank binary to /usr/bin
	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		log.Fatal("installing fortune is in your future")
	}
	if path != INSTALL_PATH+"/fishtank" {
		_ = exec.Command("cp", path, INSTALL_PATH+"/fishtank").Run()
	}

	// if the daemon is running, restart it
	_, err = http.Get("http://127.0.0.1:" + config.Port)
	if err == nil {
		fmt.Println("The fishtank service needs to be restarted. run:")
		fmt.Println("   > nohup /etc/init.d/fishtank restart 2>&1 </dev/null &")
	} else {
		fmt.Println("The fishtank service needs to be started. run:")
		fmt.Println("   > nohup /etc/init.d/fishtank start 2>&1 </dev/null &")
	}
	//_ = command.E("> nohup /etc/init.d/fishtank restart 2>&1 </dev/null &").Run()
	//  > /etc/init.d/fishtank 2>&1 </dev/null &
}

func Uninstall() {
	_ = exec.Command("update-rc.d", "-f", "fishtank", "remove").Run() //  remove startup script
	_ = exec.Command("rm", "-R", "/etc/init.d/fishtank").Run()

	_ = exec.Command("rm", "-R", "/srv/coral/").Run()
	_ = exec.Command("rm", "-R", "/srv/www/").Run()
	_ = exec.Command("rm", "-R", "/srv/logs/").Run()

	// todo: for each project remove the conf
	for j := range data.Projects {
		_ = exec.Command("rm", "-R", "/etc/nginx/sites-enabled/"+j+".conf").Run()
	}

	_ = exec.Command("rm", "-R", "/etc/fishtank/").Run()
	_ = exec.Command("rm", "-R", INSTALL_PATH+"fishtank").Run()
}
