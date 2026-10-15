package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend_lms/config"
)

type Kelas struct {
	ID            int    `json:"id"`
	NamaKelas     string `json:"nama_kelas"`
	Tingkat       string `json:"tingkat"`
	JurusanID     int    `json:"jurusan_id"`
	NamaJurusan   string `json:"nama_jurusan"`
	TahunAjaranID int    `json:"tahun_ajaran_id"`
}

func HandleKelas(w http.ResponseWriter, r *http.Request) {
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
			getKelas(w, r)
		}
	case "POST":
		createKelas(w, r)
	case "PUT":
		updateKelas(w, r, id)
	case "DELETE":
		deleteKelas(w, r, id)
	}
}

func getKelas(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(`
		SELECT k.id, k.nama_kelas, k.tingkat, k.jurusan_id, COALESCE(j.nama_jurusan,''), k.tahun_ajaran_id
		FROM kelas k
		LEFT JOIN jurusan j ON j.id = k.jurusan_id
		ORDER BY k.tingkat, k.nama_kelas
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []Kelas{}
	for rows.Next() {
		var item Kelas
		if err := rows.Scan(&item.ID, &item.NamaKelas, &item.Tingkat, &item.JurusanID, &item.NamaJurusan, &item.TahunAjaranID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func createKelas(w http.ResponseWriter, r *http.Request) {
	var item Kelas
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("INSERT INTO kelas (nama_kelas, tingkat, jurusan_id, tahun_ajaran_id) VALUES (?, ?, ?, ?)",
		item.NamaKelas, item.Tingkat, item.JurusanID, item.TahunAjaranID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kelas berhasil ditambahkan"})
}

func updateKelas(w http.ResponseWriter, r *http.Request, id string) {
	var item Kelas
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("UPDATE kelas SET nama_kelas=?, tingkat=?, jurusan_id=?, tahun_ajaran_id=? WHERE id=?",
		item.NamaKelas, item.Tingkat, item.JurusanID, item.TahunAjaranID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kelas berhasil diubah"})
}

func deleteKelas(w http.ResponseWriter, r *http.Request, id string) {
	_, err := config.DB.Exec("DELETE FROM kelas WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kelas berhasil dihapus"})
}

// ListKelas tetap ada agar tidak break kode lama
func ListKelas(w http.ResponseWriter, r *http.Request) {
	getKelas(w, r)
}
