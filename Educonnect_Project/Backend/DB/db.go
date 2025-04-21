package DB

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
	"time"
	"unicode"
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
func AssignAllCoursesToUser(userID uint64) error {
	log.Printf("📥 Versuche, alle Kurse user_id=%d zuzuweisen...", userID)

	// Alle Kurs-IDs aus der Datenbank holen
	rows, err := DB.Query(`SELECT id FROM courses`)
	if err != nil {
		log.Printf("❌ Fehler beim Abrufen der Kurse: %v", err)
		return err
	}
	defer rows.Close()

	var courseIDs []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			log.Printf("❌ Fehler beim Lesen der Kurs-ID: %v", err)
			return err
		}
		courseIDs = append(courseIDs, id)
	}

	// Frühzeitiger Exit, falls keine Kurse vorhanden sind
	if len(courseIDs) == 0 {
		log.Println("ℹ️ Keine Kurse gefunden – keine Zuweisung vorgenommen.")
		return nil
	}

	// Mehrere INSERTS vorbereiten
	tx, err := DB.Begin()
	if err != nil {
		log.Printf("❌ Fehler beim Starten der Transaktion: %v", err)
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO user_courses (user_id, course_id) VALUES (?, ?)`)
	if err != nil {
		log.Printf("❌ Fehler beim Vorbereiten des Statements: %v", err)
		return err
	}
	defer stmt.Close()

	var inserted int64
	for _, courseID := range courseIDs {
		result, err := stmt.Exec(userID, courseID)
		if err != nil {
			log.Printf("⚠️ Fehler beim Einfügen (user_id=%d, course_id=%d): %v", userID, courseID, err)
			continue
		}
		rowsAffected, _ := result.RowsAffected()
		inserted += rowsAffected
	}

	if err := tx.Commit(); err != nil {
		log.Printf("❌ Fehler beim Commit: %v", err)
		return err
	}

	log.Printf("✅ Erfolgreich %d Kurse für user_id=%d zugewiesen.", inserted, userID)
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
	CourseID       uint64 `json:"course_id"`
}

func GetTasksForCourse(courseID uint64, userID uint64) ([]TaskWithProgress, error) {
	rows, err := DB.Query(`
		SELECT 
			t.id, t.title, t.description, t.starter_code, t.expected_input, t.expected_output,
			EXISTS (
				SELECT 1 FROM submissions s
				WHERE s.user_id = ? AND s.task_id = t.id AND s.is_successful = 1
			) AS completed,
			t.course_id  -- 🆕 auch mit abfragen!
		FROM tasks t
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
			&task.ID,
			&task.Title,
			&task.Description,
			&task.StarterCode,
			&task.ExpectedInput,
			&task.ExpectedOutput,
			&task.Completed,
			&task.CourseID,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

type SubmissionInput struct {
	UserID          uint64  `json:"-"`
	TaskID          uint64  `json:"task_id"`
	Code            string  `json:"code"`
	Output          string  `json:"output"`
	ErrorMessage    *string `json:"error_message"`
	ExecutionTimeMs *int    `json:"execution_time_ms"`
	UsedHint        *bool   `json:"used_hint"`
	Tip             *string `json:"tip,omitempty"`
}

func SaveSubmissionAndUpdateProgress(input SubmissionInput) (uint64, bool, error, error) {
	log.Printf("💾 Submission verarbeiten für user_id=%d, task_id=%d", input.UserID, input.TaskID)

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("❌ Fehler bei Transaktionsstart: %v", err)
		return 0, false, err, nil
	}
	defer tx.Rollback()

	var expectedOutput string
	err = tx.QueryRow(`SELECT expected_output FROM tasks WHERE id = ?`, input.TaskID).Scan(&expectedOutput)
	if err != nil {
		log.Printf("❌ Fehler beim expected_output-Query: %v", err)
		return 0, false, err, nil
	}

	// 🧽 Robuste Clean-Funktion
	clean := func(s string) string {
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
		s = strings.ReplaceAll(s, "\ufeff", "") // BOM
		s = strings.ReplaceAll(s, "\u200B", "") // Zero-width
		s = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) && r != '\n' && r != '\t' {
				return -1
			}
			return r
		}, s)
		lines := strings.Split(s, "\n")
		for i := range lines {
			lines[i] = strings.TrimSpace(lines[i])
		}
		return strings.TrimSpace(strings.Join(lines, "\n"))
	}

	cleanedUserOutput := clean(input.Output)
	cleanedExpectedOutput := clean(expectedOutput)

	// 📊 Vergleich
	isCorrect := false
	if strings.Contains(" "+expectedOutput+" ", " or ") {
		parts := strings.Split(expectedOutput, "or")
		leftSide := clean(parts[0])
		rightSide := clean(parts[1])
		isCorrect = cleanedUserOutput == leftSide || cleanedUserOutput == rightSide
	} else {
		isCorrect = cleanedUserOutput == cleanedExpectedOutput
	}

	// 📋 Debug-Ausgabe
	log.Printf("📤 Raw Benutzeroutput: %s", input.Output)
	log.Printf("📤 Cleaned Benutzeroutput: %q", cleanedUserOutput)
	log.Printf("🎯 Cleaned Erwartet: %q", cleanedExpectedOutput)
	log.Printf("📦 Bytes Benutzer:  %v", []byte(cleanedUserOutput))
	log.Printf("📦 Bytes Erwartet: %v", []byte(cleanedExpectedOutput))
	log.Printf("📊 Ergebnis: %v", isCorrect)

	var submissionID uint64 = 0
	if isCorrect {
		log.Printf("✅ Output korrekt – Submission wird gespeichert")

		stmt, err := tx.Prepare(`
			INSERT INTO submissions 
			(user_id, task_id, code, submitted_at, is_successful, output, error_message, execution_time_ms, tip) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			log.Printf("❌ Fehler bei Prepare: %v", err)
			return 0, false, err, nil
		}
		defer stmt.Close()

		var errMsg interface{}
		if input.ErrorMessage != nil {
			errMsg = *input.ErrorMessage
		}
		var execTime interface{}
		if input.ExecutionTimeMs != nil {
			execTime = *input.ExecutionTimeMs
		}
		var tip interface{}
		if input.Tip != nil {
			tip = *input.Tip
		}

		res, err := stmt.Exec(
			input.UserID, input.TaskID, input.Code, time.Now(), true,
			input.Output, errMsg, execTime, tip,
		)
		if err != nil {
			log.Printf("❌ Fehler bei Insert Submission: %v", err)
			return 0, false, err, nil
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			log.Printf("❌ Fehler bei LastInsertId: %v", err)
			return 0, false, err, nil
		}
		submissionID = uint64(lastID)
		log.Printf("📥 Submission-ID %d gespeichert", submissionID)
	} else {
		log.Printf("⚠️ Output falsch – nur Fortschritt wird gespeichert")
	}

	// 🧠 Fortschritt aktualisieren
	var usedHint bool
	if input.UsedHint != nil {
		usedHint = *input.UsedHint
	}

	var progressID uint64
	err = tx.QueryRow(`SELECT id FROM user_task_progress WHERE user_id = ? AND task_id = ?`,
		input.UserID, input.TaskID).Scan(&progressID)

	if err == sql.ErrNoRows {
		log.Printf("ℹ️ Kein Fortschritt gefunden → neuer Eintrag")
		_, err = tx.Exec(`
			INSERT INTO user_task_progress 
			(user_id, task_id, completed, last_submission_id, used_hint, last_attempt_code) 
			VALUES (?, ?, ?, ?, ?, ?)
		`, input.UserID, input.TaskID, isCorrect,
			sql.NullInt64{Int64: int64(submissionID), Valid: submissionID != 0},
			usedHint,
			input.Code,
		)
		if err != nil {
			log.Printf("❌ Fehler bei Insert user_task_progress: %v", err)
			return 0, isCorrect, err, nil
		}
	} else if err == nil {
		log.Printf("🔄 Bestehender Fortschritt wird aktualisiert")
		_, err = tx.Exec(`
			UPDATE user_task_progress 
			SET completed = ?, last_submission_id = ?, used_hint = ?, last_attempt_code = ? 
			WHERE user_id = ? AND task_id = ?
		`, isCorrect,
			sql.NullInt64{Int64: int64(submissionID), Valid: submissionID != 0},
			usedHint,
			input.Code,
			input.UserID,
			input.TaskID,
		)
		if err != nil {
			log.Printf("❌ Fehler bei Update user_task_progress: %v", err)
			return 0, isCorrect, err, nil
		}
	} else {
		log.Printf("❌ Fehler beim Prüfen von user_task_progress: %v", err)
		return 0, isCorrect, err, nil
	}

	if err := tx.Commit(); err != nil {
		log.Printf("❌ Fehler beim Commit: %v", err)
		return 0, isCorrect, err, nil
	}

	log.Printf("✅ Fortschritt gespeichert (submission_id: %d)", submissionID)
	return submissionID, isCorrect, nil, nil
}

func GetSubmittedCodeForTask(userID, taskID uint64) (string, error) {
	var code string
	err := DB.QueryRow(`
		SELECT code FROM submissions 
		WHERE user_id = ? AND task_id = ? AND is_successful = 1
		ORDER BY submitted_at DESC
		LIMIT 1
	`, userID, taskID).Scan(&code)

	if err != nil {
		return "", err
	}
	return code, nil
}

func GetSubmittedOrAttemptedCode(userID, taskID uint64) (string, error) {
	var code string

	// 1. Zuerst versuchen aus submissions (erfolgreich)
	err := DB.QueryRow(`
		SELECT code FROM submissions 
		WHERE user_id = ? AND task_id = ? AND is_successful = 1
		ORDER BY submitted_at DESC
		LIMIT 1
	`, userID, taskID).Scan(&code)

	if err == nil {
		return code, nil
	}

	// 2. Wenn keine erfolgreiche Submission → letzten Versuch aus user_task_progress
	err = DB.QueryRow(`
		SELECT last_attempt_code FROM user_task_progress 
		WHERE user_id = ? AND task_id = ?
	`, userID, taskID).Scan(&code)

	return code, err
}
func GetUsernameByUserID(userID uint64) (string, error) {
	var username string
	err := DB.QueryRow(`SELECT username FROM users WHERE id = ?`, userID).Scan(&username)
	return username, err
}
