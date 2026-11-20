package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend_lms/config"
	"golang.org/x/crypto/bcrypt"
)

type Kurikulum struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	NIP         string `json:"nip"`
	NamaLengkap string `json:"nama_lengkap"`
	NoTelepon   string `json:"no_telepon"`
	Email       string `json:"email"`
	Username    string `json:"username"`
}

type KurikulumInput struct {
	NIP         string `json:"nip"`
	NamaLengkap string `json:"nama_lengkap"`
	NoTelepon   string `json:"no_telepon"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

func HandleKurikulum(w http.ResponseWriter, r *http.Request) {
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
			getKurikulums(w, r)
		} else {
			getKurikulumByID(w, id)
		}
	case "POST":
		createKurikulum(w, r)
	case "PUT":
		updateKurikulum(w, r, id)
	case "DELETE":
		deleteKurikulum(w, r, id)
	}
}

func getKurikulums(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(`
		SELECT k.id, k.user_id, COALESCE(k.nip,''), k.nama_lengkap,
		       COALESCE(k.no_telepon,''), COALESCE(u.email,''), COALESCE(u.username,'')
		FROM kurikulum k
		LEFT JOIN users u ON u.id = k.user_id
		ORDER BY k.nama_lengkap
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []Kurikulum{}
	for rows.Next() {
		var item Kurikulum
		if err := rows.Scan(&item.ID, &item.UserID, &item.NIP, &item.NamaLengkap,
			&item.NoTelepon, &item.Email, &item.Username); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func getKurikulumByID(w http.ResponseWriter, id string) {
	var item Kurikulum
	err := config.DB.QueryRow(`
		SELECT k.id, k.user_id, COALESCE(k.nip,''), k.nama_lengkap,
		       COALESCE(k.no_telepon,''), COALESCE(u.email,''), COALESCE(u.username,'')
		FROM kurikulum k
		LEFT JOIN users u ON u.id = k.user_id
		WHERE k.id = ?
	`, id).Scan(&item.ID, &item.UserID, &item.NIP, &item.NamaLengkap,
		&item.NoTelepon, &item.Email, &item.Username)

	if err != nil {
		http.Error(w, "Kurikulum tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func createKurikulum(w http.ResponseWriter, r *http.Request) {
	var item KurikulumInput
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(item.NamaLengkap) == "" || strings.TrimSpace(item.Username) == "" || strings.TrimSpace(item.Email) == "" {
		http.Error(w, "Nama lengkap, username, dan email wajib diisi", http.StatusBadRequest)
		return
	}

	pwd := strings.TrimSpace(item.Password)
	if pwd == "" {
		pwd = "kurikulum123"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Gagal mengenkripsi password", http.StatusInternalServerError)
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, "Gagal memulai transaksi", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, 'kurikulum')",
		item.Username, item.Email, string(hash))
	if err != nil {
		http.Error(w, "Username atau email sudah digunakan: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := res.LastInsertId()

	_, err = tx.Exec("INSERT INTO kurikulum (user_id, nip, nama_lengkap, no_telepon) VALUES (?, NULLIF(?, ''), ?, NULLIF(?, ''))",
		userID, item.NIP, item.NamaLengkap, item.NoTelepon)
	if err != nil {
		http.Error(w, "Gagal menyimpan data kurikulum: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Gagal menyimpan perubahan", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kurikulum berhasil ditambahkan"})
}

func updateKurikulum(w http.ResponseWriter, r *http.Request, id string) {
	var item KurikulumInput
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userID int
	err := config.DB.QueryRow("SELECT user_id FROM kurikulum WHERE id = ?", id).Scan(&userID)
	if err != nil {
		http.Error(w, "Kurikulum tidak ditemukan", http.StatusNotFound)
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, "Gagal memulai transaksi", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE kurikulum SET nip=NULLIF(?, ''), nama_lengkap=?, no_telepon=NULLIF(?, '') WHERE id=?",
		item.NIP, item.NamaLengkap, item.NoTelepon, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE users SET email=?, username=? WHERE id=?", item.Email, item.Username, userID)
	if err != nil {
		http.Error(w, "Username atau email sudah digunakan: "+err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(item.Password) != "" {
		if len(item.Password) < 6 {
			http.Error(w, "Password minimal 6 karakter", http.StatusBadRequest)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
		if err == nil {
			tx.Exec("UPDATE users SET password=? WHERE id=?", string(hash), userID)
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Gagal memperbarui data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kurikulum berhasil diubah"})
}

func deleteKurikulum(w http.ResponseWriter, r *http.Request, id string) {
	var userID int
	err := config.DB.QueryRow("SELECT user_id FROM kurikulum WHERE id=?", id).Scan(&userID)
	if err != nil {
		http.Error(w, "Kurikulum tidak ditemukan", http.StatusNotFound)
		return
	}

	_, err = config.DB.Exec("DELETE FROM users WHERE id=?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kurikulum berhasil dihapus"})
}
