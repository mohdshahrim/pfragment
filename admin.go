// to handle system administration involving users
package main

import (
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
	Users		[]UserStruct
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
	r.HandleFunc("/admin/usermanagement/newuser", PageAdminNewUser)
	r.HandleFunc("/admin/usermanagement/newuser/submit", AdminNewUser)
	r.HandleFunc("/admin/usermanagement/deleteuser/{id}", AdminDeleteUser)
}

func PageAdmin(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		session, _ := store.Get(r, "cookie-name")
		username := session.Values["username"].(string)
		if AccessAdmin(GetUsergroup(GetUserId(username))) {
			data := Admin(username)
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
		usergroup := GetUsergroup(GetUserId(username))
		if AccessAdmin(usergroup) {
			//data := AllUser()
			data := PageAdminStruct{
				"",
				username,
				"",
				usergroup,
				AllUser(),
			}
			tmpl := template.Must(template.ParseFiles("template/admin/usermanagement.html"))
			tmpl.Execute(w, data)
		} else {
			// assuming user came from "/user"
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

// "/admin/usermanagement/newuser"
func PageAdminNewUser(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		session, _ := store.Get(r, "cookie-name")
		username := session.Values["username"].(string)
		usergroup := GetUsergroup(GetUserId(username))
		if AccessAdmin(usergroup) {
			data := Admin(username)
			tmpl := template.Must(template.ParseFiles("template/admin/newuser.html"))
			tmpl.Execute(w, data)
		} else {
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
		//AccessAdmin(GetUsergroup(GetUserId(username)))
	default:
		return false
	}
}

func Admin(username string) PageAdminStruct {


	data := PageAdminStruct{
		"",
		username,
		"",
		"",
		[]UserStruct{}, // empty reserved for UserStruct
	}

	// Connect to SQLite database
	db, errOpen := sql.Open("sqlite3", "./database/core.db")
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	query := `SELECT id, email, usergroup FROM user WHERE username = ?`
	err := db.QueryRow(query, data.Username).Scan(&data.Id, &data.Email, &data.Usergroup)

	if err == sql.ErrNoRows {
		//fmt.Println("serious error")
		//return false
	} else if err != nil {
		log.Fatal(err)
		//return false
	}

	return data
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

// handle the form for new user submission
func AdminNewUser(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		session, _ := store.Get(r, "cookie-name")
		username := session.Values["username"].(string)
		usergroup := GetUsergroup(GetUserId(username))
		if AccessAdmin(usergroup) {
			username := r.FormValue("username")
			email := r.FormValue("email")
			usergroup := r.FormValue("usergroup")
			password := r.FormValue("password")

			db, errOpen := sql.Open("sqlite3", "./database/core.db")
			if errOpen != nil {
				log.Fatal(errOpen)
			}
			defer db.Close()

			_, err := db.Exec(`INSERT INTO user (username, email, password, usergroup) VALUES (?, ?, ?, ?)`, username, email, password, usergroup)

			if err != nil {
				log.Println(err)
			} else {
				// show success page
				data := Admin(username)
				tmpl := template.Must(template.ParseFiles("template/admin/newuserok.html"))
				tmpl.Execute(w, data)
			}
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

// handle user deletion
func AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		session, _ := store.Get(r, "cookie-name")
		username := session.Values["username"].(string)
		usergroup := GetUsergroup(GetUserId(username))
		if AccessAdmin(usergroup) {
			vars := mux.Vars(r)
			id := vars["id"]

			db, errOpen := sql.Open("sqlite3", "./database/core.db")
			if errOpen != nil {
				log.Fatal(errOpen)
			}
			defer db.Close()

			_, err := db.Exec(`DELETE FROM user WHERE id = ?`, id) // check err

			if err != nil {
				log.Println(err)
			} else {
				// finish
				http.Redirect(w, r, "/admin/usermanagement", 302)
			}
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}