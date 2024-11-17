package main

import (
	"net/http"
	"github.com/gorilla/mux"
)

func AboutHandler(r *mux.Router) {
	r.HandleFunc("/about", PageAbout)
}

func PageAbout(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "template/about.html")
}