package main

import (
    "fmt"
    "time"
    "net/http"
    "os"
    "github.com/gorilla/sessions"
)

var (
    // key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
    key = []byte("super-secret-key")
    //store = sessions.NewCookieStore(key) //NOTE this line determines how to store session. This store in memory
    store = sessions.NewFilesystemStore(SessionDirectory(), key)
)

func secret(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "cookie-name")

    // Check if user is authenticated
    if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    fmt.Println("id ",session.Values["id"]," username ",session.Values["username"])
}

// function to determine authentication
func IsAuthenticated(w http.ResponseWriter, r *http.Request) bool {
    session, _ := store.Get(r, "cookie-name")

    // Check if user is authenticated
    if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
        return false
    } else {
        return true
    }
}

func login(w http.ResponseWriter, r *http.Request, id string, username string) {
    session, _ := store.Get(r, "cookie-name")

    // Set user as authenticated
    session.Values["authenticated"] = true
    session.Values["id"] = id
    session.Values["username"] = username
    session.Values["loggedon"] = time.Now().Format(time.RFC822)

    session.Save(r, w)
}

func logout(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "cookie-name")

    // Revoke users authentication
    session.Values["authenticated"] = false
    session.Values["id"] = ""
    session.Values["username"] = ""
    session.Values["loggedon"] = ""

    session.Save(r, w)
}

// function to return directory for storing session
func SessionDirectory() string {
    str, _ := os.Getwd()
    return str + "/session"
}