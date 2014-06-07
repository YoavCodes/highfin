package nodejs

import (
	"fmt"
	"os"
	"os/exec"
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

func InstallNode(version string) {
	fmt.Println("Installing node.js v" + version)
	version_folder := BinFolder(version)

	if _, err := os.Stat(BinPath(version)); os.IsNotExist(err) {
		_ = exec.Command(`mkdir`, `-p`, version_folder).Run()

		fmt.Println(`fetching ` + GetUrl(version))
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
			fmt.Println("failed to fetch node.js v" + version)
		}

	} else {
		fmt.Println("node.js " + version + " already installed")
	}
}
