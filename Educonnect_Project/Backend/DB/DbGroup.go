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
func CreateGroup(db *sql.DB, name string, description string, createdBy uint64) (*Group, error) {
	tx, err := db.Begin()
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
func JoinGroupByCode(db *sql.DB, inviteCode string, userID uint64) error {
	var groupID uint64

	// 1. Gruppe mit Code finden
	err := db.QueryRow(`SELECT id FROM user_groups WHERE invite_code = ?`, inviteCode).Scan(&groupID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Ungültiger Einladungscode")
	} else if err != nil {
		return err
	}

	// 2. Mitgliedseintrag einfügen (nur wenn nicht vorhanden)
	_, err = db.Exec(`
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

func GetGroupMembers(db *sql.DB, groupID uint64) ([]GroupMember, error) {
	rows, err := db.Query(`
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
func RemoveGroupMember(db *sql.DB, groupID, targetUserID, requesterID uint64) error {
	// 1. Prüfen, ob der Requester Admin ist
	var role string
	err := db.QueryRow(`
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
	_, err = db.Exec(`
		DELETE FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, targetUserID)

	return err
}

func UpdateMemberRole(db *sql.DB, groupID, targetUserID, requesterID uint64, newRole string) error {
	// 1. Rolle vom Requester prüfen
	var requesterRole string
	err := db.QueryRow(`
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
	_, err = db.Exec(`
		UPDATE group_members
		SET role = ?
		WHERE group_id = ? AND user_id = ?`,
		newRole, groupID, targetUserID)

	return err
}
func SaveGroupMessage(db *sql.DB, groupID, userID uint64, message string, messageType string) error {
	log.Printf("💾 INSERT group_message | Group: %d | User: %d | Type: %s", groupID, userID, messageType)

	_, err := db.Exec(`
		INSERT INTO group_messages (group_id, user_id, message, message_type, created_at)
		VALUES (?, ?, ?, ?, NOW())
	`, groupID, userID, message, messageType)

	if err != nil {
		log.Println("❌ Insert fehlgeschlagen:", err)
	}
	return err
}

type GroupChatMessage struct {
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	User      struct {
		ID                uint64  `json:"id"`
		Username          string  `json:"username"`
		Email             string  `json:"email"`
		ProfilePictureUrl *string `json:"profile_picture_url"`
	} `json:"user"`
	MessageType string
}

func GetFullGroupMessages(db *sql.DB, groupID uint64, limit int) ([]GroupChatMessage, error) {
	rows, err := db.Query(`
		SELECT gm.message, gm.created_at,
		       u.id, u.username, u.email, u.profile_picture_url
		FROM group_messages gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ?
		ORDER BY gm.created_at DESC
		LIMIT ?`, groupID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []GroupChatMessage
	for rows.Next() {
		var msg GroupChatMessage
		err := rows.Scan(
			&msg.Message,
			&msg.CreatedAt,
			&msg.User.ID,
			&msg.User.Username,
			&msg.User.Email,
			&msg.User.ProfilePictureUrl,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
func GetGroupsForUser(db *sql.DB, userID uint64) ([]Group, error) {
	rows, err := db.Query(`
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
func GetGroupByID(db *sql.DB, groupID uint64) (*Group, error) {
	var g Group
	err := db.QueryRow(`
		SELECT id, name, description, created_by, created_at, invite_code
		FROM user_groups
		WHERE id = ?
	`, groupID).Scan(&g.ID, &g.Name, &g.Description, &g.CreatedBy, &g.CreatedAt, &g.InviteCode)

	if err != nil {
		return nil, err
	}
	return &g, nil
}
func GetUserByID(db *sql.DB, userID uint64) (User, error) {
	var user User
	err := db.QueryRow(`
		SELECT id, username, email, profile_picture_url
		FROM users
		WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.ProfilePictureUrl)
	return user, err
}
func GetGroupIDByInviteCode(db *sql.DB, code string) (uint64, error) {
	var groupID uint64
	err := db.QueryRow(`
		SELECT id FROM user_groups
		WHERE invite_code = ?
	`, code).Scan(&groupID)

	if err != nil {
		return 0, err
	}

	return groupID, nil
}
func CountAdminsInGroup(db *sql.DB, groupID uint64) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM group_members WHERE group_id = ? AND role = 'admin'", groupID).Scan(&count)
	return count, err
}
func IsUserAdminInGroup(db *sql.DB, groupID uint64, userID uint64) (bool, error) {
	var role string
	err := db.QueryRow("SELECT role FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&role)
	if err != nil {
		return false, err
	}
	return role == "admin", nil
}
func SelfLeaveGroup(db *sql.DB, groupID, userID uint64) error {
	_, err := db.Exec(`
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

func GetSubmissionByTaskAndUser(db *sql.DB, taskID uint64, userID uint64) (*Submission, error) {
	var sub Submission
	err := db.QueryRow(`
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
