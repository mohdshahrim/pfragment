package main

import (
	_ "log"
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
	_ "database/sql"
)

const (
	pcsibu = "pcsibu1"
	pckapit = "pckapit1"
)

type PageITDBStruct struct {
	Id			string
	Username	string
	Email		string
	Usergroup	string
}

func ITDBHandler(r *mux.Router) {
	r.HandleFunc("/itdb", PageITDB)
	r.HandleFunc("/itdb/setting", PageITDBSetting)
}

func (p PageITDBStruct) UserPermission(permission string, username string) bool {
	usergroup := GetUsergroup(GetUserId(username))
	return UsergroupPermission(permission, usergroup)
}

func PageITDB(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			data := PageITDBStruct {
				"",
				username,
				"",
				"",
			}
			tmpl := template.Must(template.ParseFiles("template/itdb/index.html"))
			tmpl.Execute(w, data)
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

func PageITDBSetting(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			data := PageITDBStruct {
				"",
				username,
				"",
				"",
			}
			tmpl := template.Must(template.ParseFiles("template/itdb/setting.html"))
			tmpl.Execute(w, data)
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

