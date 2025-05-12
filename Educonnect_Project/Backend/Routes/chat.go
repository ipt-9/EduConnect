package routes

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ipt-9/EduConnect/DB"
)

type GroupChatMessage struct {
	Message     string    `json:"message"`
	MessageType string    `json:"message_type"`
	TaskID      *uint64   `json:"task_id,omitempty"` // 🆕 HIER ergänzen!
	CreatedAt   time.Time `json:"created_at"`
	User        struct {
		ID                uint64  `json:"id"`
		Username          string  `json:"username"`
		Email             string  `json:"email"`
		ProfilePictureUrl *string `json:"profile_picture_url"`
	} `json:"user"`
}

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS erlauben
	},
}

var groupClients = make(map[uint64]map[*websocket.Conn]bool)
var groupClientsMutex sync.Mutex

func HandleGroupChatWS(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Ungültiger Pfad", http.StatusBadRequest)
		return
	}
	groupID, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		http.Error(w, "Ungültige Gruppen-ID", http.StatusBadRequest)
		return
	}

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "Token fehlt", http.StatusUnauthorized)
		return
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("❌ WebSocket Upgrade fehlgeschlagen:", err)
		return
	}
	defer conn.Close()
	log.Printf("🔌 WS verbunden: User %d in Gruppe %d", userID, groupID)

	groupClientsMutex.Lock()
	if groupClients[groupID] == nil {
		groupClients[groupID] = make(map[*websocket.Conn]bool)
	}
	groupClients[groupID][conn] = true
	groupClientsMutex.Unlock()

	pastMessages, err := DB.GetFullGroupMessages(claims.UserID, groupID, 1000)

	if err != nil {
		log.Printf("❌ Fehler beim Laden alter Nachrichten: %v", err)
	} else {
		for i := len(pastMessages) - 1; i >= 0; i-- {
			if err := conn.WriteJSON(pastMessages[i]); err != nil {
				log.Printf("⚠️ Fehler beim Senden alter Nachricht: %v", err)
			}
		}
	}

	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("❌ WebSocket Lesefehler: %v", err)
			break
		}

		var incoming struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		}
		if err := json.Unmarshal(rawMsg, &incoming); err != nil || strings.TrimSpace(incoming.Message) == "" {
			log.Println("⚠️ Ungültige oder leere Nachricht empfangen")
			continue
		}

		msgType := incoming.Type
		if msgType == "" {
			msgType = "text"
		}

		err = DB.SaveGroupMessage(groupID, userID, incoming.Message, msgType, nil)

		if err != nil {
			log.Printf("❌ Nachricht konnte nicht gespeichert werden: %v", err)
			continue
		}

		user, err := DB.GetUserByID(userID)
		if err != nil {
			log.Printf("❌ Benutzer konnte nicht geladen werden: %v", err)
			continue
		}

		msg := DB.GroupChatMessage{
			Message:     incoming.Message,
			MessageType: msgType,
			CreatedAt:   time.Now(),
			User: struct {
				ID                uint64  `json:"id"`
				Username          string  `json:"username"`
				Email             string  `json:"email"`
				ProfilePictureUrl *string `json:"profile_picture_url"`
			}{
				ID:                user.ID,
				Username:          user.Username,
				Email:             user.Email,
				ProfilePictureUrl: user.ProfilePictureUrl,
			},
		}

		groupClientsMutex.Lock()
		for client := range groupClients[groupID] {
			if err := client.WriteJSON(msg); err != nil {
				log.Println("⚠️ Broadcast-Fehler:", err)
				client.Close()
				delete(groupClients[groupID], client)
			}
		}
		groupClientsMutex.Unlock()
	}

	groupClientsMutex.Lock()
	delete(groupClients[groupID], conn)
	groupClientsMutex.Unlock()
	log.Printf("❎ WS getrennt: User %d aus Gruppe %d", userID, groupID)
}

func GetGroupMessagesHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)

	// 1️⃣ OPTIONS Preflight zuerst abfangen
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2️⃣ Authorization prüfen mit CORS
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		EnableCORS(w) // 💥 Wichtig: Header auch bei Fehler
		http.Error(w, "Token fehlt oder ungültig", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &groupClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		EnableCORS(w)
		http.Error(w, "Token ungültig", http.StatusUnauthorized)
		return
	}

	// 3️⃣ Gruppen-ID aus der URL extrahieren
	parts := strings.Split(r.URL.Path, "/")
	groupID, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		EnableCORS(w)
		http.Error(w, "Ungültige Gruppen-ID", http.StatusBadRequest)
		return
	}

	// 4️⃣ Datenbankabfrage
	messages, err := DB.GetFullGroupMessages(claims.UserID, groupID, 1000000)

	if err != nil {
		EnableCORS(w)
		http.Error(w, "Fehler beim Laden: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5️⃣ Erfolgreiche JSON-Antwort mit Header
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func ShareSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	log.Println("📥 Neue Anfrage auf /groups/{id}/share-submission")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Ungültige Methode", http.StatusMethodNotAllowed)
		log.Println("⛔ Methode nicht erlaubt:", r.Method)
		return
	}

	// 🔐 JWT prüfen
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		EnableCORS(w)
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		log.Println("⛔ Kein oder ungültiger Authorization Header")
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		EnableCORS(w)
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		log.Println("⛔ Token ungültig:", err)
		return
	}
	log.Printf("🔐 Authentifizierter User: %s (ID %d)", claims.Email, claims.UserID)

	// 🔢 Gruppen-ID aus Pfad
	vars := mux.Vars(r)
	groupIDStr := vars["id"]
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		EnableCORS(w)
		http.Error(w, "Ungültige Gruppen-ID", http.StatusBadRequest)
		log.Println("⛔ Fehler beim Parsen der Gruppen-ID:", err)
		return
	}

	// 📥 JSON-Body einlesen
	var req struct {
		TaskID int `json:"task_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		EnableCORS(w)
		http.Error(w, "Ungültiges JSON", http.StatusBadRequest)
		log.Println("⛔ Fehler beim Parsen des JSON:", err)
		return
	}

	// ✅ Submission, Task-Titel, Username laden
	sub, err := DB.GetSubmissionByTaskAndUser(uint64(req.TaskID), claims.UserID)
	if err != nil {
		EnableCORS(w)
		http.Error(w, "Keine gültige Lösung gefunden", http.StatusNotFound)
		log.Println("⛔ Keine gültige Submission:", err)
		return
	}
	title, _ := DB.GetTaskTitleByID(uint64(req.TaskID))
	username, _ := DB.GetUsernameByID(claims.UserID)
	user, _ := DB.GetUserByID(claims.UserID)

	// 💬 Nachricht formatieren
	msg := fmt.Sprintf(
		"✅ %s hat die Aufgabe „%s“ gelöst:\n```python\n%s\n```\n🕒 %dms\n📤 %s",
		username, title, sub.Code, sub.ExecutionTime, sub.Output,
	)

	// 💾 Nachricht speichern
	taskID := uint64(req.TaskID)
	err = DB.SaveGroupMessage(uint64(groupID), claims.UserID, msg, "submission", &taskID)

	if err != nil {
		EnableCORS(w)
		http.Error(w, "Nachricht konnte nicht gespeichert werden", http.StatusInternalServerError)
		log.Println("⛔ Fehler beim Speichern der Nachricht:", err)
		return
	}

	// 📢 Nachricht direkt über WebSocket an alle Gruppenmitglieder senden
	broadcast := DB.GroupChatMessage{
		Message:      msg,
		MessageType:  "submission",
		LinkedTaskID: &taskID, // ✅ jetzt richtig!
		CreatedAt:    time.Now(),
		User: struct {
			ID                uint64  `json:"id"`
			Username          string  `json:"username"`
			Email             string  `json:"email"`
			ProfilePictureUrl *string `json:"profile_picture_url"`
		}{
			ID:                user.ID,
			Username:          user.Username,
			Email:             user.Email,
			ProfilePictureUrl: user.ProfilePictureUrl,
		},
	}

	groupClientsMutex.Lock()
	for client := range groupClients[uint64(groupID)] {
		if err := client.WriteJSON(broadcast); err != nil {
			log.Printf("⚠️ WS Fehler beim Broadcast der Submission: %v", err)
			client.Close()
			delete(groupClients[uint64(groupID)], client)
		}
	}
	groupClientsMutex.Unlock()

	log.Println("✅ Nachricht erfolgreich gespeichert und gesendet")

	// ✅ Erfolgreiche JSON-Antwort
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "success",
		"message":        "Aufgabe erfolgreich geteilt",
		"shared_message": msg,
	})
}

func GetMySubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)

	// CORS Preflight korrekt abfangen
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		EnableCORS(w) // ❗️auch hier
		http.Error(w, "Token fehlt oder ungültig", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		EnableCORS(w) // ❗️auch hier
		http.Error(w, "Token ungültig", http.StatusUnauthorized)
		return
	}

	submissions, err := DB.GetSuccessfulSubmissionsByUser(claims.UserID)
	if err != nil {
		EnableCORS(w) // ❗️auch hier bei Fehler
		http.Error(w, "Fehler beim Abrufen der Daten: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}
