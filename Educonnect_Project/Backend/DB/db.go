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

type User struct {
	ID                uint64  `json:"id"`
	Username          string  `json:"username"`
	Email             string  `json:"email"`
	LegalName         *string `json:"legal_name"`
	PhoneNumber       *string `json:"phone_number"`
	Address           *string `json:"address"`
	ProfilePictureUrl *string `json:"profile_picture_url"`
	PasswordHash      string  `json:"-"`
}

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

func CreateUser(username, email, password string, legalName, phoneNumber, address, profilePictureUrl *string) error {
	if username == "" || email == "" || len(password) < 8 {
		return fmt.Errorf("Ungültige Eingabedaten")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Fehler beim Hashen des Passworts: %v", err)
	}

	_, err = DB.Exec(`
		INSERT INTO users 
		(username, email, password, legal_name, phone_number, address, profile_picture_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, username, email, hashedPassword, legalName, phoneNumber, address, profilePictureUrl)

	if err != nil {
		return fmt.Errorf("Fehler beim Erstellen des Users: %v", err)
	}

	log.Println("✅ Benutzer erfolgreich erstellt")
	return nil
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

func GetUserByEmail(email string) (User, error) {
	var user User
	log.Println("🔎 Suche nach Benutzer mit Email:", email)

	err := DB.QueryRow(`
        SELECT id, username, email, legal_name, phone_number, address, profile_picture_url, password
        FROM users WHERE email = ?
    `, email).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.LegalName, &user.PhoneNumber,
		&user.Address, &user.ProfilePictureUrl,
		&user.PasswordHash,
	)
	if err != nil {
		log.Printf("❌ Fehler beim Laden des Benutzers mit Email %s: %v", email, err)
		return User{}, fmt.Errorf("Benutzer konnte nicht geladen werden: %v", err)
	}

	return user, nil
}
func AssignDefaultCourseToUser(userID uint64) error {
	log.Printf("📥 Versuche Kurs 1 user_id=%d zuzuweisen...", userID)

	result, err := DB.Exec(`
		INSERT INTO user_courses (user_id, course_id)
		VALUES (?, 1)
	`, userID)
	if err != nil {
		log.Printf("❌ SQL-Fehler beim Zuweisen des Kurses: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("❓ RowsAffected konnte nicht gelesen werden: %v", err)
	} else {
		log.Printf("📌 Kurs-Zuweisung: %d Zeile(n) eingefügt", rows)
	}

	return nil
}

type CourseWithStatus struct {
	ID                  uint64 `json:"id"`
	ProgrammingLanguage string `json:"programming_language"`
	Description         string `json:"description"`
	Difficulty          string `json:"difficulty"`
	Topic               string `json:"topic"`
	Started             bool   `json:"started"`
	Completed           bool   `json:"completed"`
}

func GetCoursesForUser(userID uint64) ([]CourseWithStatus, error) {
	log.Printf("🔍 Lade Kurse für user_id: %d", userID)

	rows, err := DB.Query(`
	SELECT c.id, c.programming_language, c.description, c.difficulty, c.topic,
	       IF(uc.started_at IS NOT NULL, TRUE, FALSE) AS started,
	       IF(uc.completed_at IS NOT NULL, TRUE, FALSE) AS completed
	FROM user_courses uc
	JOIN courses c ON uc.course_id = c.id
	WHERE uc.user_id = ?
`, userID)

	if err != nil {
		log.Printf("❌ Fehler bei DB.Query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var courses []CourseWithStatus
	for rows.Next() {
		var course CourseWithStatus
		if err := rows.Scan(
			&course.ID, &course.ProgrammingLanguage, &course.Description,
			&course.Difficulty, &course.Topic,
			&course.Started, &course.Completed,
		); err != nil {
			log.Printf("❌ Scan-Fehler: %v", err)
			return nil, err
		}
		courses = append(courses, course)
	}
	log.Printf("✅ %d Kurse geladen für user_id=%d", len(courses), userID)
	return courses, nil
}

type TaskWithProgress struct {
	ID             uint64 `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	StarterCode    string `json:"starter_code"`
	ExpectedInput  string `json:"expected_input"`
	ExpectedOutput string `json:"expected_output"`
	Completed      bool   `json:"completed"`
}

func GetTasksForCourse(courseID uint64, userID uint64) ([]TaskWithProgress, error) {
	rows, err := DB.Query(`
		SELECT 
			t.id, t.title, t.description, t.starter_code, t.expected_input, t.expected_output,
			COALESCE(utp.completed, FALSE) AS completed
		FROM tasks t
		LEFT JOIN user_task_progress utp 
		  ON t.id = utp.task_id AND utp.user_id = ?
		WHERE t.course_id = ?
	`, userID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []TaskWithProgress
	for rows.Next() {
		var task TaskWithProgress
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.StarterCode,
			&task.ExpectedInput, &task.ExpectedOutput, &task.Completed,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
