// to handle system administration involving users
package main

import (
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
)

type PageAdminStruct struct {
	Id			string
	Username	string
	Email		string
	Usergroup	string
}

func AdminHandler(r *mux.Router) {
	r.HandleFunc("/admin", PageAdmin)
}

func PageAdmin(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		session, _ := store.Get(r, "cookie-name")
		username := session.Values["username"].(string)
		if AccessAdmin(GetUsergroup(GetUserId(username))) {
			data := ReadUserAccount(username) //NOTE: probably won't need all of the info
			tmpl := template.Must(template.ParseFiles("template/admin/index.html"))
			tmpl.Execute(w, data)
		} else {
			// assuming user came from "/user"
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

func (p PageAdminStruct) UserPermission(permission string, usergroup string) bool {
	switch permission {
	case "access_admin":
		return AccessAdmin(usergroup)
	default:
		return false
	}
}