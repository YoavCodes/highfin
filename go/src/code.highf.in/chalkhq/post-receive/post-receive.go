package main

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
)

var (
	oldrev  string
	newrev  string
	ref     string
	branch  string
	account string
	project string
)

func main() {
	// loop through refs pushed. ie: if multiple branches were pushed

	for {
		if n, err := fmt.Scanf("%s %s %s", &oldrev, &newrev, &ref); n != 3 || err != nil {
			break
		}
		http.Get("http://10.10.10.5/" + oldrev + "/" + newrev + "/" + ref)

		ref_arr := regexp.MustCompile("/").Split(ref, 4)
		branch = ref_arr[2]

		if branch == "dev-next" || branch == "dev-current" {
			// todo: support dev-current as well
			// /coral/account_name/project_name/code.git
			path, _ := filepath.Abs("")

			path_arr := regexp.MustCompile("/").Split(path, 5)
			account = path_arr[2]
			project = path_arr[3]

			// todo: fix server for vagrant/prod should use http://octopus.highf.in and /etc/hosts should be configured in vagrant
			// todo: this code is shared deploy code with guppy/util/git.go, refactor
			res, err := http.PostForm("http://10.10.10.50/project/deploy", url.Values{"account": {account}, "project": {project}, "branch": {branch}})
			if err != nil {
				return
			}
			defer res.Body.Close()

		}
	}

}
