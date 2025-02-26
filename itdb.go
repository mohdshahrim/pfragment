package main

import (
	"log"
	"strconv"
	"net/http"
	"strings"
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
	r.HandleFunc("/itdb/pc/{office}/edit/{id}", PageITDBPCEdit) // PC Edit
	r.HandleFunc("/itdb/pc/{office}/edit/{id}/submit", ITDBPCEditSubmit)
	r.HandleFunc("/itdb/pc/{office}/view/{id}", PageITDBPCView)
	r.HandleFunc("/itdb/pc/{office}/delete/{id}", ITDBPCDelete)
	r.HandleFunc("/itdb/printer/{office}", PageITDBPrinter)
	r.HandleFunc("/itdb/printer/{office}/add", PageITDBPrinterAdd)
	r.HandleFunc("/itdb/printer/{office}/add/submit", ITDBPrinterAddSubmit)
	r.HandleFunc("/itdb/printer/{office}/edit/{rowid}", PageITDBPrinterEdit)
	r.HandleFunc("/itdb/printer/{office}/edit/{rowid}/submit", ITDBPrinterEditSubmit)
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

// "/itdb/pc/{office}/edit/{id}"
func PageITDBPCEdit(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			office := mux.Vars(r)["office"]
			id := mux.Vars(r)["id"] // because pc tables use id instead of rowid
			idInt,_ := strconv.Atoi(id)

			userbasic := PageITDBStruct {
				"",
				username,
				"",
				usergroup,
			}

			data := struct{
				Office string
				PageITDBStruct PageITDBStruct
				PC	PC
				Printers []Printer
			}{
				office,
				userbasic,
				GetPCById(office, idInt),
				//GetPrinter(office),
				append(GetPrinterNoHost(office), HostedPrinters(office,idInt)...),
			}

			tmpl := template.Must(template.ParseFiles("template/itdb/editpc.html"))
			tmpl.Execute(w, data)
		} else{
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

// page just to display PC in tabular form for easier view
func PageITDBPCView(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			office := mux.Vars(r)["office"]
			id := mux.Vars(r)["id"] // because pc tables use id instead of rowid
			idInt,_ := strconv.Atoi(id)

			userbasic := PageITDBStruct {
				"",
				username,
				"",
				usergroup,
			}

			data := struct{
				Office string
				PageITDBStruct PageITDBStruct
				PC	PC
				Printers []Printer
			}{
				office,
				userbasic,
				GetPCById(office, idInt),
				GetPrinterNoHost(office),
			}

			tmpl := template.Must(template.ParseFiles("template/itdb/viewpc.html"))
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

// /itdb/printer/{office}/edit/{rowid}
func PageITDBPrinterEdit(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		username, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			office := mux.Vars(r)["office"]
			rowid := mux.Vars(r)["rowid"] // because pc tables use id instead of rowid
			rowidInt,_ := strconv.Atoi(rowid)

			userbasic := PageITDBStruct {
				"",
				username,
				"",
				usergroup,
			}

			data := struct{
				Office string
				PageITDBStruct PageITDBStruct
				Printer	Printer
			}{
				office,
				userbasic,
				GetPrinterByRowid(office, rowidInt),
			}

			tmpl := template.Must(template.ParseFiles("template/itdb/editprinter.html"))
			tmpl.Execute(w, data)
		} else{
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

// Functions that handles process and procedures and does not involve returning HTML page
//
//

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
		log.Fatal("GetPrinterNoHost() no rows ", err)
	} else if err != nil {
		log.Fatal("func GetPrinterNoHost() return err nil ", err)
	}

    defer row.Close()
    for row.Next() {
        printer := Printer{}
		printer.Office = office
        err := row.Scan(&printer.Rowid, &printer.Printermodel, &printer.Printerno, &printer.Printertype, &printer.Notes, &printer.Host, &printer.Nickname)
        if err != nil {
            log.Fatal(err)
        }
        printerstruct = append(printerstruct, printer)
    }

    return printerstruct
}


func HostedPrinters(office string, id int) []Printer {
	db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
	if errOpen != nil {
		log.Fatal("error opening itdb.db ", errOpen)
	}
	defer db.Close()

    var printerstruct []Printer

	query := ""

	switch(office) {
	case "sibu":
		query = "SELECT rowid, * FROM " + printersibu + " WHERE host=?"
	case "kapit":
		query = "SELECT rowid, * FROM " + printerkapit + " WHERE host=?"
	}

    row, err := db.Query(query, id)
	
	if err == sql.ErrNoRows {
		log.Fatal("HostedPrinters() no rows ", err)
	} else if err != nil {
		log.Fatal("HostedPrinters() return err nil ", err)
	}

    defer row.Close()
    for row.Next() {
        printer := Printer{}
		printer.Office = office
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
		log.Fatal("GetPC() ", err)
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

// function to get PC by its row id (not rowid)
func GetPCById(office string, id int) PC {
	db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
	if errOpen != nil {
		log.Fatal("error opening itdb.db ", errOpen)
	}
	defer db.Close()

	query := ""

	switch(office) {
	case "sibu":
		query = "SELECT * FROM " + pcsibu + " WHERE id=?"
	case "kapit":
		query = "SELECT * FROM " + pckapit + " WHERE id=?"
	}

	pcstruct := PC{}

	err := db.QueryRow(query, id).Scan(&pcstruct.Id, &pcstruct.Hostname, &pcstruct.Ip, &pcstruct.Cpumodel, &pcstruct.Cpuno, &pcstruct.Monitormodel, &pcstruct.Monitorno, &pcstruct.Printer, &pcstruct.User, &pcstruct.Department, &pcstruct.Notes)

	if err == sql.ErrNoRows {
		log.Fatal("GetPCById ", err)
	} else if err != nil {
		log.Fatal(err)
	}

	pcstruct.Office = office //most likely is needed

    return pcstruct
}

// function to get printer by its rowid
func GetPrinterByRowid(office string, rowid int) Printer {
	db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
	if errOpen != nil {
		log.Fatal("error opening itdb.db ", errOpen)
	}
	defer db.Close()

	query := ""

	switch(office) {
	case "sibu":
		query = "SELECT rowid, * FROM " + printersibu + " WHERE rowid=?"
	case "kapit":
		query = "SELECT rowid, * FROM " + printerkapit + " WHERE rowid=?"
	}

	printerstruct := Printer{}

	err := db.QueryRow(query, rowid).Scan(&printerstruct.Rowid, &printerstruct.Printermodel, &printerstruct.Printerno, &printerstruct.Printertype, &printerstruct.Notes, &printerstruct.Host, &printerstruct.Nickname)

	if err == sql.ErrNoRows {
		log.Fatal("GetPrinterByRowid ", err)
	} else if err != nil {
		log.Fatal(err)
	}

	printerstruct.Office = office //most likely is needed

    return printerstruct
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
		log.Fatal("GetPrinter() ", err)
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

// function to display printer rowid as printer name
// printerno format "1 2" (number separated by spaces)
func (p PC) PrinterName(office string, printerno string) string {
	if len(printerno) == 0 {
		return ""
	}

	printerRowids := strings.Split(printerno, " ")

	// An uninitialized slice equals to nil and has length 0.
	if len(printerRowids) == 0 {
		return ""
	}

	// loop
	finalString := ""
	for i:=0; i<len(printerRowids); i++ {
		printer := Printer{}
		rowid, _ := strconv.Atoi(printerRowids[i])
		printer = GetPrinterByRowid(office, rowid)
		finalString += printer.Printermodel + " (" + printer.Nickname + ") "
	}

	return finalString
}

func (p Printer) IndexOffset(index int) string {
	index = index + 1
	return strconv.Itoa(index)
}

// function to get hostname for related printer
func (p Printer) PrinterHostname(id int64, office string) string {
	if id == 0 {
		return "n/a"
	} else {
		return GetHostname(int(id),office)
	}
}

// function to determine whether the printer is already hosted, and will return "checked" or ""
func (p Printer) PrinterChecked(office string, rowid int) string {
	checkedStr := ""

	printertable := ""
	switch(office) {
	case "sibu":
		printertable = printersibu
	case "kapit":
		printertable = printerkapit
	}

	db, err := sql.Open("sqlite3", "./database/itdb.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var hostInt sql.NullInt64
	query := `SELECT host FROM ` + printertable + ` WHERE rowid=?`
	err = db.QueryRow(query, rowid).Scan(&hostInt)

	if hostInt.Valid {
		checkedStr = "checked"
	} else {
		checkedStr = ""
	}

	return checkedStr
}

func GetHostname(id int, office string) string {
	hostname := ""

	db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	pctable := ""
	switch(office) {
	case "sibu":
		pctable = pcsibu 
	case "kapit":
		pctable = pckapit
	}


	query := `SELECT hostname FROM ` + pctable + ` WHERE id = ?`
	err := db.QueryRow(query, id).Scan(&hostname)

	if err == sql.ErrNoRows {
		log.Fatal("GetHostname() =",err)
		return ""
	} else if err != nil {
		log.Fatal(err)
		return ""
	}

	return hostname
}

// function to handle add new PC
func ITDBPCAddSubmit(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		_, usergroup := GetUserSession(r)
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

			result, err := db.Exec(`INSERT INTO ` + pctable + ` (hostname, ip, cpu_model, cpu_no, monitor_model, monitor_no, printer, user, department, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, hostname, ip, cpu_model, cpu_no, monitor_model, monitor_no, printer, user, department, notes)

			if err != nil {
				log.Println(err)
			} else {
				if len(printer) != 0 {
					// update the printer too
					lastid, _ := result.LastInsertId() // get last id being inserted on the pc table
					ITDBPrinterHostUpdate(office, printer, int(lastid))
				}

				http.Redirect(w, r, "/itdb/pc/"+office, 302)
			}
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else{
		http.Redirect(w, r, "/", 302)
	}
}

func ITDBPCEditSubmit(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		_, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			r.ParseForm()
			//
			id := r.FormValue("id")
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
				var s []string
				for index := range r.Form["printer"] {
					s = append(s, r.Form["printer"][index])
				}
				printer = strings.Join(s, " ")
			}

			// procedures performed before the update
			intid, _ := strconv.Atoi(id)
			hostedprinters := ITDBGetHostedPrinters(office, intid)
			if len(hostedprinters)!=0 {
				// split into string slices
				var s []string
				s = strings.Fields(hostedprinters)
				
				// loop to update every printer rowids
				for index := range s {
					intindex, _ := strconv.Atoi(s[index])
					ITDBPrinterSetHostEmpty(office, intindex)
				}
			}


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

			query := `UPDATE ` + pctable + ` SET hostname=?, ip=?, cpu_model=?, cpu_no=?, monitor_model=?, monitor_no=?, printer=?, user=?, department=?, notes=? WHERE id = ?`
			_, err := db.Exec(query, hostname, ip, cpu_model, cpu_no, monitor_model, monitor_no, printer, user, department, notes, id)
			
			if err != nil {
				log.Println(err)
			} else {
				if len(printer) != 0 {
					// update the printer too
					idInt, _ := strconv.Atoi(id)
					ITDBPrinterHostUpdate(office, printer, idInt)
				}

				http.Redirect(w, r, "/itdb/pc/" + office + "/view/" + id, 302)
			}
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

func ITDBPCDelete(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		_, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			office := mux.Vars(r)["office"]
			id := mux.Vars(r)["id"] // because pc tables use id instead of rowid
			idInt,_ := strconv.Atoi(id)

			// decides which table
			pctable := ""
			switch(office) {
			case "sibu":
				pctable = pcsibu
			case "kapit":
				pctable = pckapit
			}

			db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
			if errOpen != nil {
				log.Fatal(errOpen)
			}
			defer db.Close()

			query := `DELETE FROM ` + pctable + ` WHERE id = ?`
			_, err := db.Exec(query, idInt)
			if err != nil {
				log.Fatal(err)
			}

			http.Redirect(w, r, "/itdb/pc/"+office, 302)
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else {
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

func ITDBPrinterEditSubmit(w http.ResponseWriter, r *http.Request) {
	if IsAuthenticated(w,r) {
		_, usergroup := GetUserSession(r)
		if AccessITDB(usergroup) {
			rowid := r.FormValue("rowid")
			office := r.FormValue("office")
			printermodel := r.FormValue("printermodel")
			printerno := r.FormValue("printerno")
			printertype := r.FormValue("printertype")
			notes := r.FormValue("notes")
			nickname := r.FormValue("nickname")


			// begin procedure of updating password
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

			query := `UPDATE ` + printertable + ` SET printermodel=?, printerno=?, printertype=?, notes=?, nickname=? WHERE rowid = ?`
			_, err := db.Exec(query, printermodel, printerno, printertype, notes, nickname, rowid)
			if err != nil {
				log.Fatal(err)
			}
			
			http.Redirect(w, r, "/itdb/printer/" + office, 302)
		} else {
			http.Redirect(w, r, "/user", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

// function to update printer host column
// remember: the host column is FK
// this could involve multple printers/devices
func ITDBPrinterHostUpdate(office string, printer string, pcid int) {
	printertable := ""
	switch(office) {
	case "sibu":
		printertable = printersibu
	case "kapit":
		printertable = printerkapit
	}

	// printer means the format used in storing printer as in PC
	printerRowids := strings.Split(printer, " ")

	db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	for i:=0; i<len(printerRowids); i++ {
		query := `UPDATE ` + printertable + ` SET host=? WHERE rowid=?`
		_, err := db.Exec(query, pcid, printerRowids[i])
		if err != nil {
			log.Fatal(err)
		}
	}
}

// function to get value on printer field
func ITDBGetHostedPrinters(office string, pcid int) string {
	pctable := ""
	switch(office) {
	case "sibu":
		pctable = pcsibu
	case "kapit":
		pctable = pckapit
	}

	strPrinter := ""

	db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	query := `SELECT printer FROM `+pctable+` WHERE id = ?`
	err := db.QueryRow(query, pcid).Scan(&strPrinter)

	if err == sql.ErrNoRows {
		log.Fatal(err)
	} else if err != nil {
		log.Fatal(err)
	}

	return strPrinter
}

func ITDBPrinterSetHostEmpty(office string, rowid int) {
	printertable := ""
	switch(office) {
	case "sibu":
		printertable = printersibu
	case "kapit":
		printertable = printerkapit
	}

	db, errOpen := sql.Open("sqlite3", "./database/itdb.db")
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	query := `UPDATE ` + printertable + ` SET host = NULL WHERE rowid = ?`
	db.Exec(query, rowid)
}