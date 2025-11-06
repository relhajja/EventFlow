package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, Go Web ! I'am Riad \n")
}

func indexhandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("index.html"))
	if err := t.Execute(w, "Go Templates!"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {

	http.HandleFunc("/", handler)
	http.HandleFunc("/index", indexhandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
