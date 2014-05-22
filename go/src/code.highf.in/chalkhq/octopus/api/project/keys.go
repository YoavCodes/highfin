package project

import "os"
import "fmt"

func AddKey(account string, project string, key string) {
	project_path := "/coral/" + account + "/" + project + "/"

	fmt.Println(key)

	authorized_keys, err := os.Create(project_path + ".ssh/authorized_keys")
	defer authorized_keys.Close()

	if err != nil {
		fmt.Println("error: " + err.Error())
		return
	}

	authorized_keys.WriteString(key + "\n\n")
}
