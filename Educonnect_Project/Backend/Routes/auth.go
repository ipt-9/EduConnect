package routes

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ipt-9/EduConnect/2fa"
	"github.com/ipt-9/EduConnect/DB"
	"net/http"
	"os"
	"strings"
	"time"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Fehler beim Parsen des Formulars", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if len(username) < 3 || email == "" || len(password) < 8 {
		http.Error(w, "Ungültige Eingabedaten", http.StatusBadRequest)
		return
	}

	DB.CreateUser(username, password, email)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}

func Login(w http.ResponseWriter, r *http.Request) {
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
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization Header fehlt oder ist ungültig", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	// Token löschen
	err := DB.DeleteToken(tokenStr)
	if err != nil {
		http.Error(w, "Fehler beim Logout", http.StatusInternalServerError)
		return
	}

	// Optional: E-Mail aus Token extrahieren → userID holen
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

	userID, err := DB.GetUserIDByEmail(data.Email)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusBadRequest)
		return
	}

	if !DB.Validate2FACode(userID, data.Code) {
		http.Error(w, "Ungültiger 2FA-Code", http.StatusUnauthorized)
		return
	}

	// ✅ Token erstellen
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Email: data.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Token-Fehler", http.StatusInternalServerError)
		return
	}

	// ✅ Token speichern
	err = DB.StoreToken(userID, tokenString, time.Now(), expirationTime)
	if err != nil {
		http.Error(w, "Fehler beim Speichern des Tokens", http.StatusInternalServerError)
		return
	}

	// ✅ 2FA-Code löschen
	DB.Delete2FACode(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
