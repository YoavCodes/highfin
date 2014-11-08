package util

import "fmt"
import "os"
import "os/exec"
import "net/http"
import "path/filepath"
import "io/ioutil"
import "code.highf.in/chalkhq/shared/log"

const (
	KEY_PATH string = "/vagrant/keys"
)

func ApplyKeys(email string) {
	// if there is no guppy key, then generate one
	if _, err := os.Stat(KEY_PATH + "/guppy"); os.IsNotExist(err) {
		GenerateKey(email)
	}

	// copy all keys in KEY_PATH into /home/vagrant/.ssh/ so git can find them
	// this will allow users to have keys for their own remote repos/accounts
	// KEY_PATH folder provides easy access to host os, but should be git ignored

	// todo: use sed / grep to ensure only one instance of this exists in /etc/ssh/ssh_config
	//file, err := os.OpenFile("/etc/ssh/ssh_config", os.O_APPEND, 0777)
	file, err := os.Create("/etc/ssh/ssh_config")
	defer file.Close()

	if err != nil {
		fmt.Println("error opening file")
		fmt.Println(err.Error())
	}

	_ = filepath.Walk(KEY_PATH, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == true || path[len(path)-4:] == ".pub" {
			fmt.Println(3)
			return err
		}

		_, err = file.WriteString("Identityfile " + path + "\n")

		if err != nil {
			fmt.Println("error writing to file")
			fmt.Println(err.Error())

		}

		return err
	})
}

func GenerateKey(email string) {
	// make sure /vagrant/keys folder exists
	os.MkdirAll(KEY_PATH, 666)
	blank_string := ""
	out, err := exec.Command("ssh-keygen", "-t", "ecdsa", "-N", blank_string, "-f", KEY_PATH+"/guppy", "-C", email).CombinedOutput()
	if err != nil {
		log.Log("Generating ssh keys failed: " + err.Error())
	}

	if out != nil {
		log.Log(string(out))
	}
}

func ValidateKey(account string, project string, email string) bool {
	// output message with guppy.pub so user can copy and paste
	// todo(yoav) replace with actual shark api for validating keys
	log.Log("Verifying dev has access to project..")

	public_key, err := ioutil.ReadFile(KEY_PATH + "/guppy.pub")

	if err != nil {
		// remove, regenerate, and revalidate keys
		_ = exec.Command("rm", "-R", "/vagrant/keys/guppy").Run()
		_ = exec.Command("rm", "-R", "/vagrant/keys/guppy.pub").Run()
		ApplyKeys(email)
		return ValidateKey(account, project, email)
	}

	// todo(Yoav) rebuild salmon, and then shark api, in go
	resp, _ := http.Get("http://google.com/")

	if resp == nil {
		log.Log("==== Bleep blop bloop")
		// log.Log("Our records show that you do not have access to this project")
		// log.Log("Please register your dev key with " + account + "'s " + project + " project")
		// log.Log("You can do this by signing into http://highf.in as admin")
		// log.Log("or contacting the admin of your account.")
		log.Log("Here's your public key, you'll need to add it to the project online to gain access.")
		log.Log("============\n" + string(public_key) + "\n============")
		log.Log("Press enter to check again")

		a := ""
		fmt.Scanf("%s", &a) // scan is not waiting

		return ValidateKey(account, project, email)

	} else {
		log.Log("Validation success")
	}
	// note: it will get stuck in an infinite loop of asking you to upload your keys
	// user will have to ctrl+c and only get here if it's valid
	return true
}

func GetKey() {

}
