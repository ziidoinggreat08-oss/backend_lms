package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend_lms/config"
)

type Guru struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	NIP          string `json:"nip"`
	NamaLengkap  string `json:"nama_lengkap"`
	JenisKelamin string `json:"jenis_kelamin"`
	NoTelepon    string `json:"no_telepon"`
	Alamat       string `json:"alamat"`
	Status       string `json:"status"`
	// dari tabel users
	Email    string `json:"email"`
	Username string `json:"username"`
}

func HandleGuru(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := ""
	if len(parts) > 2 {
		id = parts[2]
	}

	switch r.Method {
	case "GET":
		if id == "" {
			getGurus(w, r)
		}
	case "POST":
		createGuru(w, r)
	case "PUT":
		updateGuru(w, r, id)
	case "DELETE":
		deleteGuru(w, r, id)
	}
}

func getGurus(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(`
		SELECT g.id, g.user_id, COALESCE(g.nip,''), g.nama_lengkap, g.jenis_kelamin,
		       COALESCE(g.no_telepon,''), COALESCE(g.alamat,''), COALESCE(g.status,'aktif'),
		       COALESCE(u.email,''), COALESCE(u.username,'')
		FROM guru g
		LEFT JOIN users u ON u.id = g.user_id
		ORDER BY g.nama_lengkap
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []Guru{}
	for rows.Next() {
		var item Guru
		if err := rows.Scan(&item.ID, &item.UserID, &item.NIP, &item.NamaLengkap, &item.JenisKelamin,
			&item.NoTelepon, &item.Alamat, &item.Status, &item.Email, &item.Username); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

type GuruInput struct {
	NIP          string `json:"nip"`
	NamaLengkap  string `json:"nama_lengkap"`
	JenisKelamin string `json:"jenis_kelamin"`
	NoTelepon    string `json:"no_telepon"`
	Alamat       string `json:"alamat"`
	Status       string `json:"status"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func createGuru(w http.ResponseWriter, r *http.Request) {
	var item GuruInput
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if item.Status == "" {
		item.Status = "aktif"
	}

	// Buat user account terlebih dahulu
	res, err := config.DB.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, 'dummy_password', 'guru')",
		item.Username, item.Email)
	if err != nil {
		http.Error(w, "Gagal membuat akun: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userID, _ := res.LastInsertId()

	_, err = config.DB.Exec("INSERT INTO guru (user_id, nip, nama_lengkap, jenis_kelamin, no_telepon, alamat, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, item.NIP, item.NamaLengkap, item.JenisKelamin, item.NoTelepon, item.Alamat, item.Status)
	if err != nil {
		// Rollback user jika insert guru gagal
		config.DB.Exec("DELETE FROM users WHERE id=?", userID)
		http.Error(w, "Gagal menyimpan data guru: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Guru berhasil ditambahkan"})
}

func updateGuru(w http.ResponseWriter, r *http.Request, id string) {
	var item GuruInput
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update profil guru
	_, err := config.DB.Exec("UPDATE guru SET nip=?, nama_lengkap=?, jenis_kelamin=?, no_telepon=?, alamat=?, status=? WHERE id=?",
		item.NIP, item.NamaLengkap, item.JenisKelamin, item.NoTelepon, item.Alamat, item.Status, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update email & username di tabel users juga
	config.DB.Exec("UPDATE users u JOIN guru g ON u.id=g.user_id SET u.email=?, u.username=? WHERE g.id=?",
		item.Email, item.Username, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Guru berhasil diubah"})
}

func deleteGuru(w http.ResponseWriter, r *http.Request, id string) {
	// Ambil user_id dulu
	var userID int
	config.DB.QueryRow("SELECT user_id FROM guru WHERE id=?", id).Scan(&userID)

	_, err := config.DB.Exec("DELETE FROM guru WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if userID > 0 {
		config.DB.Exec("DELETE FROM users WHERE id=?", userID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Guru berhasil dihapus"})
}
