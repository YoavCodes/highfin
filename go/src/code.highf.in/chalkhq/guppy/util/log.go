package util

import "fmt"

// this will be global within the util package
func Log(msg ...string) {
	title := "GUPPY"

	for i := 0; i < len(msg); i++ {
		fmt.Println("\033[0;97;41m ===" + title + "=== \033[0m\033[0;97;46m " + msg[i] + " \033[0m")
	}

}

// logs an error E(err) or util.E(err)
func LogE(err error) {
	if err != nil {
		Log("error: " + err.Error())
	}
}
