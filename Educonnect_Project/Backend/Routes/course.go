package routes

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ipt-9/EduConnect/DB"
	"github.com/ipt-9/EduConnect/Tip"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func GetMyCourses(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges Token", http.StatusUnauthorized)
		return
	}

	userID, err := DB.GetUserIDByEmail(claims.Email)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusInternalServerError)
		return
	}

	courses, err := DB.GetCoursesForUser(userID)
	if err != nil {
		http.Error(w, "Fehler beim Laden der Kurse", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}
func GetTasksByCourse(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges Token", http.StatusUnauthorized)
		return
	}

	userID, err := DB.GetUserIDByEmail(claims.Email)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	courseIDStr := vars["id"]
	if courseIDStr == "" {
		http.Error(w, "Course ID fehlt", http.StatusBadRequest)
		return
	}

	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültige Course ID", http.StatusBadRequest)
		return
	}

	tasks, err := DB.GetTasksForCourse(courseID, userID)
	if err != nil {
		http.Error(w, "Fehler beim Laden der Aufgaben", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}
func SubmitTaskSolution(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges Token", http.StatusUnauthorized)
		return
	}

	userID, err := DB.GetUserIDByEmail(claims.Email)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusInternalServerError)
		return
	}

	var input DB.SubmissionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültiges JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(input.Code) == "" {
		http.Error(w, "Code darf nicht leer sein", http.StatusBadRequest)
		return
	}

	input.UserID = userID

	// 💾 Bewertung & Speichern
	subID, isCorrect, _, err := DB.SaveSubmissionAndUpdateProgress(input)
	if err != nil {
		http.Error(w, "Fehler beim Speichern der Lösung", http.StatusInternalServerError)
		return
	}

	log.Printf("📊 isCorrect: %v", isCorrect)
	log.Printf("📊 subID:     %d", subID)

	success := isCorrect

	var generatedTip string

	// 💡 Nur bei falscher Lösung: Tipp-Logik
	if !success {
		task, err := DB.GetTaskByID(input.TaskID)
		if err != nil {
			http.Error(w, "Aufgabe nicht gefunden", http.StatusNotFound)
			return
		}

		existingTip, err := DB.GetTipByTaskAndOutput(input.TaskID, input.Output)
		if err != nil {
			log.Println("⚠️ Fehler beim Nachschauen von vorhandenem Tipp:", err)
		}

		if existingTip != "" {
			generatedTip = existingTip
			go func() {
				if err := DB.SaveUserTipUsage(userID, input.TaskID, existingTip); err != nil {
					log.Println("❌ Fehler beim Speichern in user_tip_usage:", err)
				}
			}()
		} else {
			// 🪄 Gemini-Tipp generieren
			prompt := Tip.BuildGeminiPrompt(task, input.Code, task.ExpectedOutput, input.Output)
			generatedTip, err = Tip.FetchTipFromGemini(prompt)
			if err != nil {
				log.Println("❌ Fehler bei Tipp-Generierung:", err)
				generatedTip = "Leider konnte kein Tipp generiert werden."
			}

			errorType := Tip.DetectErrorType(input.Output, isCorrect)
			errorToken := Tip.ExtractErrorToken(input.Output)

			go func() {
				if err := DB.SaveGeneratedTip(input.TaskID, errorToken, generatedTip, errorType); err != nil {
					log.Println("❌ Fehler beim Speichern des Tipps:", err)
					return
				}
				if err := DB.SaveUserTipUsage(userID, input.TaskID, generatedTip); err != nil {
					log.Println("❌ Fehler beim Speichern in user_tip_usage:", err)
				}
			}()
		}
	}

	// ✅ Erfolgsfall: Gruppen-Notification
	if success {
		log.Println("✅ Aufgabe erfolgreich abgeschlossen – starte Notification-Logik")

		taskTitle, err := DB.GetTaskTitleByID(input.TaskID)
		if err != nil {
			log.Println("⚠️ Konnte Aufgabentitel nicht laden:", err)
			taskTitle = "Unbekannte Aufgabe"
		}

		username, err := DB.GetUsernameByID(userID)
		if err != nil {
			log.Println("⚠️ Konnte Username nicht laden:", err)
			username = "Ein Mitglied"
		}

		groupIDs, err := DB.GetGroupIDsForUser(userID)
		if err != nil {
			log.Println("⚠️ Konnte Gruppen nicht laden:", err)
		} else {
			for _, gid := range groupIDs {
				msg := fmt.Sprintf("✅ %s hat die Aufgabe „%s“ abgeschlossen.", username, taskTitle)
				err := DB.CreateGroupNotification(gid, &userID, "TASK_COMPLETED", msg)
				if err != nil {
					log.Println("❌ Fehler beim Speichern der Notification:", err)
				} else {
					log.Printf("🔔 Notification gespeichert für Gruppe %d\n", gid)
				}
			}
		}
	}

	// 📤 Antwort
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"submission_id": subID,
		"success":       success,
	}

	// Nur bei Fehler & vorhandenem Tipp
	if !success && generatedTip != "" {
		resp["tip"] = generatedTip
	}

	json.NewEncoder(w).Encode(resp)
}

func GetSubmittedCode(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges Token", http.StatusUnauthorized)
		return
	}

	userID, err := DB.GetUserIDByEmail(claims.Email)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	taskIDStr := vars["task_id"]
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültige Task-ID", http.StatusBadRequest)
		return
	}

	// 🔁 Holt Code aus Submission (bei Erfolg) oder user_task_progress (bei Fehlschlag)
	code, err := DB.GetSubmittedOrAttemptedCode(userID, taskID)
	if err != nil || code == "" {
		http.Error(w, "Kein Code gefunden", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"code": code,
	})
}
