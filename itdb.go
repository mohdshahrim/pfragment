package main

import (
	_ "fmt"
	"log"
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
	"database/sql"
)

const (
	pcsibu = "pcsibu1"
	pckapit = "pckapit1"
)

// since we cannot modify existing struct, we can embed a struct into another struct
// https://stackoverflow.com/a/29019923
type PageITDBAddPC struct {
	Office		string
	PageITDBStruct
	PC
	Printers	[]Printer
}

type PageITDBStruct struct {
	Id			string
	Username	string
	Email		string
	Usergroup	string
}

type PC struct {
	Id				int
	Hostname		string
	Ip				string
	Cpumodel		string
	Cpuno			string
	Monitormodel	string
	Monitorno		string
	Printer			string
	User			string
	Department		string
	Notes			string
}

type Printer struct {
	Rowid			int
	Printermodel	string
	Printerno		string
	Printertype		string
	Notes			sql.NullString
	Host			sql.NullInt64
	Nickname		string
}

func ITDBHandler(r *mux.Router) {
	r.HandleFunc("/itdb", PageITDB)
	r.HandleFunc("/itdb/setting", PageITDBSetting)
	r.HandleFunc("/itdb/pc/{office}", PageITDBPC)
	r.HandleFunc("/itdb/pc/{office}/add", PageITDBPCAdd)
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

// "/itdb/pc/{office}"
func PageITDBPC(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			office := mux.Vars(r)["office"]
			userbasic := PageITDBStruct {
				"",
				username,
				"",
				usergroup,
			}
			data := PageITDBAddPC {
				Office: office,
				PageITDBStruct: userbasic,
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

// "/itdb/pc/{office}/add"
func PageITDBPCAdd(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			office := mux.Vars(r)["office"]

			userbasic := PageITDBStruct {
				"",
				username,
				"",
				usergroup,
			}
			data := PageITDBAddPC {
				Office: office,
				PageITDBStruct: userbasic,
				Printers: GetPrinterNoHost(office),
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

// function to return all printers that has no host
func GetPrinterNoHost(office string) []Printer {
	db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
	if errOpen != nil {
		log.Fatal("error opening itdb.db ", errOpen)
	}
	defer db.Close()

    var printerstruct []Printer

	query := ""

	switch(office) {
	case "sibu":
		query = "SELECT rowid, * FROM printersibu1 WHERE host IS null OR host=''"
	case "kapit":
		query = "SELECT rowid, * FROM printerkapit1 WHERE host IS null OR host=''"
	}

    row, err := db.Query(query)
	
	if err == sql.ErrNoRows {
		log.Fatal("func GetPrinterNoHost() no rows ", err)
	} else if err != nil {
		log.Fatal("func GetPrinterNoHost() return err nil ", err)
	}

    defer row.Close()
    for row.Next() {
        printer := Printer{}
        err := row.Scan(&printer.Rowid, &printer.Printermodel, &printer.Printerno, &printer.Printertype, &printer.Notes, &printer.Host, &printer.Nickname)
        if err != nil {
            log.Fatal(err)
        }
        printerstruct = append(printerstruct, printer)
    }

    return printerstruct
}
