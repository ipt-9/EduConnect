package routes

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/ipt-9/EduConnect/DB"
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
