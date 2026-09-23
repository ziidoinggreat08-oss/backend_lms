package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend_lms/config"
	"backend_lms/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const siswaSelect = `SELECT s.id, s.user_id, s.nis, COALESCE(s.nisn, ''), s.nama_lengkap,
	u.username, u.email, s.jenis_kelamin, COALESCE(s.tempat_lahir, ''),
	COALESCE(DATE_FORMAT(s.tanggal_lahir, '%Y-%m-%d'), ''), COALESCE(s.no_telepon, ''),
	COALESCE(s.alamat, ''), s.kelas_id, k.nama_kelas, s.jurusan_id, j.nama_jurusan, s.status
	FROM siswa s JOIN users u ON u.id = s.user_id
	JOIN kelas k ON k.id = s.kelas_id JOIN jurusan j ON j.id = s.jurusan_id`

func setJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func writeError(w http.ResponseWriter, status int, message string) {
	setJSON(w)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if auth == "" {
		writeError(w, http.StatusUnauthorized, "Token login diperlukan")
		return false
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode token tidak valid")
		}
		return getJWTKey(), nil
	})
	if err != nil || !token.Valid || claims.Role != "admin" {
		writeError(w, http.StatusForbidden, "Akses hanya untuk admin")
		return false
	}
	return true
}

func decodeSiswaRequest(w http.ResponseWriter, r *http.Request, creating bool) (models.SiswaRequest, bool) {
	var req models.SiswaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Data siswa tidak valid")
		return req, false
	}
	req.NIS, req.NamaLengkap, req.Username, req.Email = strings.TrimSpace(req.NIS), strings.TrimSpace(req.NamaLengkap), strings.TrimSpace(req.Username), strings.TrimSpace(req.Email)
	if req.NIS == "" || req.NamaLengkap == "" || req.Username == "" || req.Email == "" || req.KelasID < 1 || req.JurusanID < 1 || req.JenisKelamin == "" {
		writeError(w, http.StatusBadRequest, "Lengkapi data wajib siswa")
		return req, false
	}
	if creating && len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "Password minimal 6 karakter")
		return req, false
	}
	if req.Status == "" {
		req.Status = "aktif"
	}
	if req.Status != "aktif" && req.Status != "nonaktif" && req.Status != "lulus" && req.Status != "keluar" {
		writeError(w, http.StatusBadRequest, "Status siswa tidak valid")
		return req, false
	}
	return req, true
}

func kelasSesuaiJurusan(kelasID, jurusanID int) bool {
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM kelas WHERE id = ? AND jurusan_id = ?", kelasID, jurusanID).Scan(&count)
	return err == nil && count == 1
}

func scanSiswa(scanner interface{ Scan(...any) error }) (models.Siswa, error) {
	var s models.Siswa
	err := scanner.Scan(&s.ID, &s.UserID, &s.NIS, &s.NISN, &s.NamaLengkap, &s.Username, &s.Email, &s.JenisKelamin, &s.TempatLahir, &s.TanggalLahir, &s.NoTelepon, &s.Alamat, &s.KelasID, &s.NamaKelas, &s.JurusanID, &s.NamaJurusan, &s.Status)
	return s, err
}

func ListSiswa(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "GET" {
		writeError(w, 405, "Metode tidak diizinkan")
		return
	}
	if !requireAdmin(w, r) {
		return
	}
	rows, err := config.DB.Query(siswaSelect + " ORDER BY s.nama_lengkap")
	if err != nil {
		writeError(w, 500, "Gagal mengambil data siswa")
		return
	}
	defer rows.Close()
	siswa := make([]models.Siswa, 0)
	for rows.Next() {
		item, err := scanSiswa(rows)
		if err != nil {
			writeError(w, 500, "Gagal membaca data siswa")
			return
		}
		siswa = append(siswa, item)
	}
	json.NewEncoder(w).Encode(siswa)
}

func ListKelas(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "GET" {
		writeError(w, 405, "Metode tidak diizinkan")
		return
	}
	if !requireAdmin(w, r) {
		return
	}
	rows, err := config.DB.Query(`SELECT k.id, k.nama_kelas, j.id, j.nama_jurusan FROM kelas k JOIN jurusan j ON j.id = k.jurusan_id ORDER BY k.nama_kelas`)
	if err != nil {
		writeError(w, 500, "Gagal mengambil data kelas")
		return
	}
	defer rows.Close()
	kelas := make([]models.KelasOption, 0)
	for rows.Next() {
		var item models.KelasOption
		if err := rows.Scan(&item.ID, &item.NamaKelas, &item.JurusanID, &item.NamaJurusan); err != nil {
			writeError(w, 500, "Gagal membaca data kelas")
			return
		}
		kelas = append(kelas, item)
	}
	json.NewEncoder(w).Encode(kelas)
}

func CreateSiswa(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		writeError(w, 405, "Metode tidak diizinkan")
		return
	}
	if !requireAdmin(w, r) {
		return
	}
	req, ok := decodeSiswaRequest(w, r, true)
	if !ok {
		return
	}
	if !kelasSesuaiJurusan(req.KelasID, req.JurusanID) {
		writeError(w, 400, "Kelas harus sesuai dengan jurusan yang dipilih")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, 500, "Gagal mengamankan password")
		return
	}
	tx, err := config.DB.Begin()
	if err != nil {
		writeError(w, 500, "Gagal memulai penyimpanan")
		return
	}
	defer tx.Rollback()
	result, err := tx.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, 'siswa')", req.Username, req.Email, string(hash))
	if err != nil {
		writeError(w, 409, "Username atau email sudah digunakan")
		return
	}
	userID, _ := result.LastInsertId()
	result, err = tx.Exec(`INSERT INTO siswa (user_id, nis, nisn, nama_lengkap, jenis_kelamin, tempat_lahir, tanggal_lahir, no_telepon, alamat, kelas_id, jurusan_id, status) VALUES (?, ?, NULLIF(?, ''), ?, ?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?)`, userID, req.NIS, req.NISN, req.NamaLengkap, req.JenisKelamin, req.TempatLahir, req.TanggalLahir, req.NoTelepon, req.Alamat, req.KelasID, req.JurusanID, req.Status)
	if err != nil {
		writeError(w, 409, "NIS atau NISN sudah digunakan")
		return
	}
	studentID, _ := result.LastInsertId()
	if err = tx.Commit(); err != nil {
		writeError(w, 500, "Gagal menyimpan siswa")
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"message": "Siswa berhasil ditambahkan", "id": studentID})
}

func StudentByID(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method == "OPTIONS" {
		return
	}
	if !requireAdmin(w, r) {
		return
	}
	idText := strings.TrimPrefix(r.URL.Path, "/api/siswa/")
	id, err := strconv.Atoi(idText)
	if err != nil || id < 1 {
		writeError(w, 400, "ID siswa tidak valid")
		return
	}
	switch r.Method {
	case "GET":
		getSiswa(w, id)
	case "PUT":
		updateSiswa(w, r, id)
	case "DELETE":
		deleteSiswa(w, id)
	default:
		writeError(w, 405, "Metode tidak diizinkan")
	}
}

func getSiswa(w http.ResponseWriter, id int) {
	s, err := scanSiswa(config.DB.QueryRow(siswaSelect+" WHERE s.id = ?", id))
	if err == sql.ErrNoRows {
		writeError(w, 404, "Siswa tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, 500, "Gagal mengambil siswa")
		return
	}
	json.NewEncoder(w).Encode(s)
}

func updateSiswa(w http.ResponseWriter, r *http.Request, id int) {
	req, ok := decodeSiswaRequest(w, r, false)
	if !ok {
		return
	}
	if !kelasSesuaiJurusan(req.KelasID, req.JurusanID) {
		writeError(w, 400, "Kelas harus sesuai dengan jurusan yang dipilih")
		return
	}
	var userID int
	if err := config.DB.QueryRow("SELECT user_id FROM siswa WHERE id = ?", id).Scan(&userID); err == sql.ErrNoRows {
		writeError(w, 404, "Siswa tidak ditemukan")
		return
	} else if err != nil {
		writeError(w, 500, "Gagal memperbarui siswa")
		return
	}
	tx, err := config.DB.Begin()
	if err != nil {
		writeError(w, 500, "Gagal memulai pembaruan")
		return
	}
	defer tx.Rollback()
	if _, err = tx.Exec("UPDATE users SET username = ?, email = ? WHERE id = ?", req.Username, req.Email, userID); err != nil {
		writeError(w, 409, "Username atau email sudah digunakan")
		return
	}
	if strings.TrimSpace(req.Password) != "" {
		if len(req.Password) < 6 {
			writeError(w, 400, "Password minimal 6 karakter")
			return
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if _, err = tx.Exec("UPDATE users SET password = ? WHERE id = ?", string(hash), userID); err != nil {
			writeError(w, 500, "Gagal memperbarui password")
			return
		}
	}
	_, err = tx.Exec(`UPDATE siswa SET nis=?, nisn=NULLIF(?, ''), nama_lengkap=?, jenis_kelamin=?, tempat_lahir=NULLIF(?, ''), tanggal_lahir=NULLIF(?, ''), no_telepon=NULLIF(?, ''), alamat=NULLIF(?, ''), kelas_id=?, jurusan_id=?, status=? WHERE id=?`, req.NIS, req.NISN, req.NamaLengkap, req.JenisKelamin, req.TempatLahir, req.TanggalLahir, req.NoTelepon, req.Alamat, req.KelasID, req.JurusanID, req.Status, id)
	if err != nil {
		writeError(w, 409, "NIS atau NISN sudah digunakan")
		return
	}
	if err = tx.Commit(); err != nil {
		writeError(w, 500, "Gagal memperbarui siswa")
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Data siswa berhasil diperbarui"})
}

func deleteSiswa(w http.ResponseWriter, id int) {
	result, err := config.DB.Exec("DELETE u FROM users u JOIN siswa s ON s.user_id = u.id WHERE s.id = ?", id)
	if err != nil {
		writeError(w, 500, "Gagal menghapus siswa")
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		writeError(w, 404, "Siswa tidak ditemukan")
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Siswa berhasil dihapus"})
}
