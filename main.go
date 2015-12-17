package main

import (
	"fmt"
	"fsto/fstx"
	"log"
	"net/http"
)

// Fsto server
func Fsto(w http.ResponseWriter, req *http.Request) {
	a := fstx.B{}
	b := a.Mproc()
	fmt.Fprintf(w, "%s", b.Hproc())
}

func main() {
	fs := http.FileServer(http.Dir("/"))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("./src"))))
	http.HandleFunc("/", Fsto)
	err := http.ListenAndServe(":8000", fs)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
