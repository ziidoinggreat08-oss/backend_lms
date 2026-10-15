package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend_lms/config"
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

type KurikulumInput struct {
	NIP         string `json:"nip"`
	NamaLengkap string `json:"nama_lengkap"`
	NoTelepon   string `json:"no_telepon"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

func createKurikulum(w http.ResponseWriter, r *http.Request) {
	var item KurikulumInput
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := config.DB.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, 'dummy_password', 'kurikulum')",
		item.Username, item.Email)
	if err != nil {
		http.Error(w, "Gagal membuat akun: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userID, _ := res.LastInsertId()

	_, err = config.DB.Exec("INSERT INTO kurikulum (user_id, nip, nama_lengkap, no_telepon) VALUES (?, ?, ?, ?)",
		userID, item.NIP, item.NamaLengkap, item.NoTelepon)
	if err != nil {
		config.DB.Exec("DELETE FROM users WHERE id=?", userID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	_, err := config.DB.Exec("UPDATE kurikulum SET nip=?, nama_lengkap=?, no_telepon=? WHERE id=?",
		item.NIP, item.NamaLengkap, item.NoTelepon, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	config.DB.Exec("UPDATE users u JOIN kurikulum k ON u.id=k.user_id SET u.email=?, u.username=? WHERE k.id=?",
		item.Email, item.Username, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kurikulum berhasil diubah"})
}

func deleteKurikulum(w http.ResponseWriter, r *http.Request, id string) {
	var userID int
	config.DB.QueryRow("SELECT user_id FROM kurikulum WHERE id=?", id).Scan(&userID)

	_, err := config.DB.Exec("DELETE FROM kurikulum WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if userID > 0 {
		config.DB.Exec("DELETE FROM users WHERE id=?", userID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kurikulum berhasil dihapus"})
}
