package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"PawTribalWars/db"
)

var jwtKey = []byte("super_secret_key")

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// ===== Walidacja =====

// silne hasło: min. 8 znaków, 1 wielka, 1 mała, 1 cyfra, 1 znak specjalny
func isPasswordStrong(password string) bool {
	if len(password) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

// prosty regex do sprawdzania formatu emaila
func isEmailValid(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// ===== Handlery =====

// RegisterHandler godoc
// @Summary Rejestracja użytkownika
// @Description Tworzy nowego użytkownika wraz ze startową wioską, zasobami, budynkami i jednostkami
// @Tags auth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Nazwa użytkownika"
// @Param email formData string true "Adres e-mail"
// @Param password formData string true "Hasło (min 8 znaków, wielka, mała, cyfra, znak specjalny)"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid input"
// @Failure 409 {string} string "User exists or DB error"
// @Router /register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if len(username) < 4 {
		http.Error(w, "Username too short", http.StatusBadRequest)
		return
	}
	if !isEmailValid(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	if !isPasswordStrong(password) {
		http.Error(w, "Password too weak. Must be at least 8 characters, with upper, lower, digit, and special character.", http.StatusBadRequest)
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// dodajemy użytkownika
	var userID int
	err := db.DB.QueryRow(
		"INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
		username, email, string(hashed),
	).Scan(&userID)
	if err != nil {
		http.Error(w, "User exists or DB error", http.StatusConflict)
		return
	}

	// startowa wioska
	var villageID int
	err = db.DB.QueryRow(
		"INSERT INTO villages (user_id, name) VALUES ($1, $2) RETURNING id",
		userID, "Startowa wioska",
	).Scan(&villageID)
	if err != nil {
		http.Error(w, "DB error on village init", http.StatusInternalServerError)
		return
	}

	// startowe surowce
	_, err = db.DB.Exec(
		"INSERT INTO resources (village_id, wood, clay, iron) VALUES ($1, 100, 100, 100)",
		villageID,
	)
	if err != nil {
		http.Error(w, "DB error on resources init", http.StatusInternalServerError)
		return
	}

	// startowe budynki lvl 1
	buildings := []string{"townhall", "lumbermill", "claypit", "ironmine", "warehouse", "barracks"}
	for _, b := range buildings {
		_, err = db.DB.Exec(
			"INSERT INTO buildings (village_id, type, level) VALUES ($1, $2, 1)",
			villageID, b,
		)
		if err != nil {
			http.Error(w, "DB error on buildings init", http.StatusInternalServerError)
			return
		}
	}

	// startowe jednostki
	_, err = db.DB.Exec("INSERT INTO units (village_id, type, count) VALUES ($1, $2, $3)",
		villageID, "spearman", 5,
	)
	if err != nil {
		http.Error(w, "DB error on units init", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered with starting village, resources, buildings and units",
	})
}

// LoginHandler godoc
// @Summary Logowanie użytkownika
// @Description Zwraca token JWT po poprawnym zalogowaniu
// @Tags auth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Nazwa użytkownika"
// @Param password formData string true "Hasło"
// @Success 200 {object} map[string]string
// @Failure 401 {string} string "Invalid credentials"
// @Failure 500 {string} string "DB error"
// @Router /login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	var id int
	var hashed string
	var role string
	err := db.DB.QueryRow("SELECT id, password_hash, role FROM users WHERE username=$1", username).
		Scan(&id, &hashed, &role)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expiration := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtKey)

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// LogoutHandler godoc
// @Summary Wylogowanie użytkownika
// @Description Nie unieważnia tokena – klient powinien go odrzucić po stronie frontendu
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /logout [post]
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "User logged out. Discard token client-side."})
}

// === JWT Middleware ===
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// dodajemy dane z tokena do contextu
		ctx := context.WithValue(r.Context(), "username", claims.Username)
		ctx = context.WithValue(ctx, "role", claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
