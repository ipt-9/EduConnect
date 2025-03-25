package main

import (
	"fmt"
	"github.com/ipt-9/EduConnect/DB"
	"log"
	"net/http"
)

func main() {
	// Verbindung zur Datenbank herstellen
	if err := DB.Connect(); err != nil {
		log.Fatalf("Fehler bei der Verbindung zur Datenbank: %v", err)
	}
	defer DB.Close()

	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/protected", protected)
	http.ListenAndServe(":8080", nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid method", er)
		return
	}

	users := DB.ReadUsers()
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	if _, exists := users[username]; exists {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	} else {
		if len(username) >= 3 && email != "" && len(password) >= 8 {
			fmt.Println(username, email, password)
			DB.CreateUser(username, email, password)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User registered successfully"))
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid method", er)
		return
	}
}

func protected(w http.ResponseWriter, r *http.Request) {

}
