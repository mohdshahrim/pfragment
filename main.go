package main

import (
	"fmt"
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
)

type PageIndexStruct struct {
	Message string
	Version string
}


func main() {
	// mux
	r := mux.NewRouter()

	// for assets files
	fs := http.FileServer(http.Dir("./asset/"))
	r.PathPrefix("/asset/").Handler(http.StripPrefix("/asset/", fs))

	// routes handled within main.go
	r.HandleFunc("/", PageIndex("")) // index page

	// routes handled in separate go files
	UserHandler(r) // user.go
	AboutHandler(r) // about.go
	AdminHandler(r) // admin.go

	// start the server
	fmt.Println("Starting server...")
	fmt.Println("Go to http://localhost:8000/")
	http.ListenAndServe(":8000", r)
}

// function to return index page
func PageIndex(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("template/index.html"))
		data := PageIndexStruct{
			message,
			"version 1.0.0 (07/11/2024)",
		}
		tmpl.Execute(w, data)
	}
}

func PageIndexRedirect(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("template/index.html"))
	data := PageIndexStruct{
		"wrong username or password",
		"version 1.0.0 (07/11/2024)",
	}
	tmpl.Execute(w, data)
}