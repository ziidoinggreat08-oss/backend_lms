package controllers

import (
	"encoding/json"
	"net/http"

	"backend_lms/config"
)

type DashboardStats struct {
	TotalSiswa   int `json:"total_siswa"`
	TotalGuru    int `json:"total_guru"`
	TotalKelas   int `json:"total_kelas"`
	TotalJurusan int `json:"total_jurusan"`
	TotalMapel   int `json:"total_mapel"`
}

func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var stats DashboardStats

	// Hitung total Siswa
	config.DB.QueryRow("SELECT COUNT(*) FROM siswa").Scan(&stats.TotalSiswa)
	// Hitung total Guru
	config.DB.QueryRow("SELECT COUNT(*) FROM guru").Scan(&stats.TotalGuru)
	// Hitung total Kelas
	config.DB.QueryRow("SELECT COUNT(*) FROM kelas").Scan(&stats.TotalKelas)
	// Hitung total Jurusan
	config.DB.QueryRow("SELECT COUNT(*) FROM jurusan").Scan(&stats.TotalJurusan)
	// Hitung total Mata Pelajaran
	config.DB.QueryRow("SELECT COUNT(*) FROM mata_pelajaran").Scan(&stats.TotalMapel)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
