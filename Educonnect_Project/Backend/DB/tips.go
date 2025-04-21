package DB

import (
	"database/sql"
	"log"
)

func SaveUserTipUsage(userID, taskID uint64, tipText string) error {
	var tipID uint64
	err := DB.QueryRow(`SELECT id FROM task_tips WHERE task_id = ? AND tip_text = ? LIMIT 1`, taskID, tipText).Scan(&tipID)
	if err != nil {
		log.Printf("⚠️ Tipp nicht gefunden in task_tips: %v", err)
		return err
	}

	_, err = DB.Exec(`INSERT INTO user_tip_usage (user_id, task_id, tip_id) VALUES (?, ?, ?)`, userID, taskID, tipID)
	if err != nil {
		log.Printf("❌ Fehler beim Einfügen in user_tip_usage: %v", err)
	}
	return err
}

type Task struct {
	ID             uint64
	CourseID       uint64
	Title          string
	Description    string
	StarterCode    string
	ExpectedInput  string
	ExpectedOutput string
	Difficulty     string
	Topic          string
}

func GetTaskByID(taskID uint64) (Task, error) {
	var task Task

	err := DB.QueryRow(`
		SELECT id, course_id, title, description, starter_code,
		       expected_input, expected_output, difficulty, topic
		FROM tasks
		WHERE id = ?
	`, taskID).Scan(
		&task.ID,
		&task.CourseID,
		&task.Title,
		&task.Description,
		&task.StarterCode,
		&task.ExpectedInput,
		&task.ExpectedOutput,
		&task.Difficulty,
		&task.Topic,
	)

	if err != nil {
		return Task{}, err
	}

	return task, nil
}
func SaveGeneratedTip(taskID uint64, userOutput, tipText, errorType string) error {
	_, err := DB.Exec(`
		INSERT INTO task_tips (task_id, condition_output, tip_text, error_type, is_default)
		VALUES (?, ?, ?, ?, false)
	`, taskID, userOutput, tipText, errorType)
	if err != nil {
		log.Printf("❌ Fehler beim Speichern des generierten Tipps: %v", err)
	}
	return err
}

func GetTipByTaskAndOutput(taskID uint64, userOutput string) (string, error) {
	var tip string
	err := DB.QueryRow(`
		SELECT tip_text FROM task_tips
		WHERE task_id = ? AND condition_output = ?
		LIMIT 1
	`, taskID, userOutput).Scan(&tip)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // kein Treffer
		}
		return "", err
	}
	return tip, nil
}

type TipEntry struct {
	TipID   uint64 `json:"tip_id"`
	Text    string `json:"text"`
	Type    string `json:"type"` // optional
	Created string `json:"created_at,omitempty"`
}

func GetTipsForUserAndTask(userID, taskID uint64) ([]TipEntry, error) {
	rows, err := DB.Query(`
		SELECT tt.id, tt.tip_text, tt.error_type, tt.created_at
		FROM user_tip_usage utu
		JOIN task_tips tt ON utu.tip_id = tt.id
		WHERE utu.user_id = ? AND utu.task_id = ?
		ORDER BY tt.created_at ASC
	`, userID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tips []TipEntry
	for rows.Next() {
		var tip TipEntry
		err := rows.Scan(&tip.TipID, &tip.Text, &tip.Type, &tip.Created)
		if err != nil {
			log.Println("❌ Scan-Fehler in GetTipsForUserAndTask:", err)
			continue
		}
		tips = append(tips, tip)
	}

	return tips, nil
}
