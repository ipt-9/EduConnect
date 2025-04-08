package DB

import (
	"database/sql"
	"time"
)

func GetTaskTitleByID(db *sql.DB, taskID uint64) (string, error) {
	var title string
	err := db.QueryRow("SELECT title FROM tasks WHERE id = ?", taskID).Scan(&title)
	return title, err
}
func GetUsernameByID(db *sql.DB, userID uint64) (string, error) {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	return username, err
}
func CreateGroupNotification(db *sql.DB, groupID int64, userID *uint64, notifType string, message string) error {
	query := `INSERT INTO group_notifications (group_id, user_id, type, message) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, groupID, userID, notifType, message)
	return err
}
func GetGroupIDsForUser(db *sql.DB, userID uint64) ([]int64, error) {
	rows, err := db.Query(`
		SELECT group_id FROM group_members
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groupIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, id)
	}

	return groupIDs, nil
}
func IsUserInGroup(db *sql.DB, groupID int64, userID uint64) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM group_members
			WHERE group_id = ? AND user_id = ?
		)
	`, groupID, userID).Scan(&exists)
	return exists, err
}
func GetGroupNotifications(db *sql.DB, groupID int64) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
		SELECT message, created_at
		FROM group_notifications
		WHERE group_id = ?
		ORDER BY created_at DESC
		LIMIT 50
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []map[string]interface{}
	for rows.Next() {
		var msg string
		var createdAt time.Time
		rows.Scan(&msg, &createdAt)

		notifications = append(notifications, map[string]interface{}{
			"message":    msg,
			"created_at": createdAt,
		})
	}

	return notifications, nil
}
