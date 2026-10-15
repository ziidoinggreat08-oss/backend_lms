package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend_lms/config"
)

type MataPelajaran struct {
	ID        int    `json:"id"`
	KodeMapel string `json:"kode_mapel"`
	NamaMapel string `json:"nama_mapel"`
	Kelompok  string `json:"kelompok"`
	Deskripsi string `json:"deskripsi"`
}

func HandleMataPelajaran(w http.ResponseWriter, r *http.Request) {
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
			getMataPelajarans(w, r)
		}
	case "POST":
		createMataPelajaran(w, r)
	case "PUT":
		updateMataPelajaran(w, r, id)
	case "DELETE":
		deleteMataPelajaran(w, r, id)
	}
}

func getMataPelajarans(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, kode_mapel, nama_mapel, COALESCE(kelompok,''), COALESCE(deskripsi,'') FROM mata_pelajaran ORDER BY nama_mapel")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []MataPelajaran{}
	for rows.Next() {
		var item MataPelajaran
		if err := rows.Scan(&item.ID, &item.KodeMapel, &item.NamaMapel, &item.Kelompok, &item.Deskripsi); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func createMataPelajaran(w http.ResponseWriter, r *http.Request) {
	var item MataPelajaran
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("INSERT INTO mata_pelajaran (kode_mapel, nama_mapel, kelompok, deskripsi) VALUES (?, ?, ?, ?)",
		item.KodeMapel, item.NamaMapel, item.Kelompok, item.Deskripsi)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Mata pelajaran berhasil ditambahkan"})
}

func updateMataPelajaran(w http.ResponseWriter, r *http.Request, id string) {
	var item MataPelajaran
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("UPDATE mata_pelajaran SET kode_mapel=?, nama_mapel=?, kelompok=?, deskripsi=? WHERE id=?",
		item.KodeMapel, item.NamaMapel, item.Kelompok, item.Deskripsi, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Mata pelajaran berhasil diubah"})
}

func deleteMataPelajaran(w http.ResponseWriter, r *http.Request, id string) {
	_, err := config.DB.Exec("DELETE FROM mata_pelajaran WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Mata pelajaran berhasil dihapus"})
}
