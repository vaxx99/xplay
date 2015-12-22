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
	    mux := http.NewServeMux()
	    mux.HandleFunc("/", Fsto)
	    log.Fatal(http.ListenAndServe(":8000", mux))
	}
