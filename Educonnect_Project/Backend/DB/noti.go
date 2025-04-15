package DB

import (
	"time"
)

func GetTaskTitleByID(taskID uint64) (string, error) {
	var title string
	err := DB.QueryRow("SELECT title FROM tasks WHERE id = ?", taskID).Scan(&title)
	return title, err
}

func GetUsernameByID(userID uint64) (string, error) {
	var username string
	err := DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	return username, err
}

func CreateGroupNotification(groupID int64, userID *uint64, notifType string, message string) error {
	query := `INSERT INTO group_notifications (group_id, user_id, type, message) VALUES (?, ?, ?, ?)`
	_, err := DB.Exec(query, groupID, userID, notifType, message)
	return err
}

func GetGroupIDsForUser(userID uint64) ([]int64, error) {
	rows, err := DB.Query(`
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

func IsUserInGroup(groupID int64, userID uint64) (bool, error) {
	var exists bool
	err := DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM group_members
			WHERE group_id = ? AND user_id = ?
		)
	`, groupID, userID).Scan(&exists)
	return exists, err
}

func GetGroupNotifications(groupID int64) ([]map[string]interface{}, error) {
	rows, err := DB.Query(`
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
		if err := rows.Scan(&msg, &createdAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, map[string]interface{}{
			"message":    msg,
			"created_at": createdAt,
		})
	}

	return notifications, nil
}

type SubmissionInfo struct {
	TaskID    uint64 `json:"task_id"`
	TaskTitle string `json:"task_title"`
}

// GetSuccessfulSubmissionsByUser gibt alle erfolgreich gelösten Aufgaben eines Users zurück
func GetSuccessfulSubmissionsByUser(userID uint64) ([]SubmissionInfo, error) {
	rows, err := DB.Query(`
		SELECT s.task_id, t.title
		FROM submissions s
		JOIN tasks t ON s.task_id = t.id
		WHERE s.user_id = ? AND s.is_successful = 1
		GROUP BY s.task_id
		ORDER BY MAX(s.submitted_at) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SubmissionInfo
	for rows.Next() {
		var s SubmissionInfo
		if err := rows.Scan(&s.TaskID, &s.TaskTitle); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}
