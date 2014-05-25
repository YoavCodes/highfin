package util

import "fmt"
import "os"
import "os/exec"
import "net/http"
import "path/filepath"
import "io/ioutil"

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
	_ = filepath.Walk(KEY_PATH, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == true || path[:len(path)-4] == ".pub" {
			return err
		}

		// todo: use sed / grep to ensure only one instance of this exists in /etc/ssh/ssh_config

		file, _ := os.OpenFile("/etc/ssh/ssh_config", os.O_APPEND, 777)

		defer file.Close()

		_, _ = file.WriteString("Identityfile /vagrant/keys/guppy")

		// fmt.Println("add key", path)
		// _, file := filepath.Split(path)

		// new_path := "/home/vagrant/.ssh/" + file

		// cmd := exec.Command("cp", "-rfp", path, new_path)
		// _, _ = cmd.CombinedOutput()

		return err
	})
}

func GenerateKey(email string) {
	// make sure /vagrant/keys folder exists
	os.MkdirAll(KEY_PATH, 666)
	blank_string := ""
	out, err := exec.Command("ssh-keygen", "-t", "ecdsa", "-N", blank_string, "-f", KEY_PATH+"/guppy", "-C", email).CombinedOutput()
	if err != nil {
		Log("Generating ssh keys failed: " + err.Error())
	}

	if out != nil {
		Log(string(out))
	}
}

func ValidateKey(account string, project string, email string) bool {
	// output message with guppy.pub so user can copy and paste
	// todo(yoav) replace with actual shark api for validating keys
	Log("Verifying dev has access to project..")

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
		Log("==== Bleep blop bloop")
		// Log("Our records show that you do not have access to this project")
		// Log("Please register your dev key with " + account + "'s " + project + " project")
		// Log("You can do this by signing into http://highf.in as admin")
		// Log("or contacting the admin of your account.")
		Log("Here's your public key, you'll need to add it to the project online to gain access.")
		Log("============\n" + string(public_key) + "\n============")
		Log("Press enter to check again")

		a := ""
		fmt.Scanf("%s", &a) // scan is not waiting

		return ValidateKey(account, project, email)

	} else {
		Log("Validation success")
	}
	// note: it will get stuck in an infinite loop of asking you to upload your keys
	// user will have to ctrl+c and only get here if it's valid
	return true
}

func GetKey() {

}
