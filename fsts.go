package main

import (
	"fmt"
	"fsto/fstx"
	"log"
	"net/http"
)

func Home(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s",
		fstx.HTML(fstx.Pars()))
}

func main() {
	http.HandleFunc("/", Home)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
