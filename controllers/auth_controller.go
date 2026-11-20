package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"backend_lms/config"
	"backend_lms/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func getJWTKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "edu_class_secret_key_123"
	}
	return []byte(secret)
}

func sendAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"message": message,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		sendAuthError(w, http.StatusMethodNotAllowed, "Metode tidak diizinkan")
		return
	}

	var req models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendAuthError(w, http.StatusBadRequest, "Payload JSON tidak valid")
		return
	}

	identifier := strings.TrimSpace(req.Email)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Username)
	}

	if identifier == "" || strings.TrimSpace(req.Password) == "" {
		sendAuthError(w, http.StatusBadRequest, "Email, Username, atau NIS dan Password wajib diisi")
		return
	}

	// Cari user di database berdasarkan email, username, ATAU NIS siswa
	var user models.User
	var hashedPassword string

	query := `
		SELECT u.id, u.username, u.email, u.password, u.role 
		FROM users u 
		LEFT JOIN siswa s ON s.user_id = u.id 
		WHERE LOWER(u.email) = LOWER(?) OR LOWER(u.username) = LOWER(?) OR s.nis = ?
		LIMIT 1
	`
	err = config.DB.QueryRow(query, identifier, identifier, identifier).Scan(&user.ID, &user.Username, &user.Email, &hashedPassword, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			sendAuthError(w, http.StatusUnauthorized, "Akun tidak terdaftar. Pastikan Email, Username, atau NIS sudah benar.")
			return
		}
		if strings.Contains(err.Error(), "127.0.0.1:3306") || strings.Contains(err.Error(), "actively refused") || strings.Contains(err.Error(), "connectex") {
			sendAuthError(w, http.StatusInternalServerError, "Database MySQL belum berjalan di Laragon. Buka Laragon lalu klik 'Start All'.")
			return
		}
		sendAuthError(w, http.StatusInternalServerError, "Kesalahan server: "+err.Error())
		return
	}

	// Validasi password menggunakan bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		sendAuthError(w, http.StatusUnauthorized, "Password salah. Silakan periksa kembali password Anda.")
		return
	}

	// Jika role siswa, pastikan status siswa aktif
	if user.Role == "siswa" {
		var statusSiswa string
		err = config.DB.QueryRow("SELECT COALESCE(status, 'aktif') FROM siswa WHERE user_id = ?", user.ID).Scan(&statusSiswa)
		if err == nil && statusSiswa != "" && statusSiswa != "aktif" {
			sendAuthError(w, http.StatusForbidden, "Akun siswa Anda berstatus non-aktif ("+statusSiswa+"). Silakan hubungi admin sekolah.")
			return
		}
	}

	// Hitung masa aktif token dari ENV
	expireHours := 24
	if envHours := os.Getenv("JWT_EXPIRE_HOURS"); envHours != "" {
		if h, err := strconv.Atoi(envHours); err == nil && h > 0 {
			expireHours = h
		}
	}

	expirationTime := time.Now().Add(time.Duration(expireHours) * time.Hour)
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
		sendAuthError(w, http.StatusInternalServerError, "Gagal membuat token autentikasi")
		return
	}

	// Sembunyikan hash password dari response
	user.Password = ""

	resp := models.LoginResponse{
		Message: "Login berhasil",
		Token:   tokenString,
		Role:    user.Role,
		User:    user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
