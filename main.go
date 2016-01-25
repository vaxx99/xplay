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
//	log.Print(req,"\n\n")
	fmt.Fprintf(w, "%s", b.Hproc())
}


func main() {
    http.HandleFunc("/", Fsto)
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
	log.Fatal("ListenAndServe: ", err)
    }
}
