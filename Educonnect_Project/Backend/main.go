package main

import (
	"github.com/ipt-9/EduConnect/DB"
	"github.com/ipt-9/EduConnect/Routes"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
	if err := DB.Connect(); err != nil {
		log.Fatalf("Fehler bei der Verbindung zur Datenbank: %v", err)
	}
	defer DB.Close()

	err := godotenv.Load("configuration.env")
	if err != nil {
		log.Fatal("❌ Fehler beim Laden der .env Datei")
	}

	http.HandleFunc("/register", routes.Register)
	http.HandleFunc("/login", routes.Login)
	http.HandleFunc("/protected", routes.Protected)
	http.HandleFunc("/logout", routes.Logout)
	http.HandleFunc("/verify-2fa", routes.Verify2FA)

	log.Println("🚀 Server läuft auf http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
