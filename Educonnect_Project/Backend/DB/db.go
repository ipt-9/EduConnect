package DB

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

var DB *sql.DB

func Connect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		"EduAdmin", "EduPasswort123", "138.199.221.113", 3306, "EduDB")

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("Fehler beim Öffnen der Datenbank: %v", err)
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	if err = DB.Ping(); err != nil {
		DB.Close()
		return fmt.Errorf("Fehler bei der Verbindung zur Datenbank: %v", err)
	}

	log.Println("Erfolgreich mit dem Webserver verbunden!")
	return nil
}

func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Fehler beim Schließen der Datenbank: %v", err)
		} else {
			log.Println("Datenbankverbindung geschlossen.")
		}
	}
}

func CreateUser(username, password, email string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Fehler beim Hashen des Passworts: %v", err)
	}

	_, err = DB.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, hashedPassword)
	if err != nil {
		log.Fatalf("Fehler beim Erstellen des Users: %v", err)
	} else {
		log.Println("Benutzer erfolgreich erstellt")
	}
}

func ValidateUser(email, password string) bool {
	var storedHash string
	err := DB.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&storedHash)
	if err != nil {
		log.Println("Benutzer nicht gefunden oder Fehler:", err)
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		log.Println("Passwort stimmt nicht")
		return false
	}

	return true
}
func GetUserIDByEmail(email string) (uint64, error) {
	var id uint64
	err := DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("User-ID konnte nicht gefunden werden: %v", err)
	}
	return id, nil
}

func StoreToken(userID uint64, token string, issuedAt, expiresAt time.Time) error {
	_, err := DB.Exec(
		"INSERT INTO tokens (user_id, token, issued_at, expires_at) VALUES (?, ?, ?, ?)",
		userID, token, issuedAt, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("Fehler beim Speichern des Tokens: %v", err)
	}
	log.Printf("📝 Token f\u00fcr user_id %d gespeichert\n", userID)
	return nil
}
func DeleteToken(token string) error {
	_, err := DB.Exec("DELETE FROM tokens WHERE token = ?", token)
	return err
}
func Store2FACode(userID uint64, code string, expiresAt time.Time) error {
	_, err := DB.Exec(`
		INSERT INTO email_2fa_tokens (user_id, code, expires_at)
		VALUES (?, ?, ?)
	`, userID, code, expiresAt)
	if err != nil {
		return fmt.Errorf("Fehler beim Speichern des 2FA-Codes: %v", err)
	}
	log.Printf("📧 2FA-Code für user_id %d gespeichert\n", userID)
	return nil
}

func Validate2FACode(userID uint64, code string) bool {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM email_2fa_tokens 
		WHERE user_id = ? AND code = ? AND expires_at > NOW()
	`, userID, code).Scan(&count)
	if err != nil {
		log.Printf("Fehler bei der 2FA-Code-Prüfung: %v", err)
		return false
	}
	return count > 0
}
func Delete2FACode(userID uint64) error {
	_, err := DB.Exec("DELETE FROM email_2fa_tokens WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("Fehler beim Löschen des 2FA-Codes: %v", err)
	}
	return err
}
