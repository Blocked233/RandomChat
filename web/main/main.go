package main

import (
	"fmt"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "index.html")

}

func main() {

	http.HandleFunc("/", index)
	http.HandleFunc("/ws", socketHandler)

	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		fmt.Println("Start Error!", err)
	}

}
