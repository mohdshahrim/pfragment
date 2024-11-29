package main

import (
	"fmt"
	"log"
	"strconv"
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
	"database/sql"
)

const (
	pcsibu = "pcsibu1"
	pckapit = "pckapit1"
	printersibu = "printersibu1"
	printerkapit = "printerkapit1"
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

type PrinterList struct {
	Office string
	Printers []Printer
}

type PageITDBStruct struct {
	Id			string
	Username	string
	Email		string
	Usergroup	string
}

type PC struct {
	Office			string
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
	Office			string
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
	r.HandleFunc("/itdb/printer/{office}", PageITDBPrinter)
	r.HandleFunc("/itdb/printer/{office}/add", PageITDBPrinterAdd)
	r.HandleFunc("/itdb/printer/{office}/add/submit", ITDBPrinterAddSubmit)
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

// /itdb/printer/{office}
func PageITDBPrinter(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		_, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			office := mux.Vars(r)["office"]
			data := PrinterList {
				Office: office,
				Printers: GetPrinter(office),
			}

			tmpl := template.Must(template.ParseFiles("template/itdb/printerlist.html"))
			tmpl.Execute(w, data)
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else{
		http.Redirect(w, r, "/", 302)
	}
}

// "/itdb/printer/{office}/add"
func PageITDBPrinterAdd(w http.ResponseWriter, r *http.Request) {
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

			data := struct {
				Office string
				PageITDBStruct PageITDBStruct
			}{
				office,
				userbasic,
			}

			tmpl := template.Must(template.ParseFiles("template/itdb/addprinter.html"))
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
        err := row.Scan(&pc.Id, &pc.Hostname, &pc.Ip, &pc.Cpumodel, &pc.Cpuno, &pc.Monitormodel, &pc.Monitorno, &pc.Printer, &pc.User, &pc.Department, &pc.Notes)
        if err != nil {
            log.Fatal(err)
        }

		pc.Office = office //assigns at each row, because when inside range, global ".Office" is not recognized

        pcstruct = append(pcstruct, pc)
    }

    return pcstruct
}

// function to get all printers as per office
func GetPrinter(office string) []Printer {
	db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
	if errOpen != nil {
		log.Fatal("error opening itdb.db ", errOpen)
	}
	defer db.Close()

    var printerstruct []Printer

	query := ""

	switch(office) {
	case "sibu":
		query = "SELECT rowid, * FROM " + printersibu 
	case "kapit":
		query = "SELECT rowid, * FROM " + printerkapit
	}

    row, err := db.Query(query)
	
	if err == sql.ErrNoRows {
		// if it indeed has no rows, it means the table is still new
		//log.Fatal("func GetPrinter() no rows ", err)
		return printerstruct
	} else if err != nil {
		log.Fatal("func GetPrinter() return error :", err)
	}

    defer row.Close()
    for row.Next() {
        printer := Printer{}
        err := row.Scan(&printer.Rowid, &printer.Printermodel, &printer.Printerno, &printer.Printertype, &printer.Notes, &printer.Host, &printer.Nickname)
        if err != nil {
            log.Fatal(err)
        }

		printer.Office = office //assigns at each row, because when inside range, global ".Office" is not recognized

        printerstruct = append(printerstruct, printer)
    }

    return printerstruct
}

// function to offset index at range so that it begins at 1
func (p PC) IndexOffset(index int) string {
	index = index + 1
	return strconv.Itoa(index)
}

func (p Printer) IndexOffset(index int) string {
	index = index + 1
	return strconv.Itoa(index)
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

// function to handle add new printer
func ITDBPrinterAddSubmit(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		_, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			r.ParseForm()

			office := r.FormValue("office")
			printermodel := r.FormValue("printermodel")
			printerno := r.FormValue("printerno")
			printertype := r.FormValue("printertype")
			notes := r.FormValue("notes")
			nickname := r.FormValue("nickname")

			db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
			if errOpen != nil {
				log.Fatal(errOpen)
			}
			defer db.Close()

			// decides which table
			printertable := ""
			switch(office) {
			case "sibu":
				printertable = printersibu
			case "kapit":
				printertable = printerkapit
			}

			_, err := db.Exec(`INSERT INTO ` + printertable + ` (printermodel, printerno, printertype, notes, nickname) VALUES (?, ?, ?, ?, ?)`, printermodel, printerno, printertype, notes, nickname)

			if err != nil {
				log.Println(err)
			} else {
				//success
				http.Redirect(w, r, "/itdb/printer/" + office + "", 302)
			}
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else{
		http.Redirect(w, r, "/", 302)
	}
}