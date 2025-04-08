package routes

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/ipt-9/EduConnect/DB"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func GetGroupNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 🔐 Bearer Token prüfen
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiger Token", http.StatusUnauthorized)
		return
	}

	// 👤 user_id aus Claims lesen
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		http.Error(w, "Fehlende user_id im Token", http.StatusUnauthorized)
		return
	}
	userID := uint64(userIDFloat)

	// 📦 groupID aus URL-Parametern lesen
	vars := mux.Vars(r)
	groupIDStr := vars["groupID"]
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültige Gruppen-ID", http.StatusBadRequest)
		return
	}

	// ✅ Ist User Mitglied der Gruppe?
	isMember, err := DB.IsUserInGroup(DB.DB, groupID, userID)
	if err != nil {
		log.Println("❌ DB-Fehler bei Gruppenmitgliedschaft:", err)
		http.Error(w, "Interner Serverfehler", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Nicht berechtigt – kein Gruppenmitglied", http.StatusForbidden)
		return
	}

	// 📬 Notifications aus DB holen
	notifications, err := DB.GetGroupNotifications(DB.DB, groupID)
	if err != nil {
		log.Println("❌ Fehler beim Laden der Notifications:", err)
		http.Error(w, "Fehler beim Laden der Benachrichtigungen", http.StatusInternalServerError)
		return
	}

	// ✅ Antwort senden
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}
