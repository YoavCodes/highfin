package nodejs

import (
	"code.highf.in/chalkhq/shared/command"
	"code.highf.in/chalkhq/shared/log"
	"os"
	"os/exec"
	"strings"
)

func GetUrl(version string) string {
	return `http://nodejs.org/dist/v` + version + `/node-v` + version + `-linux-x64.tar.gz`
}

func BinFolder(version string) string {
	return `/usr/local/n/` + version
}

func BinPath(version string) string {
	return `/usr/local/n/` + version + `/bin/node`
}

func NpmPath(version string) string {
	return `/usr/local/n/` + version + `/bin/npm`
}

func LessPath(version string) string {
	return `/usr/local/n/` + version + `/bin/lessc`
}

func InstallNode(version string) {

	version_folder := BinFolder(version)

	if _, err := os.Stat(BinPath(version)); os.IsNotExist(err) {
		log.Log("Installing node.js v" + version)
		_ = exec.Command(`mkdir`, `-p`, version_folder).Run()

		log.Log(`fetching ` + GetUrl(version))
		cmd := exec.Command(`curl`, `-L`, GetUrl(version))

		cmd2 := exec.Command(`tar`, `-zx`, `--strip`, `1`, `-C`, version_folder)

		cmd2.Stdin, _ = cmd.StdoutPipe()

		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr

		_ = cmd2.Start()
		err := cmd.Start()
		_ = cmd2.Wait()
		_ = cmd.Wait()

		if err != nil {
			log.Log("failed to fetch node.js v" + version)
			return
		}

		// install lessc (command line less compiler) for the current version
		command.E(NpmPath(version) + " install less -g")

	} else {
		log.Log("using nodejs v" + version + "")
	}
}

// todo: not currently used anywhere
func Npm(args []string) {
	// todo: setup "current" version with /n/current symlinked to the current version
	version_folder := NpmPath("0.10.28")
	//args = args[1:]
	args[0] = version_folder
	command_string := strings.Join(args, " ")
	cmd := command.E(command_string)
	cmd.Run()

}
