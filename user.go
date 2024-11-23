package main

import (
	"fmt"
	"log"
	"net/http"
	"html/template"
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

type PageUserStruct struct {
	Username string
	LoggedOn string
}

type PageAccountStruct struct {
	Id			string
	Username	string
	Email		string
	Usergroup	string
}

type PagePasswordStruct struct {
	Username	string
	Message		string
}

func UserHandler(r *mux.Router) {
	r.HandleFunc("/user", PageUser)
	//r.HandleFunc("/user/login", UserLogin).Methods("POST")
	r.HandleFunc("/user/login", UserLogin)
	r.HandleFunc("/user/account", PageAccount)
	r.HandleFunc("/user/password", PageUpdatePassword)
	r.HandleFunc("/user/password/update", UserUpdatePassword)
	r.HandleFunc("/user/logout", UserLogout)
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		PageIndexRedirect(w,r)
	}

	if UsernameExist(r.FormValue("username")) {
		if PasswordIsValid(r.FormValue("username"), r.FormValue("password")) {
			// obtain id and username, to be put in session
			username := r.FormValue("username")
			id := GetUserId(username)

			login(w,r,id,username)

			PageRedirect(w,r)
		} else {
			// redirect user back to login
			fmt.Println("password invalid")
			PageIndexRedirect(w,r)
		}
	} else {
		// redirect user back to login
		fmt.Println("username invalid")
		PageIndexRedirect(w,r)
	}
}

func UserLogout(w http.ResponseWriter, r *http.Request) {
	logout(w,r)
	http.Redirect(w, r, "/", 302)
}

func PageRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/user", 302)
}

func PageUser(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		tmpl := template.Must(template.ParseFiles("template/user/index.html"))

		session, _ := store.Get(r, "cookie-name")
		username := session.Values["username"].(string)
		loggedon := session.Values["loggedon"].(string)

		data := PageUserStruct{
			username,
			loggedon,
		}
		tmpl.Execute(w, data)
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

func PageAccount(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		session, _ := store.Get(r, "cookie-name")
		data := ReadUserAccount(session.Values["username"].(string))
		tmpl := template.Must(template.ParseFiles("template/user/account.html"))
		tmpl.Execute(w, data)
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

func PageUpdatePassword(w http.ResponseWriter, r *http.Request) {	
	if IsAuthenticated(w,r) {
		session, _ := store.Get(r, "cookie-name")
		username := session.Values["username"].(string)

		data := PagePasswordStruct{username, ""}

		if UpdateOwnPassword(GetUsergroup(GetUserId(username))) {	
			tmpl := template.Must(template.ParseFiles("template/user/password.html"))
			tmpl.Execute(w, data)
		} else {
			//NOTE: assuming user previously came from "/user/account"
			http.Redirect(w, r, "/user/account", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

func (p PageAccountStruct) UserPermission(permission string, usergroup string) bool {
	return UsergroupPermission(permission, usergroup)
}

// this particular functions, unlike most UserPermission function, take username as the second parameter
// because its struct do not have usergroup defined. See PageUserStruct.
func (p PageUserStruct) UserPermission(permission string, username string) bool {
	usergroup := GetUsergroup(GetUserId(username))
	return UsergroupPermission(permission, usergroup)
}

func (p PagePasswordStruct) UserPermission(permission string, username string) bool {
	usergroup := GetUsergroup(GetUserId(username))
	return UsergroupPermission(permission, usergroup)
}

// performs password update procedure
func UserUpdatePassword(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username := r.FormValue("username")

		if UsernameExist(username) {
			oldpassword := r.FormValue("oldpassword")

			if PasswordIsValid(username, oldpassword) {
				newpassword := r.FormValue("newpassword")
				confirmpassword := r.FormValue("confirmpassword")

				if newpassword==confirmpassword {
					id := GetUserId(username)

					// begin procedure of updating password
					db, errOpen := sql.Open("sqlite3", "./database/core.db")
					if errOpen != nil {
						log.Fatal(errOpen)
					}
					defer db.Close()

					query := `UPDATE user SET password = ? WHERE id = ?`
					_, err := db.Exec(query, newpassword, id)
					if err != nil {
						log.Fatal(err)
					}
					// success
					data := PagePasswordStruct{username, "Password update success"}

					tmpl := template.Must(template.ParseFiles("template/user/password.html"))
					tmpl.Execute(w, data)
				} else {
					data := PagePasswordStruct{username, "Error. Invalid password confirmation."}
		
					tmpl := template.Must(template.ParseFiles("template/user/password.html"))
					tmpl.Execute(w, data)
				}
			} else {
				data := PagePasswordStruct{username, "Error. Old password is incorrect."}

				tmpl := template.Must(template.ParseFiles("template/user/password.html"))
				tmpl.Execute(w, data)
			}
		} else {
			data := PagePasswordStruct{"", "Error. Username invalid. Please consider relogin."}

			tmpl := template.Must(template.ParseFiles("template/user/password.html"))
			tmpl.Execute(w, data)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}

}

// function to verify whether the username exist or not
func UsernameExist(username string) bool {
	// Connect to SQLite database
	db, errOpen := sql.Open("sqlite3", "./database/core.db")
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	query := `SELECT COUNT(*) FROM user WHERE username = ?`
	var count int
	errQuery := db.QueryRow(query, username).Scan(&count)
	if errQuery != nil {
		log.Fatal(errQuery)
	}

	// If count > 0, the username exists
	return count > 0
}

// function to validate password
func PasswordIsValid(username string, password string) bool {
	// Connect to SQLite database
	db, errOpen := sql.Open("sqlite3", "./database/core.db")
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	query := `SELECT password FROM user WHERE username = ?`
	password_hash := ""
	err := db.QueryRow(query, username).Scan(&password_hash)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		log.Fatal(err)
		return false
	}

	// check if given password same with in the table
	if password == password_hash {
		return true
	} else {
		return false
	}
}

// function to get id based on username
func GetUserId(username string) string {
	strId := ""

	db, errOpen := sql.Open("sqlite3", "./database/core.db")
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	query := `SELECT id FROM user WHERE username = ?`
	err := db.QueryRow(query, username).Scan(&strId)

	if err == sql.ErrNoRows {
		log.Fatal(err)
		return ""
	} else if err != nil {
		log.Fatal(err)
		return ""
	}

	return strId
}

// function to get usergroup based on id
func GetUsergroup(id string) string {
	strUsergroup := ""

	db, errOpen := sql.Open("sqlite3", "./database/core.db")
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	query := `SELECT usergroup FROM user WHERE id = ?`
	err := db.QueryRow(query, id).Scan(&strUsergroup)

	if err == sql.ErrNoRows {
		log.Fatal(err)
		return ""
	} else if err != nil {
		log.Fatal(err)
		return ""
	}

	return strUsergroup
}

// function to get basic user info from db based on username
func ReadUserAccount(username string) PageAccountStruct {
	data := PageAccountStruct{
		"",
		username,
		"",
		"",
	}

	// Connect to SQLite database
	db, errOpen := sql.Open("sqlite3", "./database/core.db")
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	//defer db.Close()

	query := `SELECT id, email, usergroup FROM user WHERE username = ?`
	err := db.QueryRow(query, data.Username).Scan(&data.Id, &data.Email, &data.Usergroup)

	if err == sql.ErrNoRows {
		fmt.Println("serious error")
		//return false
	} else if err != nil {
		log.Fatal(err)
		//return false
	}

	return data
}