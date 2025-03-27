package main

import (
	"github.com/ipt-9/EduConnect/DB"
	"github.com/ipt-9/EduConnect/Routes"
	"log"
	"net/http"
)

func main() {
	if err := DB.Connect(); err != nil {
		log.Fatalf("Fehler bei der Verbindung zur Datenbank: %v", err)
	}
	defer DB.Close()

	http.HandleFunc("/register", routes.Register)
	http.HandleFunc("/login", routes.Login)
	http.HandleFunc("/protected", routes.Protected)
	http.HandleFunc("/logout", routes.Logout)

	log.Println("🚀 Server läuft auf http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
