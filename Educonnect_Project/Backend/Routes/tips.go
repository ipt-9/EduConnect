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

func GetUserTipsForTaskHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
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
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiger Token", http.StatusUnauthorized)
		return
	}

	// 👤 user_id aus Token
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		http.Error(w, "Fehlende user_id im Token", http.StatusUnauthorized)
		return
	}
	userID := uint64(userIDFloat)

	// 📦 taskID aus URL lesen
	vars := mux.Vars(r)
	taskIDStr := vars["taskID"]
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültige taskID", http.StatusBadRequest)
		return
	}

	// 🧠 Tipps aus DB holen
	tips, err := DB.GetTipsForUserAndTask(userID, taskID)
	if err != nil {
		log.Println("❌ Fehler beim Laden der Tipps:", err)
		http.Error(w, "Fehler beim Laden der Tipps", http.StatusInternalServerError)
		return
	}

	// ✅ Antwort
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tips)
}
