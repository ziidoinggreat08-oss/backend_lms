package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"backend_lms/config"
	"backend_lms/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Custom Claims untuk token JWT
type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func getJWTKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "edu_class_secret_key_123" // fallback default
	}
	return []byte(secret)
}

func Login(w http.ResponseWriter, r *http.Request) {
	// CORS Headers - Pastikan bisa dipanggil dari React Vite (localhost:5173)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Preflight request (OPTIONS)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Payload JSON tidak valid", http.StatusBadRequest)
		return
	}

	// Cari user di database berdasarkan email
	var user models.User
	var hashedPassword string

	query := "SELECT id, username, email, password, role FROM users WHERE email = ?"
	err = config.DB.QueryRow(query, req.Email).Scan(&user.ID, &user.Username, &user.Email, &hashedPassword, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Email tidak terdaftar", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Kesalahan server", http.StatusInternalServerError)
		return
	}

	// Validasi password menggunakan bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		http.Error(w, "Password salah", http.StatusUnauthorized)
		return
	}

	// Password valid! Generate Token JWT
	expirationTime := time.Now().Add(24 * time.Hour) // Token aktif 1 hari
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(getJWTKey())
	if err != nil {
		http.Error(w, "Gagal membuat token", http.StatusInternalServerError)
		return
	}

	// Format response yang dikirim kembali ke React
	res := models.LoginResponse{
		Message: "Login berhasil",
		Token:   tokenString,
		Role:    user.Role,
		User:    user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
