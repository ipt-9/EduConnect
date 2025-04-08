package routes

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/ipt-9/EduConnect/DB"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type groupClaims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type createGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1. Bearer Token extrahieren
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		log.Println("🚫 Kein Authorization Header vorhanden oder falsch formatiert")
		http.Error(w, "Authorization Header fehlt oder ist ungültig", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &groupClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		log.Printf("🚫 Ungültiger Token: %v\n", err)
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		return
	}

	// Debug: JWT Claims ausgeben
	log.Printf("📦 JWT Claims: user_id=%d | email=%s\n", claims.UserID, claims.Email)

	// 2. Body auslesen
	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("🚫 Ungültiger JSON-Body:", err)
		http.Error(w, "Ungültiger JSON-Body", http.StatusBadRequest)
		return
	}

	// Debug: Gruppendaten
	log.Printf("📨 Neue Gruppe: Name=%s | Beschreibung=%s\n", req.Name, req.Description)

	// 3. Gruppe anlegen
	log.Printf("🛠️  Starte Erstellen der Gruppe durch User-ID: %d\n", claims.UserID)
	group, err := DB.CreateGroup(DB.DB, req.Name, req.Description, claims.UserID)
	if err != nil {
		log.Printf("🔥 Fehler beim Erstellen der Gruppe: %v\n", err)
		http.Error(w, "Fehler beim Erstellen der Gruppe: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Erfolg zurückgeben
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}
func JoinGroupByCodeHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Ungültige Methode", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt oder ist ungültig", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &groupClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Einladungscode fehlt", http.StatusBadRequest)
		return
	}

	err = DB.JoinGroupByCode(DB.DB, code, claims.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Beitritt fehlgeschlagen: %v", err), http.StatusBadRequest)
		return
	}

	// 🔔 Notification hinzufügen
	groupID, _ := DB.GetGroupIDByInviteCode(DB.DB, code)
	username, _ := DB.GetUsernameByID(DB.DB, claims.UserID)
	msg := fmt.Sprintf("👥 %s ist der Gruppe beigetreten.", username)
	_ = DB.CreateGroupNotification(DB.DB, int64(groupID), &claims.UserID, "GROUP_EVENT", msg)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("🎉 Du bist der Gruppe erfolgreich beigetreten!"))
}

func GetGroupMembersHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 🔐 Token prüfen
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &groupClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		return
	}

	// 🔢 Gruppen-ID aus der URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Ungültige Gruppen-ID", http.StatusBadRequest)
		return
	}
	groupID, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Ungültige Gruppen-ID", http.StatusBadRequest)
		return
	}

	// 📥 Mitglieder abrufen
	members, err := DB.GetGroupMembers(DB.DB, groupID)
	if err != nil {
		http.Error(w, "Fehler beim Laden der Gruppenmitglieder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}
func RemoveGroupMemberHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, "Ungültige Methode", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &groupClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Ungültiger Pfad", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Ungültige Gruppen-ID", http.StatusBadRequest)
		return
	}
	targetUserID, err := strconv.ParseUint(parts[4], 10, 64)
	if err != nil {
		http.Error(w, "Ungültige User-ID", http.StatusBadRequest)
		return
	}

	log.Printf("🔍 claims.UserID: %d, targetUserID: %d, groupID: %d", claims.UserID, targetUserID, groupID)

	// Nur wenn sich jemand selbst entfernen will und admin ist
	if claims.UserID == targetUserID {
		isAdmin, err := DB.IsUserAdminInGroup(DB.DB, groupID, targetUserID)
		if err != nil {
			log.Printf("❌ Fehler bei IsUserAdminInGroup: %v", err)
			http.Error(w, "Fehler beim Admin-Check: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("👤 User %d ist Admin in Gruppe %d: %v", targetUserID, groupID, isAdmin)

		if isAdmin {
			adminCount, err := DB.CountAdminsInGroup(DB.DB, groupID)
			if err != nil {
				log.Printf("❌ Fehler bei CountAdminsInGroup: %v", err)
				http.Error(w, "Fehler beim Zählen der Admins: "+err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("👑 Adminanzahl in Gruppe %d: %d", groupID, adminCount)

			if adminCount <= 1 {
				log.Printf("⛔ Letzter Admin (%d) darf sich nicht selbst entfernen", targetUserID)
				http.Error(w, "Fehler beim Entfernen: Admins können sich nicht selbst entfernen", http.StatusForbidden)
				return
			}
		}
	}

	err = DB.RemoveGroupMember(DB.DB, groupID, targetUserID, claims.UserID)
	if err != nil {
		log.Printf("❌ Fehler bei RemoveGroupMember: %v", err)
		http.Error(w, fmt.Sprintf("Fehler beim Entfernen: %v", err), http.StatusForbidden)
		return
	}

	username, _ := DB.GetUsernameByID(DB.DB, uint64(targetUserID))
	msg := fmt.Sprintf("🚪 %s hat die Gruppe verlassen.", username)
	_ = DB.CreateGroupNotification(DB.DB, int64(groupID), &claims.UserID, "GROUP_EVENT", msg)

	log.Printf("✅ %s (ID %d) wurde erfolgreich aus Gruppe %d entfernt", username, targetUserID, groupID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "✅ Mitglied erfolgreich entfernt",
	})
}

func UpdateMemberRoleHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPut {
		http.Error(w, "Ungültige Methode", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &groupClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		http.Error(w, "Ungültiger Pfad", http.StatusBadRequest)
		return
	}
	groupID, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Ungültige Gruppen-ID", http.StatusBadRequest)
		return
	}
	targetUserID, err := strconv.ParseUint(parts[4], 10, 64)
	if err != nil {
		http.Error(w, "Ungültige User-ID", http.StatusBadRequest)
		return
	}

	type roleRequest struct {
		Role string `json:"role"`
	}
	var req roleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ungültiger Body", http.StatusBadRequest)
		return
	}
	if req.Role != "admin" && req.Role != "member" {
		http.Error(w, "Ungültige Rolle. Erlaubt: admin, member", http.StatusBadRequest)
		return
	}

	err = DB.UpdateMemberRole(DB.DB, groupID, targetUserID, claims.UserID, req.Role)
	if err != nil {
		http.Error(w, "Fehler beim Aktualisieren der Rolle: "+err.Error(), http.StatusForbidden)
		return
	}

	// 🔔 Notification hinzufügen
	username, _ := DB.GetUsernameByID(DB.DB, targetUserID)
	msg := fmt.Sprintf("🔒 %s wurde zum Gruppen-%s gemacht.", username, req.Role)
	_ = DB.CreateGroupNotification(DB.DB, int64(groupID), &claims.UserID, "GROUP_EVENT", msg)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "✅ Rolle erfolgreich aktualisiert",
	})

}

func GetUserGroupsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 🔐 Token prüfen
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &groupClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		return
	}

	// 📥 Gruppen aus DB laden
	groups, err := DB.GetGroupsForUser(DB.DB, claims.UserID)
	if err != nil {
		http.Error(w, "Fehler beim Laden der Gruppen: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ JSON zurückgeben
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}
func GetGroupByIDHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	groupID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültige Gruppen-ID", http.StatusBadRequest)
		return
	}

	// 🔐 JWT prüfen
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &groupClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Token ungültig", http.StatusUnauthorized)
		return
	}
	// 📦 Datenbank aufrufen
	group, err := DB.GetGroupByID(DB.DB, groupID)
	if err != nil {
		http.Error(w, "Gruppe nicht gefunden: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// levin & tomas were here
