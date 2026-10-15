package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend_lms/config"
)

type KepalaSekolah struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	NIP         string `json:"nip"`
	NamaLengkap string `json:"nama_lengkap"`
	NoTelepon   string `json:"no_telepon"`
	Email       string `json:"email"`
	Username    string `json:"username"`
}

func HandleKepalaSekolah(w http.ResponseWriter, r *http.Request) {
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
			getKepalaSekolahs(w, r)
		}
	case "POST":
		createKepalaSekolah(w, r)
	case "PUT":
		updateKepalaSekolah(w, r, id)
	case "DELETE":
		deleteKepalaSekolah(w, r, id)
	}
}

func getKepalaSekolahs(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(`
		SELECT ks.id, ks.user_id, COALESCE(ks.nip,''), ks.nama_lengkap,
		       COALESCE(ks.no_telepon,''), COALESCE(u.email,''), COALESCE(u.username,'')
		FROM kepala_sekolah ks
		LEFT JOIN users u ON u.id = ks.user_id
		ORDER BY ks.nama_lengkap
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []KepalaSekolah{}
	for rows.Next() {
		var item KepalaSekolah
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

type KepalaSekolahInput struct {
	NIP         string `json:"nip"`
	NamaLengkap string `json:"nama_lengkap"`
	NoTelepon   string `json:"no_telepon"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

func createKepalaSekolah(w http.ResponseWriter, r *http.Request) {
	var item KepalaSekolahInput
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := config.DB.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, 'dummy_password', 'kepala_sekolah')",
		item.Username, item.Email)
	if err != nil {
		http.Error(w, "Gagal membuat akun: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userID, _ := res.LastInsertId()

	_, err = config.DB.Exec("INSERT INTO kepala_sekolah (user_id, nip, nama_lengkap, no_telepon) VALUES (?, ?, ?, ?)",
		userID, item.NIP, item.NamaLengkap, item.NoTelepon)
	if err != nil {
		config.DB.Exec("DELETE FROM users WHERE id=?", userID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kepala Sekolah berhasil ditambahkan"})
}

func updateKepalaSekolah(w http.ResponseWriter, r *http.Request, id string) {
	var item KepalaSekolahInput
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("UPDATE kepala_sekolah SET nip=?, nama_lengkap=?, no_telepon=? WHERE id=?",
		item.NIP, item.NamaLengkap, item.NoTelepon, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	config.DB.Exec("UPDATE users u JOIN kepala_sekolah ks ON u.id=ks.user_id SET u.email=?, u.username=? WHERE ks.id=?",
		item.Email, item.Username, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kepala Sekolah berhasil diubah"})
}

func deleteKepalaSekolah(w http.ResponseWriter, r *http.Request, id string) {
	var userID int
	config.DB.QueryRow("SELECT user_id FROM kepala_sekolah WHERE id=?", id).Scan(&userID)

	_, err := config.DB.Exec("DELETE FROM kepala_sekolah WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if userID > 0 {
		config.DB.Exec("DELETE FROM users WHERE id=?", userID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kepala Sekolah berhasil dihapus"})
}
