package routes

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ipt-9/EduConnect/DB"
)

type ChatMessage struct {
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	User      struct {
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
	enableCORS(w)

	// 1. Gruppe-ID aus URL holen
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

	// 2. Token aus Query holen und prüfen
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "Token fehlt", http.StatusUnauthorized)
		return
	}

	claims := &groupClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID

	// 3. Verbindung upgraden
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("❌ Upgrade fehlgeschlagen:", err)
		return
	}
	defer conn.Close()

	log.Printf("🔌 WS verbunden: User %d in Gruppe %d", userID, groupID)

	// 4. WebSocket Client speichern
	groupClientsMutex.Lock()
	if groupClients[groupID] == nil {
		groupClients[groupID] = make(map[*websocket.Conn]bool)
	}
	groupClients[groupID][conn] = true
	groupClientsMutex.Unlock()

	// 5. Alle bisherigen Nachrichten an den neuen Client schicken
	pastMessages, err := DB.GetFullGroupMessages(DB.DB, groupID, 1000000)
	if err != nil {
		log.Printf("❌ Fehler beim Laden alter Nachrichten: %v", err)
	} else {
		for i := len(pastMessages) - 1; i >= 0; i-- {
			if err := conn.WriteJSON(pastMessages[i]); err != nil {
				log.Printf("⚠️ Fehler beim Senden alter Nachricht: %v", err)
			}
		}
	}

	// 6. Nachrichtenschleife
	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("❌ WebSocket Lesefehler: %v", err)
			break
		}

		var incoming struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(rawMsg, &incoming); err != nil || strings.TrimSpace(incoming.Message) == "" {
			log.Println("⚠️ Leere oder ungültige Nachricht empfangen")
			continue
		}

		// Nachricht in DB speichern
		if err := DB.SaveGroupMessage(DB.DB, groupID, userID, incoming.Message); err != nil {
			log.Printf("❌ Nachricht konnte nicht gespeichert werden: %v", err)
			continue
		}

		// Benutzer laden (für Broadcast)
		user, err := DB.GetUserByID(DB.DB, userID)

		if err != nil {
			log.Printf("❌ Benutzer nicht gefunden: %v", err)
			continue
		}

		msg := DB.GroupChatMessage{
			Message:   incoming.Message,
			CreatedAt: time.Now(),
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

		// Broadcast an alle
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

	// 7. Verbindung aufräumen
	groupClientsMutex.Lock()
	delete(groupClients[groupID], conn)
	groupClientsMutex.Unlock()

	log.Printf("❎ WS getrennt: User %d aus Gruppe %d", userID, groupID)
}

func GetGroupMessagesHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	// JWT prüfen (wie gehabt)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Token fehlt oder ungültig", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &groupClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Token ungültig", http.StatusUnauthorized)
		return
	}

	// Gruppe-ID aus Pfad
	parts := strings.Split(r.URL.Path, "/")
	groupID, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Ungültige Gruppen-ID", http.StatusBadRequest)
		return
	}

	// DB-Call
	messages, err := DB.GetFullGroupMessages(DB.DB, groupID, 1000000)
	if err != nil {
		http.Error(w, "Fehler beim Laden: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
