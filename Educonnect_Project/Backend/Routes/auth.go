package main

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ipt-9/EduConnect/DB"
	"log"
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

	log.Println("📥 Login-Versuch gestartet")

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Ungültiger Request Body", http.StatusBadRequest)
		log.Println("❌ Ungültiger JSON-Body im Login")
		return
	}

	if !DB.ValidateUser(creds.Email, creds.Password) {
		http.Error(w, "E-Mail oder Passwort ist falsch", http.StatusUnauthorized)
		log.Println("❌ Login fehlgeschlagen: E-Mail oder Passwort falsch")
		return
	}

	log.Printf("✅ Login erfolgreich für %s\n", creds.Email)

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Email: creds.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Fehler beim Erstellen des Tokens", http.StatusInternalServerError)
		log.Printf("❌ Fehler beim Signieren des Tokens für %s: %v\n", creds.Email, err)
		return
	}

	userID, err := DB.GetUserIDByEmail(creds.Email)
	if err != nil {
		log.Printf("⚠️ Benutzer-ID konnte nicht ermittelt werden: %v\n", err)
	}

	err = DB.StoreToken(userID, tokenString, time.Now(), expirationTime)
	if err != nil {
		log.Printf("⚠️ Fehler beim Speichern des Tokens: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
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
	err := DB.DeleteToken(tokenStr)
	if err != nil {
		http.Error(w, "Fehler beim Logout", http.StatusInternalServerError)
		log.Printf("❌ Fehler beim Token-Löschen: %v", err)
		return
	}

	log.Println("🚪 Benutzer ausgeloggt")
	w.Write([]byte("Logout erfolgreich"))
}
