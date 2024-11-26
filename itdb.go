package main

import (
	"fmt"
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

// since we cannot modify existing struct, we can embed a struct into another struct
// https://stackoverflow.com/a/29019923
type PageITDBAddPC struct {
	Region string
	PageITDBStruct
}

type PageITDBStruct struct {
	Id			string
	Username	string
	Email		string
	Usergroup	string
}

func ITDBHandler(r *mux.Router) {
	r.HandleFunc("/itdb", PageITDB)
	r.HandleFunc("/itdb/setting", PageITDBSetting)
	r.HandleFunc("/itdb/pc/{region}", PageITDBPC)
	r.HandleFunc("/itdb/pc/{region}/add", PageITDBPCAdd)
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

// "/itdb/pc/{region}"
func PageITDBPC(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			data := PageITDBStruct {
				"",
				username,
				"",
				"",
			}

			tmpl := template.Must(template.ParseFiles("template/itdb/pclist.html"))
			tmpl.Execute(w, data)
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

// "/itdb/pc/{region}/add"
func PageITDBPCAdd(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			region := mux.Vars(r)["region"]
			fmt.Println("region is ", region) //TODO: please remove when no longer needed

			userbasic := PageITDBStruct {
				"",
				username,
				"",
				usergroup,
			}
			data := PageITDBAddPC {
				Region: region,
				PageITDBStruct: userbasic,
			}

			tmpl := template.Must(template.ParseFiles("template/itdb/addpc.html"))
			tmpl.Execute(w, data)			
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

