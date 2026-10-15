package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"backend_lms/config"
)

type Jurusan struct {
	ID          int    `json:"id"`
	KodeJurusan string `json:"kode_jurusan"`
	NamaJurusan string `json:"nama_jurusan"`
	Deskripsi   string `json:"deskripsi"`
}

func HandleJurusan(w http.ResponseWriter, r *http.Request) {
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
			getJurusans(w, r)
		}
	case "POST":
		createJurusan(w, r)
	case "PUT":
		updateJurusan(w, r, id)
	case "DELETE":
		deleteJurusan(w, r, id)
	}
}

func getJurusans(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, kode_jurusan, nama_jurusan, COALESCE(deskripsi,'') FROM jurusan ORDER BY id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []Jurusan{}
	for rows.Next() {
		var item Jurusan
		if err := rows.Scan(&item.ID, &item.KodeJurusan, &item.NamaJurusan, &item.Deskripsi); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func createJurusan(w http.ResponseWriter, r *http.Request) {
	var item Jurusan
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("INSERT INTO jurusan (kode_jurusan, nama_jurusan, deskripsi) VALUES (?, ?, ?)",
		item.KodeJurusan, item.NamaJurusan, item.Deskripsi)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Jurusan berhasil ditambahkan"})
}

func updateJurusan(w http.ResponseWriter, r *http.Request, id string) {
	var item Jurusan
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("UPDATE jurusan SET kode_jurusan=?, nama_jurusan=?, deskripsi=? WHERE id=?",
		item.KodeJurusan, item.NamaJurusan, item.Deskripsi, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Jurusan berhasil diubah"})
}

func deleteJurusan(w http.ResponseWriter, r *http.Request, id string) {
	_, err := config.DB.Exec("DELETE FROM jurusan WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Jurusan berhasil dihapus"})
}

// ListJurusan tetap ada agar tidak break kode lama jika ada referensi
func ListJurusan(w http.ResponseWriter, r *http.Request) {
	var _ sql.DB
	getJurusans(w, r)
}
