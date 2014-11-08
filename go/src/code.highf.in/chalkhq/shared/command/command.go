package command

import (
	"os"
	"os/exec"
	"strings"
)

func E(command string) *exec.Cmd {

	args := strings.Split(command, ` `)
	command_string := strings.Trim(args[0], ` `)
	for i := 0; i < len(args); i++ {
		args[i] = strings.Trim(args[i], ` `)
	}

	_cmd := exec.Command(command_string)
	_cmd.Args = args
	_cmd.Stdout = os.Stdout
	_cmd.Stderr = os.Stderr

	return _cmd

}
