package DB

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var DB *sql.DB

func Connect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		"root", "Kunstturnen07", "localhost", 3306, "EduConnect")

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("Fehler beim Öffnen der Datenbank: %v", err)
	}

	DB.SetMaxOpenConns(10) // Maximale Anzahl geöffneter Verbindungen
	DB.SetMaxIdleConns(5)  // Maximale Anzahl Leerlaufverbindungen

	if err = DB.Ping(); err != nil {
		DB.Close()
		return fmt.Errorf("Fehler bei der Verbindung zur Datenbank: %v", err)
	}

	log.Println("Erfolgreich mit der lokalen MySQL-Datenbank verbunden!")
	return nil
}

// Close schließt die Verbindung zur Datenbank
func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Fehler beim Schließen der Datenbank: %v", err)
		} else {
			log.Println(" Datenbankverbindung geschlossen.")
		}
	}
}

func ReadUsers() map[string]string {
	// Verbindungsprüfung, falls Connect fehlschlägt, direkt nil zurückgeben
	if err := Connect(); err != nil {
		log.Println("Fehler bei der Verbindung zur Datenbank:", err)
		return nil
	}

	rows, err := DB.Query("SELECT Username, Email FROM User")
	if err != nil {
		log.Fatalf("Fehler beim Abrufen der Benutzer: %v", err)
	}
	defer rows.Close()

	userMap := make(map[string]string)
	for rows.Next() {
		var email, username string
		if err := rows.Scan(&username, &email); err != nil {
			log.Println("Fehler beim Lesen einer Zeile:", err)
			continue
		}
		userMap[username] = email
	}
	fmt.Println(userMap)
	return userMap
}

func CreateUser(username, password, email string) {
	if err := Connect(); err != nil {
		log.Fatalf("Fehler bei der Verbindung zur Datenbank: %v", err)
	}

	// Verwende Exec für INSERT-Befehle
	_, err := DB.Exec("INSERT INTO User (Username, Email, Password) VALUES (?, ?, ?)", username, email, password)
	if err != nil {
		log.Fatalf("Fehler beim Erstellen des Users: %v", err)
	} else {
		log.Println("Benutzer erfolgreich erstellt")
	}
}
