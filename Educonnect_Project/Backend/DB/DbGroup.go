package DB

import (
	"database/sql"
	"fmt"
	"github.com/ipt-9/EduConnect/utils"
	"log"
	"time"
)

type Group struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   uint64    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	InviteCode  string    `json:"invite_code"`
}

// CreateGroup erstellt eine neue Gruppe mit Invite-Code und trägt den Ersteller als Admin ein
func CreateGroup(name string, description string, createdBy uint64) (*Group, error) {
	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Invite-Code generieren
	inviteCode := utils.GenerateInviteCode()

	// 2. Gruppe mit Invite-Code einfügen
	result, err := tx.Exec(`
		INSERT INTO user_groups (name, description, created_by, invite_code)
		VALUES (?, ?, ?, ?)`,
		name, description, createdBy, inviteCode)
	if err != nil {
		return nil, err
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 3. Ersteller als Admin einfügen
	_, err = tx.Exec(`
		INSERT INTO group_members (group_id, user_id, role)
		VALUES (?, ?, 'admin')`,
		groupID, createdBy)
	if err != nil {
		return nil, err
	}

	// 4. Commit
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &Group{
		ID:          uint64(groupID),
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(), // Alternativ: per SELECT NOW() holen
		InviteCode:  inviteCode,
	}, nil
}

func JoinGroupByCode(inviteCode string, userID uint64) error {
	var groupID uint64

	// 1. Gruppe mit Code finden
	err := DB.QueryRow(`SELECT id FROM user_groups WHERE invite_code = ?`, inviteCode).Scan(&groupID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Ungültiger Einladungscode")
	} else if err != nil {
		return err
	}

	// 2. Mitgliedseintrag einfügen (nur wenn nicht vorhanden)
	_, err = DB.Exec(`
		INSERT INTO group_members (group_id, user_id, role)
		VALUES (?, ?, 'member')
		ON DUPLICATE KEY UPDATE role = role`, // verhindert Duplikat-Fehler
		groupID, userID)

	return err
}

type GroupMember struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func GetGroupMembers(groupID uint64) ([]GroupMember, error) {
	rows, err := DB.Query(`
		SELECT u.id, u.username, u.email, gm.role
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ?
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []GroupMember
	for rows.Next() {
		var m GroupMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.Email, &m.Role); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}
func RemoveGroupMember(groupID, targetUserID, requesterID uint64) error {
	// 1. Prüfen, ob der Requester Admin ist
	var role string
	err := DB.QueryRow(`
		SELECT role FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, requesterID).Scan(&role)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Du bist kein Mitglied dieser Gruppe")
	}
	if err != nil {
		return err
	}
	if role != "admin" {
		return fmt.Errorf("Nur Admins dürfen Mitglieder entfernen")
	}

	// 2. Entfernen
	_, err = DB.Exec(`
		DELETE FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, targetUserID)

	return err
}

func UpdateMemberRole(groupID, targetUserID, requesterID uint64, newRole string) error {
	// 1. Rolle vom Requester prüfen
	var requesterRole string
	err := DB.QueryRow(`
		SELECT role FROM group_members
		WHERE group_id = ? AND user_id = ?`,
		groupID, requesterID).Scan(&requesterRole)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Du bist kein Mitglied dieser Gruppe")
	}
	if err != nil {
		return err
	}
	if requesterRole != "admin" {
		return fmt.Errorf("Nur Admins dürfen Rollen ändern")
	}

	// 2. Selbstschutz: Admin darf sich nicht selbst runterstufen
	if requesterID == targetUserID && newRole != "admin" {
		return fmt.Errorf("Du kannst dich nicht selbst runterstufen")
	}

	// 3. Rolle aktualisieren
	_, err = DB.Exec(`
		UPDATE group_members
		SET role = ?
		WHERE group_id = ? AND user_id = ?`,
		newRole, groupID, targetUserID)

	return err
}
func SaveGroupMessage(groupID, userID uint64, message string, messageType string, linkedTaskId *uint64) error {
	log.Printf("💾 INSERT group_message | Group: %d | User: %d | Type: %s", groupID, userID, messageType)

	res, err := DB.Exec(`
		INSERT INTO group_messages (group_id, user_id, message, message_type, linked_task_id, created_at)
		VALUES (?, ?, ?, ?,?, NOW())
	`, groupID, userID, message, messageType, linkedTaskId)

	if err != nil {
		log.Printf("❌ Fehler beim INSERT: %v", err)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("⚠️ Fehler beim Auslesen von RowsAffected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("⚠️ WARNUNG: Der Insert hat 0 Zeilen verändert (keine Speicherung erfolgt!)")
	} else {
		log.Printf("✅ %d Zeile(n) erfolgreich eingefügt", rowsAffected)
	}

	return nil
}

type GroupChatMessage struct {
	Message      string    `json:"message"`
	MessageType  string    `json:"message_type"`
	LinkedTaskID *uint64   `json:"linked_task_id,omitempty"` // ✅ Spaltenname wie in DB!
	CreatedAt    time.Time `json:"created_at"`
	User         struct {
		ID                uint64  `json:"id"`
		Username          string  `json:"username"`
		Email             string  `json:"email"`
		ProfilePictureUrl *string `json:"profile_picture_url"`
	} `json:"user"`
}

func GetFullGroupMessages(userID uint64, groupID uint64, limit int) ([]GroupChatMessage, error) {
	// Prüfen, ob User Mitglied der Gruppe ist
	isMember, err := IsUserMemberOfGroup(userID, groupID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("Zugriff verweigert: kein Mitglied der Gruppe %d", groupID)
	}

	// Nachrichten aus der Gruppe holen
	rows, err := DB.Query(`
		SELECT 
			gm.message,
			gm.created_at,
			gm.message_type,
			gm.linked_task_id,
			u.id,
			u.username,
			u.email,
			u.profile_picture_url
		FROM group_messages gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ?
		ORDER BY gm.created_at DESC
		LIMIT ?
	`, groupID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []GroupChatMessage
	for rows.Next() {
		var msg GroupChatMessage
		var linkedTaskID *uint64

		err := rows.Scan(
			&msg.Message,
			&msg.CreatedAt,
			&msg.MessageType,
			&linkedTaskID,
			&msg.User.ID,
			&msg.User.Username,
			&msg.User.Email,
			&msg.User.ProfilePictureUrl,
		)
		if err != nil {
			return nil, err
		}

		msg.LinkedTaskID = linkedTaskID
		messages = append(messages, msg)
	}

	return messages, nil
}
func IsUserMemberOfGroup(userID uint64, groupID uint64) (bool, error) {
	row := DB.QueryRow(`
		SELECT COUNT(*)
		FROM group_members
		WHERE user_id = ? AND group_id = ?
	`, userID, groupID)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetGroupsForUser(userID uint64) ([]Group, error) {
	rows, err := DB.Query(`
		SELECT g.id, g.name, g.description, g.created_by, g.created_at, g.invite_code
		FROM user_groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedBy, &g.CreatedAt, &g.InviteCode); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	return groups, nil
}
func GetGroupByID(groupID uint64) (*Group, error) {
	var g Group
	err := DB.QueryRow(`
		SELECT id, name, description, created_by, created_at, invite_code
		FROM user_groups
		WHERE id = ?
	`, groupID).Scan(&g.ID, &g.Name, &g.Description, &g.CreatedBy, &g.CreatedAt, &g.InviteCode)

	if err != nil {
		return nil, err
	}
	return &g, nil
}

func GetUserByID(userID uint64) (User, error) {
	var user User
	err := DB.QueryRow(`
		SELECT id, username, email, profile_picture_url
		FROM users
		WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.ProfilePictureUrl)
	return user, err
}

func GetGroupIDByInviteCode(code string) (uint64, error) {
	var groupID uint64
	err := DB.QueryRow(`
		SELECT id FROM user_groups
		WHERE invite_code = ?
	`, code).Scan(&groupID)

	if err != nil {
		return 0, err
	}

	return groupID, nil
}

func CountAdminsInGroup(groupID uint64) (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) 
		FROM group_members 
		WHERE group_id = ? AND role = 'admin'`,
		groupID).Scan(&count)
	return count, err
}

func IsUserAdminInGroup(groupID uint64, userID uint64) (bool, error) {
	var role string
	err := DB.QueryRow(`
		SELECT role 
		FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, userID).Scan(&role)
	if err != nil {
		return false, err
	}
	return role == "admin", nil
}

func SelfLeaveGroup(groupID, userID uint64) error {
	_, err := DB.Exec(`
		DELETE FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, userID)
	return err
}

type Submission struct {
	Code          string
	Output        string
	ExecutionTime int
}

func GetSubmissionByTaskAndUser(taskID uint64, userID uint64) (*Submission, error) {
	var sub Submission
	err := DB.QueryRow(`
		SELECT code, output, execution_time_ms
		FROM submissions
		WHERE task_id = ? AND user_id = ? AND is_successful = 1
		ORDER BY submitted_at DESC
		LIMIT 1
	`, taskID, userID).Scan(&sub.Code, &sub.Output, &sub.ExecutionTime)

	if err != nil {
		return nil, err
	}
	return &sub, nil
}

type LastCourseAndTaskInfo struct {
	CourseID        int    `json:"course_id"`
	CourseTitle     string `json:"course_title"`
	Language        string `json:"language"`
	Difficulty      string `json:"difficulty"`
	Topic           string `json:"topic"`
	ProgressPercent int    `json:"progress_percent"`

	TaskID          int    `json:"task_id"`
	TaskTitle       string `json:"task_title"`
	TaskDescription string `json:"task_description"`
}

func GetLastVisitedCourseAndTask(userID uint64) (*LastCourseAndTaskInfo, error) {
	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Hole letzte Submission samt Kurs und Task Details
	query := `
		SELECT c.id, c.topic, c.programming_language, c.difficulty, c.topic,
		       t.id, t.title, t.description
		FROM submissions s
		JOIN tasks t ON s.task_id = t.id
		JOIN courses c ON t.course_id = c.id
		WHERE s.user_id = ?
		ORDER BY s.submitted_at DESC
		LIMIT 1
	`

	row := tx.QueryRow(query, userID)

	var result LastCourseAndTaskInfo
	if err := row.Scan(&result.CourseID, &result.CourseTitle, &result.Language, &result.Difficulty, &result.Topic,
		&result.TaskID, &result.TaskTitle, &result.TaskDescription); err != nil {
		return nil, err
	}

	// Gesamtanzahl Aufgaben im Kurs
	var totalTasks int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM tasks WHERE course_id = ?`, result.CourseID).Scan(&totalTasks); err != nil {
		return nil, err
	}

	// Erledigte Aufgaben zählen
	var completedTasks int
	if err := tx.QueryRow(`
		SELECT COUNT(DISTINCT t.id)
		FROM submissions s
		JOIN tasks t ON s.task_id = t.id
		WHERE s.user_id = ? AND t.course_id = ? AND s.is_successful = 1
	`, userID, result.CourseID).Scan(&completedTasks); err != nil {
		return nil, err
	}

	// Berechne den Fortschritt
	progressPercent := 0
	if totalTasks > 0 {
		progressPercent = int(float64(completedTasks) / float64(totalTasks) * 100)
	}
	result.ProgressPercent = progressPercent

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &result, nil
}
func CountCompletedCourses(userID uint64) (int, error) {
	row := DB.QueryRow(`
		SELECT COUNT(*)
		FROM user_courses
		WHERE user_id = ? AND completed_at IS NOT NULL
	`, userID)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
func CountCompletedTasks(userID uint64) (int, error) {
	row := DB.QueryRow(`
		SELECT COUNT(*)
		FROM submissions
		WHERE user_id = ?
	`, userID)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
func ActivateUserSubscription(userID uint64) error {
	log.Printf("📥 Versuche, Subscription für user_id=%d zu aktivieren...", userID)

	result, err := DB.Exec(`UPDATE users SET has_subscription = TRUE WHERE id = ?`, userID)
	if err != nil {
		log.Printf("❌ Fehler beim Update: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("⚠️ Fehler beim Abrufen der RowsAffected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("⚠️ Keine Zeile aktualisiert für user_id=%d", userID)
		return fmt.Errorf("kein Benutzer mit user_id=%d gefunden", userID)
	}

	log.Printf("✅ Subscription für user_id=%d erfolgreich aktiviert", userID)
	return nil
}
func CheckUserSubscription(userID uint64) (bool, error) {
	var hasSubscription bool
	err := DB.QueryRow("SELECT has_subscription FROM users WHERE id = ?", userID).Scan(&hasSubscription)
	return hasSubscription, err
}
