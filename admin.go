// to handle system administration involving users
package main

import (
	"fmt"
	"log"
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
	"database/sql"
)

type PageAdminStruct struct {
	Id			string
	Username	string
	Email		string
	Usergroup	string
}

type UserStruct struct {
	Id			string
	Username	string
	Email		string
	Password	string
	Usergroup	string
}

func AdminHandler(r *mux.Router) {
	r.HandleFunc("/admin", PageAdmin)
	r.HandleFunc("/admin/usermanagement", PageAdminUserManagement)
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

// "/admin/usermanagement"
func PageAdminUserManagement(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		session, _ := store.Get(r, "cookie-name")
		username := session.Values["username"].(string)
		if AccessAdmin(GetUsergroup(GetUserId(username))) {
			data := AllUser()
			tmpl := template.Must(template.ParseFiles("template/admin/usermanagement.html"))
			tmpl.Execute(w, data)
		} else {
			// assuming user came from "/user"
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		// assuming user came from "/user"
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

func AllUser() []UserStruct {
	db, errOpen := sql.Open("sqlite3", "./database/core.db")
	if errOpen != nil {
		log.Fatal("error opening core.db ", errOpen)
	}
	defer db.Close()

    var userstruct []UserStruct
    row, err := db.Query("SELECT * FROM user")
	
	if err == sql.ErrNoRows {
		log.Fatal("func AllUser() no rows ", err)
	} else if err != nil {
		log.Fatal("func AllUser() return err nil ", err)
	}

    defer row.Close()
    for row.Next() {
        user := UserStruct{}
        err := row.Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.Usergroup)
        if err != nil {
            log.Fatal(err)
        }
        userstruct = append(userstruct, user)
    }

    return userstruct
}