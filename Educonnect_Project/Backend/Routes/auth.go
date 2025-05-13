package routes

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ipt-9/EduConnect/2fa"
	"github.com/ipt-9/EduConnect/DB"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var jwtKey []byte

func InitJWT() {
	jwtKey = []byte(os.Getenv("JWT_SECRET"))
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	Email  string `json:"email"`
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

func Register(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Fehler beim Parsen des Formulars", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if len(username) < 3 || email == "" || len(password) < 8 {
		http.Error(w, "Ungültige Eingabedaten", http.StatusBadRequest)
		return
	}

	optional := func(val string) *string {
		val = strings.TrimSpace(val)
		if val == "" {
			return nil
		}
		return &val
	}

	legalName := optional(r.FormValue("legal_name"))
	phoneNumber := optional(r.FormValue("phone_number"))
	address := optional(r.FormValue("address"))
	profilePictureUrl := optional(r.FormValue("profile_picture_url"))

	err := DB.CreateUser(username, email, password, legalName, phoneNumber, address, profilePictureUrl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Fehler bei der Registrierung: %v", err), http.StatusInternalServerError)
		return
	}

	// 🔐 Benutzer-ID abrufen
	userID, err := DB.GetUserIDByEmail(email)
	if err != nil {
		log.Println("❌ GetUserIDByEmail-Fehler:", err)
		http.Error(w, "Benutzer wurde erstellt, aber konnte nicht gefunden werden", http.StatusInternalServerError)
		return
	}

	log.Printf("🧠 Benutzer-ID gefunden: %d", userID)

	// 🧠 Standardkurs
	if err := DB.AssignAllCoursesToUser(userID); err != nil {
		http.Error(w, "Fehler beim automatischen Zuweisen des Kurses", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("✅ User registered successfully (Basic Python hinzugefügt)"))
}

func Login(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Ungültiger Request Body", http.StatusBadRequest)
		return
	}

	if !DB.ValidateUser(creds.Email, creds.Password) {
		http.Error(w, "E-Mail oder Passwort ist falsch", http.StatusUnauthorized)
		return
	}

	userID, err := DB.GetUserIDByEmail(creds.Email)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusInternalServerError)
		return
	}

	err = twofa.Send2FACode(userID, creds.Email)
	if err != nil {
		http.Error(w, "Fehler beim Senden des 2FA-Codes", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("📧 2FA-Code wurde an deine E-Mail gesendet. Bitte verifizieren unter /verify-2fa."))
}

func Protected(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt oder ist ungültig", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("🔒 Zugriff gewährt: Willkommen " + claims.Email))
}

func Logout(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt oder ist ungültig", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	err := DB.DeleteToken(tokenStr)
	if err != nil {
		http.Error(w, "Fehler beim Logout", http.StatusInternalServerError)
		return
	}

	claims := &Claims{}
	_, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err == nil {
		userID, err := DB.GetUserIDByEmail(claims.Email)
		if err == nil {
			DB.Delete2FACode(userID)
		}
	}

	w.Write([]byte("🚪 Erfolgreich ausgeloggt"))
}

func Verify2FA(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	type reqData struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	var data reqData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Ungültiger Body", http.StatusBadRequest)
		return
	}

	log.Println("✅ 2FA verifiziert – JWT wird erstellt für:", data.Email)

	userID, err := DB.GetUserIDByEmail(data.Email)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusBadRequest)
		return
	}

	if !DB.Validate2FACode(userID, data.Code) {
		http.Error(w, "Ungültiger 2FA-Code", http.StatusUnauthorized)
		return
	}

	// 🔄 Username laden
	username, err := DB.GetUsernameByUserID(userID)
	if err != nil {
		http.Error(w, "Username konnte nicht geladen werden", http.StatusInternalServerError)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"user_id":  userID,
		"email":    data.Email,
		"username": username, // 👈 Hier wird's wichtig!
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Token-Fehler", http.StatusInternalServerError)
		return
	}

	err = DB.StoreToken(userID, tokenString, time.Now(), expirationTime)
	if err != nil {
		http.Error(w, "Fehler beim Speichern des Tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func Me(w http.ResponseWriter, r *http.Request) {

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
		http.Error(w, "Authorization Header fehlt oder ist ungültig", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Ungültiges oder abgelaufenes Token", http.StatusUnauthorized)
		return
	}

	log.Println("🔍 Token erhalten:", tokenStr)
	log.Println("📧 Email aus Token:", claims.Email)

	if claims.Email == "" {
		http.Error(w, "Token enthält keine E-Mail", http.StatusUnauthorized)
		return
	}

	user, err := DB.GetUserByEmail(claims.Email)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
