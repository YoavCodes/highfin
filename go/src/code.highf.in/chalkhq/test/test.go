package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("This is app 1 starting up...")
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("This is app 1 saying hello")
		w.Write([]byte("app 1 says hello babycakes!!"))
	})

	http.ListenAndServe(":8080", nil)
}
