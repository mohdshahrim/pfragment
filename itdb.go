package main

import (
	"fmt"
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

type PCList struct {
	Office	string
	PCs	[]PC
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
	r.HandleFunc("/itdb/pc/{office}/add/submit", ITDBPCAddSubmit)
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
		_, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			office := mux.Vars(r)["office"]
			data := PCList {
				Office: office,
				PCs: GetPC(office),
			}
			fmt.Println(data)
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

// function to get all PCs as per office
func GetPC(office string) []PC {
	db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
	if errOpen != nil {
		log.Fatal("error opening itdb.db ", errOpen)
	}
	defer db.Close()

    var pcstruct []PC

	query := ""

	switch(office) {
	case "sibu":
		query = "SELECT * FROM " + pcsibu 
	case "kapit":
		query = "SELECT * FROM " + pckapit
	}

    row, err := db.Query(query)
	
	if err == sql.ErrNoRows {
		// if it indeed has no rows, it means the table is still new
		//log.Fatal("func GetPC() no rows ", err)
		return pcstruct
	} else if err != nil {
		log.Fatal("func GetPC() return error :", err)
	}

    defer row.Close()
    for row.Next() {
        pc := PC{}
		ns := struct {
			Hostname		sql.NullString
			Ip				sql.NullString
			Cpumodel		sql.NullString
			Cpuno			sql.NullString
			Monitormodel	sql.NullString
			Monitorno		sql.NullString
			Printer			sql.NullString
			User			sql.NullString
			Department		sql.NullString
			Notes			sql.NullString
		} {

		}
        err := row.Scan(&pc.Id, &ns.Hostname, &ns.Ip, &ns.Cpumodel, &ns.Cpuno, &ns.Monitormodel, &ns.Monitorno, &ns.Printer, &ns.User, &ns.Department, &ns.Notes)
        if err != nil {
            log.Fatal(err)
        }

		// reassigns ns to pc
		pc.Hostname = ns.Hostname.String
		pc.Ip = ns.Ip.String
		pc.Cpumodel = ns.Cpumodel.String
		pc.Cpuno = ns.Cpuno.String
		pc.Monitormodel = ns.Monitormodel.String
		pc.Monitorno = ns.Monitorno.String
		pc.Printer = ns.Printer.String
		pc.User = ns.User.String
		pc.Department = ns.Department.String
		pc.Notes = ns.Notes.String

		//fmt.Println(pc)//TEST

        pcstruct = append(pcstruct, pc)
    }

    return pcstruct
}

// function to handle add new PC
func ITDBPCAddSubmit(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			r.ParseForm()

			office := r.FormValue("office")
			hostname := r.FormValue("hostname")
			ip := r.FormValue("ip")
			cpu_model := r.FormValue("cpu_model")
			cpu_no := r.FormValue("cpu_no")
			monitor_model := r.FormValue("monitor_model")
			monitor_no := r.FormValue("monitor_no")
			user := r.FormValue("user")
			department := r.FormValue("department")
			notes := r.FormValue("notes")
			// have to check because sometimes there are no printer to be set
			printer := "" //default
			if r.PostForm.Has("printer") {
				printer = r.FormValue("printer")
			}

			fmt.Println("printer is ", printer , " office is ", office)//TEST

			db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
			if errOpen != nil {
				log.Fatal(errOpen)
			}
			defer db.Close()

			// decides which table
			pctable := ""
			switch(office) {
			case "sibu":
				pctable = pcsibu
			case "kapit":
				pctable = pckapit
			}

			_, err := db.Exec(`INSERT INTO ` + pctable + ` (hostname, ip, cpu_model, cpu_no, monitor_model, monitor_no, printer, user, department, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, hostname, ip, cpu_model, cpu_no, monitor_model, monitor_no, printer, user, department, notes)

			if err != nil {
				log.Println(err)
			} else {
				data := PageITDBStruct {
					"",
					username,
					"",
					"",
				}
				tmpl := template.Must(template.ParseFiles("template/itdb/index.html"))
				tmpl.Execute(w, data)
			}
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else{
		http.Redirect(w, r, "/", 302)
	}
}

func PCStructTest() string {
	return ""
}
