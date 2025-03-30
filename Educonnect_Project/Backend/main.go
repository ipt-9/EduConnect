package main

import (
	"github.com/gorilla/mux"
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

	r := mux.NewRouter()
	routes.InitJWT()

	// 📌 Alle bisherigen Routen
	r.HandleFunc("/register", routes.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/login", routes.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/protected", routes.Protected).Methods("GET", "OPTIONS")
	r.HandleFunc("/logout", routes.Logout).Methods("POST", "OPTIONS")
	r.HandleFunc("/verify-2fa", routes.Verify2FA).Methods("POST", "OPTIONS")
	r.HandleFunc("/me", routes.Me).Methods("GET", "OPTIONS")
	r.HandleFunc("/my-courses", routes.GetMyCourses).Methods("GET", "OPTIONS")

	// 🆕 Neue REST-Route mit Pfadparameter
	r.HandleFunc("/courses/{id}/tasks", routes.GetTasksByCourse).Methods("GET", "OPTIONS")

	log.Println("🚀 Server läuft auf http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
