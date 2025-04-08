package routes

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ipt-9/EduConnect/DB"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func GetMyCourses(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
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
	enableCORS(w)
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
	enableCORS(w)

	// Preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	// 🔐 Bearer Token prüfen
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

	// 🧠 User ID laden
	userID, err := DB.GetUserIDByEmail(claims.Email)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusInternalServerError)
		return
	}

	// 🧾 Body parsen
	var input DB.SubmissionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültiges JSON", http.StatusBadRequest)
		return
	}

	// ✅ Code darf nicht leer sein
	if strings.TrimSpace(input.Code) == "" {
		http.Error(w, "Code darf nicht leer sein", http.StatusBadRequest)
		return
	}

	input.UserID = userID

	// 💾 Lösung speichern + Fortschritt aktualisieren
	subID, err := DB.SaveSubmissionAndUpdateProgress(input)
	if err != nil {
		http.Error(w, "Fehler beim Speichern der Lösung", http.StatusInternalServerError)
		return
	}

	success := subID != 0

	// ✅ Benachrichtigung nur bei Erfolg
	if success {
		log.Println("✅ Aufgabe erfolgreich abgeschlossen – starte Notification-Logik")

		// Titel der Aufgabe laden
		taskTitle, err := DB.GetTaskTitleByID(DB.DB, input.TaskID)
		if err != nil {
			log.Println("⚠️ Konnte Aufgabentitel nicht laden:", err)
			taskTitle = "Unbekannte Aufgabe"
		}

		// Username laden
		username, err := DB.GetUsernameByID(DB.DB, userID)
		if err != nil {
			log.Println("⚠️ Konnte Username nicht laden:", err)
			username = "Ein Mitglied"
		}

		// Gruppen-IDs laden
		groupIDs, err := DB.GetGroupIDsForUser(DB.DB, userID)
		if err != nil {
			log.Println("⚠️ Konnte Gruppen nicht laden:", err)
		} else {
			for _, gid := range groupIDs {
				msg := fmt.Sprintf("✅ %s hat die Aufgabe „%s“ abgeschlossen.", username, taskTitle)
				err := DB.CreateGroupNotification(DB.DB, gid, &userID, "TASK_COMPLETED", msg)
				if err != nil {
					log.Println("❌ Fehler beim Speichern der Notification:", err)
				} else {
					log.Printf("🔔 Notification gespeichert für Gruppe %d\n", gid)
				}
			}
		}
	}

	// 🔁 Antwort senden
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"submission_id": subID,
		"success":       success,
	})
}

func GetSubmittedCode(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

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
