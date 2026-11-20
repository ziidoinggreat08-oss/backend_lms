package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"backend_lms/config"
	"github.com/golang-jwt/jwt/v5"
)

// Helper: ambil user_id dan role dari Authorization Bearer token
func getClaimsFromRequest(r *http.Request) (*Claims, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, false
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})

	if err != nil || !token.Valid {
		return nil, false
	}
	return claims, true
}

// Helper: cari ID guru berdasarkan user_id login
func getGuruIDByUserID(userID int) int {
	var guruID int
	err := config.DB.QueryRow("SELECT id FROM guru WHERE user_id = ?", userID).Scan(&guruID)
	if err != nil {
		return 1 // fallback ke 1 jika belum termapping
	}
	return guruID
}

// -------------------------------------------------------------
// 1. DASHBOARD & STATS GURU
// -------------------------------------------------------------
type GuruDashboardData struct {
	NamaGuru      string       `json:"nama_guru"`
	NIP           string       `json:"nip"`
	TotalKelas    int          `json:"total_kelas"`
	TotalMateri   int          `json:"total_materi"`
	TotalTugas    int          `json:"total_tugas"`
	TotalKuis     int          `json:"total_kuis"`
	TotalUjian    int          `json:"total_ujian"`
	KelasAjar     []KelasAjar  `json:"kelas_ajar"`
	MateriTerbaru []MateriItem `json:"materi_terbaru"`
	TugasTerbaru  []TugasItem  `json:"tugas_terbaru"`
}

type KelasAjar struct {
	KelasID   int    `json:"kelas_id"`
	NamaKelas string `json:"nama_kelas"`
	Tingkat   string `json:"tingkat"`
	MapelID   int    `json:"mapel_id"`
	NamaMapel string `json:"nama_mapel"`
	KodeMapel string `json:"kode_mapel"`
}

func HandleGuruDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	guruID := 1
	if claims, ok := getClaimsFromRequest(r); ok {
		guruID = getGuruIDByUserID(claims.UserID)
	}

	var data GuruDashboardData

	// Info Guru
	config.DB.QueryRow("SELECT nama_lengkap, COALESCE(nip, '') FROM guru WHERE id = ?", guruID).
		Scan(&data.NamaGuru, &data.NIP)

	// Hitung Statistik
	config.DB.QueryRow("SELECT COUNT(DISTINCT kelas_id) FROM guru_mengajar WHERE guru_id = ?", guruID).Scan(&data.TotalKelas)
	config.DB.QueryRow("SELECT COUNT(*) FROM materi WHERE guru_id = ?", guruID).Scan(&data.TotalMateri)
	config.DB.QueryRow("SELECT COUNT(*) FROM tugas WHERE guru_id = ?", guruID).Scan(&data.TotalTugas)
	config.DB.QueryRow("SELECT COUNT(*) FROM kuis_ujian WHERE guru_id = ? AND tipe = 'kuis'", guruID).Scan(&data.TotalKuis)
	config.DB.QueryRow("SELECT COUNT(*) FROM kuis_ujian WHERE guru_id = ? AND tipe = 'ujian'", guruID).Scan(&data.TotalUjian)

	// Daftar Kelas yang Diajar
	rows, err := config.DB.Query(`
		SELECT gm.kelas_id, k.nama_kelas, k.tingkat, gm.mata_pelajaran_id, mp.nama_mapel, mp.kode_mapel
		FROM guru_mengajar gm
		JOIN kelas k ON k.id = gm.kelas_id
		JOIN mata_pelajaran mp ON mp.id = gm.mata_pelajaran_id
		WHERE gm.guru_id = ?
		ORDER BY k.tingkat, k.nama_kelas
	`, guruID)

	data.KelasAjar = []KelasAjar{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ka KelasAjar
			if err := rows.Scan(&ka.KelasID, &ka.NamaKelas, &ka.Tingkat, &ka.MapelID, &ka.NamaMapel, &ka.KodeMapel); err == nil {
				data.KelasAjar = append(data.KelasAjar, ka)
			}
		}
	}

	// Materi Terbaru (Limit 5)
	rowsMateri, err := config.DB.Query(`
		SELECT m.id, m.guru_id, m.mata_pelajaran_id, mp.nama_mapel, m.kelas_id, k.nama_kelas,
		       m.judul, COALESCE(m.deskripsi, ''), COALESCE(m.file_url, ''), COALESCE(m.tautan_eksternal, ''),
		       DATE_FORMAT(m.created_at, '%Y-%m-%d %H:%i')
		FROM materi m
		JOIN mata_pelajaran mp ON mp.id = m.mata_pelajaran_id
		JOIN kelas k ON k.id = m.kelas_id
		WHERE m.guru_id = ?
		ORDER BY m.id DESC LIMIT 5
	`, guruID)
	data.MateriTerbaru = []MateriItem{}
	if err == nil {
		defer rowsMateri.Close()
		for rowsMateri.Next() {
			var m MateriItem
			if err := rowsMateri.Scan(&m.ID, &m.GuruID, &m.MapelID, &m.NamaMapel, &m.KelasID, &m.NamaKelas, &m.Judul, &m.Deskripsi, &m.FileURL, &m.Tautan, &m.CreatedAt); err == nil {
				data.MateriTerbaru = append(data.MateriTerbaru, m)
			}
		}
	}

	// Tugas Terbaru (Limit 5)
	rowsTugas, err := config.DB.Query(`
		SELECT t.id, t.guru_id, t.mata_pelajaran_id, mp.nama_mapel, t.kelas_id, k.nama_kelas,
		       t.judul, COALESCE(t.deskripsi, ''), DATE_FORMAT(t.deadline, '%Y-%m-%d %H:%i'),
		       DATE_FORMAT(t.created_at, '%Y-%m-%d %H:%i')
		FROM tugas t
		JOIN mata_pelajaran mp ON mp.id = t.mata_pelajaran_id
		JOIN kelas k ON k.id = t.kelas_id
		WHERE t.guru_id = ?
		ORDER BY t.id DESC LIMIT 5
	`, guruID)
	data.TugasTerbaru = []TugasItem{}
	if err == nil {
		defer rowsTugas.Close()
		for rowsTugas.Next() {
			var t TugasItem
			if err := rowsTugas.Scan(&t.ID, &t.GuruID, &t.MapelID, &t.NamaMapel, &t.KelasID, &t.NamaKelas, &t.Judul, &t.Deskripsi, &t.Deadline, &t.CreatedAt); err == nil {
				data.TugasTerbaru = append(data.TugasTerbaru, t)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// -------------------------------------------------------------
// 2. MATERI PEMBELAJARAN
// -------------------------------------------------------------
type MateriItem struct {
	ID        int    `json:"id"`
	GuruID    int    `json:"guru_id"`
	MapelID   int    `json:"mata_pelajaran_id"`
	NamaMapel string `json:"nama_mapel"`
	KelasID   int    `json:"kelas_id"`
	NamaKelas string `json:"nama_kelas"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	FileURL   string `json:"file_url"`
	Tautan    string `json:"tautan_eksternal"`
	CreatedAt string `json:"created_at"`
}

type MateriInput struct {
	MapelID   int    `json:"mata_pelajaran_id"`
	KelasID   int    `json:"kelas_id"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	FileURL   string `json:"file_url"`
	Tautan    string `json:"tautan_eksternal"`
}

func HandleGuruMateri(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	guruID := 1
	if claims, ok := getClaimsFromRequest(r); ok {
		guruID = getGuruIDByUserID(claims.UserID)
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := ""
	if len(parts) > 3 {
		id = parts[3]
	}

	switch r.Method {
	case "GET":
		getMateriList(w, r, guruID)
	case "POST":
		createMateri(w, r, guruID)
	case "PUT":
		updateMateri(w, r, guruID, id)
	case "DELETE":
		deleteMateri(w, r, guruID, id)
	}
}

func getMateriList(w http.ResponseWriter, r *http.Request, guruID int) {
	kelasFilter := r.URL.Query().Get("kelas_id")
	query := `
		SELECT m.id, m.guru_id, m.mata_pelajaran_id, mp.nama_mapel, m.kelas_id, k.nama_kelas,
		       m.judul, COALESCE(m.deskripsi, ''), COALESCE(m.file_url, ''), COALESCE(m.tautan_eksternal, ''),
		       DATE_FORMAT(m.created_at, '%Y-%m-%d %H:%i')
		FROM materi m
		JOIN mata_pelajaran mp ON mp.id = m.mata_pelajaran_id
		JOIN kelas k ON k.id = m.kelas_id
		WHERE m.guru_id = ?
	`
	args := []interface{}{guruID}

	if kelasFilter != "" {
		if kID, err := strconv.Atoi(kelasFilter); err == nil && kID > 0 {
			query += " AND m.kelas_id = ?"
			args = append(args, kID)
		}
	}
	query += " ORDER BY m.id DESC"

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []MateriItem{}
	for rows.Next() {
		var m MateriItem
		if err := rows.Scan(&m.ID, &m.GuruID, &m.MapelID, &m.NamaMapel, &m.KelasID, &m.NamaKelas, &m.Judul, &m.Deskripsi, &m.FileURL, &m.Tautan, &m.CreatedAt); err == nil {
			items = append(items, m)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func createMateri(w http.ResponseWriter, r *http.Request, guruID int) {
	var input MateriInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(input.Judul) == "" || input.MapelID <= 0 || input.KelasID <= 0 {
		http.Error(w, "Judul, mata pelajaran, dan kelas wajib diisi", http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec(`
		INSERT INTO materi (guru_id, mata_pelajaran_id, kelas_id, judul, deskripsi, file_url, tautan_eksternal)
		VALUES (?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''))
	`, guruID, input.MapelID, input.KelasID, input.Judul, input.Deskripsi, input.FileURL, input.Tautan)

	if err != nil {
		http.Error(w, "Gagal menyimpan materi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Materi pembelajaran berhasil ditambahkan"})
}

func updateMateri(w http.ResponseWriter, r *http.Request, guruID int, id string) {
	var input MateriInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec(`
		UPDATE materi SET mata_pelajaran_id=?, kelas_id=?, judul=?, deskripsi=?, file_url=NULLIF(?, ''), tautan_eksternal=NULLIF(?, '')
		WHERE id=? AND guru_id=?
	`, input.MapelID, input.KelasID, input.Judul, input.Deskripsi, input.FileURL, input.Tautan, id, guruID)

	if err != nil {
		http.Error(w, "Gagal mengubah materi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Materi berhasil diperbarui"})
}

func deleteMateri(w http.ResponseWriter, r *http.Request, guruID int, id string) {
	_, err := config.DB.Exec("DELETE FROM materi WHERE id=? AND guru_id=?", id, guruID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Materi berhasil dihapus"})
}

// -------------------------------------------------------------
// 3. TUGAS & PENILAIAN TUGAS
// -------------------------------------------------------------
type TugasItem struct {
	ID          int    `json:"id"`
	GuruID      int    `json:"guru_id"`
	MapelID     int    `json:"mata_pelajaran_id"`
	NamaMapel   string `json:"nama_mapel"`
	KelasID     int    `json:"kelas_id"`
	NamaKelas   string `json:"nama_kelas"`
	Judul       string `json:"judul"`
	Deskripsi   string `json:"deskripsi"`
	Deadline    string `json:"deadline"`
	CreatedAt   string `json:"created_at"`
	TotalKumpul int    `json:"total_kumpul"`
}

type TugasInput struct {
	MapelID   int    `json:"mata_pelajaran_id"`
	KelasID   int    `json:"kelas_id"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Deadline  string `json:"deadline"`
}

func HandleGuruTugas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	guruID := 1
	if claims, ok := getClaimsFromRequest(r); ok {
		guruID = getGuruIDByUserID(claims.UserID)
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := ""
	if len(parts) > 3 {
		id = parts[3]
	}

	switch r.Method {
	case "GET":
		getTugasList(w, r, guruID)
	case "POST":
		createTugas(w, r, guruID)
	case "PUT":
		updateTugas(w, r, guruID, id)
	case "DELETE":
		deleteTugas(w, r, guruID, id)
	}
}

func getTugasList(w http.ResponseWriter, r *http.Request, guruID int) {
	query := `
		SELECT t.id, t.guru_id, t.mata_pelajaran_id, mp.nama_mapel, t.kelas_id, k.nama_kelas,
		       t.judul, COALESCE(t.deskripsi, ''), DATE_FORMAT(t.deadline, '%Y-%m-%d %H:%i'),
		       DATE_FORMAT(t.created_at, '%Y-%m-%d %H:%i'),
		       (SELECT COUNT(*) FROM pengumpulan_tugas pt WHERE pt.tugas_id = t.id) as total_kumpul
		FROM tugas t
		JOIN mata_pelajaran mp ON mp.id = t.mata_pelajaran_id
		JOIN kelas k ON k.id = t.kelas_id
		WHERE t.guru_id = ?
		ORDER BY t.id DESC
	`
	rows, err := config.DB.Query(query, guruID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []TugasItem{}
	for rows.Next() {
		var t TugasItem
		if err := rows.Scan(&t.ID, &t.GuruID, &t.MapelID, &t.NamaMapel, &t.KelasID, &t.NamaKelas, &t.Judul, &t.Deskripsi, &t.Deadline, &t.CreatedAt, &t.TotalKumpul); err == nil {
			items = append(items, t)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func createTugas(w http.ResponseWriter, r *http.Request, guruID int) {
	var input TugasInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(input.Judul) == "" || input.MapelID <= 0 || input.KelasID <= 0 || strings.TrimSpace(input.Deadline) == "" {
		http.Error(w, "Judul, mata pelajaran, kelas, dan batas waktu (deadline) wajib diisi", http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec(`
		INSERT INTO tugas (guru_id, mata_pelajaran_id, kelas_id, judul, deskripsi, deadline)
		VALUES (?, ?, ?, ?, ?, ?)
	`, guruID, input.MapelID, input.KelasID, input.Judul, input.Deskripsi, input.Deadline)

	if err != nil {
		http.Error(w, "Gagal membuat tugas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tugas berhasil dibuat"})
}

func updateTugas(w http.ResponseWriter, r *http.Request, guruID int, id string) {
	var input TugasInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec(`
		UPDATE tugas SET mata_pelajaran_id=?, kelas_id=?, judul=?, deskripsi=?, deadline=?
		WHERE id=? AND guru_id=?
	`, input.MapelID, input.KelasID, input.Judul, input.Deskripsi, input.Deadline, id, guruID)

	if err != nil {
		http.Error(w, "Gagal mengubah tugas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tugas berhasil diubah"})
}

func deleteTugas(w http.ResponseWriter, r *http.Request, guruID int, id string) {
	_, err := config.DB.Exec("DELETE FROM tugas WHERE id=? AND guru_id=?", id, guruID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tugas berhasil dihapus"})
}

// -------------------------------------------------------------
// 4. KUIS & UJIAN
// -------------------------------------------------------------
type KuisUjianItem struct {
	ID          int    `json:"id"`
	GuruID      int    `json:"guru_id"`
	MapelID     int    `json:"mata_pelajaran_id"`
	NamaMapel   string `json:"nama_mapel"`
	KelasID     int    `json:"kelas_id"`
	NamaKelas   string `json:"nama_kelas"`
	Judul       string `json:"judul"`
	Tipe        string `json:"tipe"` // 'kuis' | 'ujian'
	DurasiMenit int    `json:"durasi_menit"`
	Deskripsi   string `json:"deskripsi"`
	Status      string `json:"status"` // 'draft' | 'aktif' | 'selesai'
	CreatedAt   string `json:"created_at"`
	JumlahSoal  int    `json:"jumlah_soal"`
}

type SoalPilihanInput struct {
	Teks  string `json:"teks"`
	Benar bool   `json:"benar"`
}

// Soal dibuat bersama evaluasi agar soal dan opsi tidak pernah tersimpan
// sebagian apabila validasi atau penyimpanan salah satu bagian gagal.
type SoalEvaluasiInput struct {
	Pertanyaan string             `json:"pertanyaan"`
	Bobot      int                `json:"bobot"`
	Pilihan    []SoalPilihanInput `json:"pilihan"`
}

type KuisUjianInput struct {
	MapelID     int                 `json:"mata_pelajaran_id"`
	KelasID     int                 `json:"kelas_id"`
	Judul       string              `json:"judul"`
	Tipe        string              `json:"tipe"`
	DurasiMenit int                 `json:"durasi_menit"`
	Deskripsi   string              `json:"deskripsi"`
	Status      string              `json:"status"`
	Soal        []SoalEvaluasiInput `json:"soal"`
}

func HandleGuruKuisUjian(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := ""
	if len(parts) > 3 {
		id = parts[3]
	}
	forcedType := ""
	if len(parts) >= 3 && (parts[2] == "kuis" || parts[2] == "ujian") {
		forcedType = parts[2]
	}

	claims, ok := getClaimsFromRequest(r)
	if !ok {
		http.Error(w, "Autentikasi diperlukan", http.StatusUnauthorized)
		return
	}
	if claims.Role != "guru" {
		http.Error(w, "Akses hanya untuk guru", http.StatusForbidden)
		return
	}
	guruID := getGuruIDByUserID(claims.UserID)
	if guruID <= 0 {
		http.Error(w, "Profil guru tidak ditemukan", http.StatusForbidden)
		return
	}

	switch r.Method {
	case "GET":
		getKuisUjianList(w, r, guruID, forcedType)
	case "POST":
		createKuisUjian(w, r, guruID, forcedType)
	case "PUT":
		updateKuisUjian(w, r, guruID, id, forcedType)
	case "DELETE":
		deleteKuisUjian(w, r, guruID, id)
	}
}

func getKuisUjianList(w http.ResponseWriter, r *http.Request, guruID int, forcedType string) {
	tipeFilter := r.URL.Query().Get("tipe")
	if forcedType != "" {
		tipeFilter = forcedType
	}
	query := `
		SELECT ku.id, ku.guru_id, ku.mata_pelajaran_id, mp.nama_mapel, ku.kelas_id, k.nama_kelas,
		       ku.judul, ku.tipe, ku.durasi_menit, COALESCE(ku.deskripsi, ''), ku.status,
		       DATE_FORMAT(ku.created_at, '%Y-%m-%d %H:%i'),
		       (SELECT COUNT(*) FROM soal_kuis_ujian s WHERE s.kuis_ujian_id = ku.id)
		FROM kuis_ujian ku
		JOIN mata_pelajaran mp ON mp.id = ku.mata_pelajaran_id
		JOIN kelas k ON k.id = ku.kelas_id
		WHERE ku.guru_id = ?
	`
	args := []interface{}{guruID}

	if tipeFilter != "" {
		query += " AND ku.tipe = ?"
		args = append(args, tipeFilter)
	}
	query += " ORDER BY ku.id DESC"

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []KuisUjianItem{}
	for rows.Next() {
		var item KuisUjianItem
		if err := rows.Scan(&item.ID, &item.GuruID, &item.MapelID, &item.NamaMapel, &item.KelasID, &item.NamaKelas, &item.Judul, &item.Tipe, &item.DurasiMenit, &item.Deskripsi, &item.Status, &item.CreatedAt, &item.JumlahSoal); err == nil {
			items = append(items, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func createKuisUjian(w http.ResponseWriter, r *http.Request, guruID int, forcedType string) {
	var input KuisUjianInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(input.Judul) == "" || input.MapelID <= 0 || input.KelasID <= 0 {
		http.Error(w, "Judul, mata pelajaran, dan kelas wajib diisi", http.StatusBadRequest)
		return
	}
	if forcedType != "" {
		input.Tipe = forcedType
	}
	if input.Tipe != "kuis" && input.Tipe != "ujian" {
		http.Error(w, "Tipe evaluasi tidak valid", http.StatusBadRequest)
		return
	}
	if input.DurasiMenit <= 0 {
		input.DurasiMenit = 60
	}
	if input.Tipe == "" {
		input.Tipe = "kuis"
	}
	if input.Status == "" {
		input.Status = "draft"
	}

	if err := validateSoal(input.Soal); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, "Gagal memulai penyimpanan", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	result, err := tx.Exec(`
		INSERT INTO kuis_ujian (guru_id, mata_pelajaran_id, kelas_id, judul, tipe, durasi_menit, deskripsi, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, guruID, input.MapelID, input.KelasID, input.Judul, input.Tipe, input.DurasiMenit, input.Deskripsi, input.Status)

	if err != nil {
		http.Error(w, "Gagal membuat kuis/ujian: "+err.Error(), http.StatusInternalServerError)
		return
	}
	evaluasiID, _ := result.LastInsertId()
	if err := saveSoal(tx, evaluasiID, input.Soal); err != nil {
		http.Error(w, "Gagal menyimpan soal", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, "Gagal menyimpan evaluasi", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Kuis/Ujian berhasil dibuat", "id": evaluasiID})
}

func updateKuisUjian(w http.ResponseWriter, r *http.Request, guruID int, id string, forcedType string) {
	var input KuisUjianInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if id == "" || strings.TrimSpace(input.Judul) == "" || input.MapelID <= 0 || input.KelasID <= 0 || input.DurasiMenit <= 0 {
		http.Error(w, "Data evaluasi tidak valid", http.StatusBadRequest)
		return
	}
	if forcedType != "" {
		input.Tipe = forcedType
	}
	if input.Tipe != "kuis" && input.Tipe != "ujian" {
		http.Error(w, "Tipe evaluasi tidak valid", http.StatusBadRequest)
		return
	}
	if err := validateSoal(input.Soal); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, "Gagal memulai penyimpanan", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	result, err := tx.Exec(`
		UPDATE kuis_ujian SET mata_pelajaran_id=?, kelas_id=?, judul=?, tipe=?, durasi_menit=?, deskripsi=?, status=?
		WHERE id=? AND guru_id=?
	`, input.MapelID, input.KelasID, input.Judul, input.Tipe, input.DurasiMenit, input.Deskripsi, input.Status, id, guruID)

	if err != nil {
		http.Error(w, "Gagal mengubah kuis/ujian: "+err.Error(), http.StatusInternalServerError)
		return
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		http.Error(w, "Evaluasi tidak ditemukan", http.StatusNotFound)
		return
	}
	if _, err := tx.Exec("DELETE FROM soal_kuis_ujian WHERE kuis_ujian_id=?", id); err != nil {
		http.Error(w, "Gagal memperbarui soal", http.StatusInternalServerError)
		return
	}
	evaluasiID, _ := strconv.ParseInt(id, 10, 64)
	if err := saveSoal(tx, evaluasiID, input.Soal); err != nil {
		http.Error(w, "Gagal menyimpan soal", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, "Gagal menyimpan evaluasi", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kuis/Ujian berhasil diubah"})
}

func validateSoal(soal []SoalEvaluasiInput) error {
	if len(soal) == 0 {
		return fmt.Errorf("Minimal satu soal wajib dibuat")
	}
	for nomor, s := range soal {
		if strings.TrimSpace(s.Pertanyaan) == "" || len(s.Pilihan) < 2 || s.Bobot <= 0 {
			return fmt.Errorf("Soal nomor %d belum lengkap", nomor+1)
		}
		benar := 0
		for _, p := range s.Pilihan {
			if strings.TrimSpace(p.Teks) == "" {
				return fmt.Errorf("Pilihan pada soal nomor %d tidak boleh kosong", nomor+1)
			}
			if p.Benar {
				benar++
			}
		}
		if benar != 1 {
			return fmt.Errorf("Soal nomor %d harus mempunyai tepat satu jawaban benar", nomor+1)
		}
	}
	return nil
}

func saveSoal(tx *sql.Tx, evaluasiID int64, soal []SoalEvaluasiInput) error {
	for nomor, s := range soal {
		result, err := tx.Exec("INSERT INTO soal_kuis_ujian (kuis_ujian_id, pertanyaan, bobot, urutan) VALUES (?, ?, ?, ?)", evaluasiID, strings.TrimSpace(s.Pertanyaan), s.Bobot, nomor+1)
		if err != nil {
			return err
		}
		soalID, _ := result.LastInsertId()
		for urutan, p := range s.Pilihan {
			if _, err := tx.Exec("INSERT INTO pilihan_soal_kuis_ujian (soal_id, teks, is_benar, urutan) VALUES (?, ?, ?, ?)", soalID, strings.TrimSpace(p.Teks), p.Benar, urutan+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteKuisUjian(w http.ResponseWriter, r *http.Request, guruID int, id string) {
	_, err := config.DB.Exec("DELETE FROM kuis_ujian WHERE id=? AND guru_id=?", id, guruID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kuis/Ujian berhasil dihapus"})
}

// -------------------------------------------------------------
// 5. REKAP & INPUT NILAI SISWA
// -------------------------------------------------------------
type NilaiSiswaItem struct {
	ID         int     `json:"id"`
	SiswaID    int     `json:"siswa_id"`
	NIS        string  `json:"nis"`
	NamaSiswa  string  `json:"nama_siswa"`
	KelasID    int     `json:"kelas_id"`
	NamaKelas  string  `json:"nama_kelas"`
	MapelID    int     `json:"mata_pelajaran_id"`
	NamaMapel  string  `json:"nama_mapel"`
	NilaiTugas float64 `json:"nilai_tugas"`
	NilaiKuis  float64 `json:"nilai_kuis"`
	NilaiUjian float64 `json:"nilai_ujian"`
	NilaiAkhir float64 `json:"nilai_akhir"`
	Catatan    string  `json:"catatan"`
}

type NilaiInput struct {
	SiswaID    int     `json:"siswa_id"`
	MapelID    int     `json:"mata_pelajaran_id"`
	KelasID    int     `json:"kelas_id"`
	NilaiTugas float64 `json:"nilai_tugas"`
	NilaiKuis  float64 `json:"nilai_kuis"`
	NilaiUjian float64 `json:"nilai_ujian"`
	Catatan    string  `json:"catatan"`
}

func HandleGuruNilai(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	guruID := 1
	if claims, ok := getClaimsFromRequest(r); ok {
		guruID = getGuruIDByUserID(claims.UserID)
	}

	switch r.Method {
	case "GET":
		getNilaiList(w, r, guruID)
	case "POST":
		saveNilai(w, r, guruID)
	}
}

func getNilaiList(w http.ResponseWriter, r *http.Request, guruID int) {
	kelasFilter := r.URL.Query().Get("kelas_id")
	mapelFilter := r.URL.Query().Get("mapel_id")

	query := `
		SELECT COALESCE(n.id, 0), s.id as siswa_id, s.nis, s.nama_lengkap,
		       s.kelas_id, k.nama_kelas,
		       COALESCE(n.mata_pelajaran_id, ?), COALESCE(mp.nama_mapel, ''),
		       COALESCE(n.nilai_tugas, 0), COALESCE(n.nilai_kuis, 0), COALESCE(n.nilai_ujian, 0),
		       COALESCE(n.nilai_akhir, 0), COALESCE(n.catatan, '')
		FROM siswa s
		JOIN kelas k ON k.id = s.kelas_id
		LEFT JOIN nilai_siswa n ON n.siswa_id = s.id AND n.mata_pelajaran_id = ?
		LEFT JOIN mata_pelajaran mp ON mp.id = ?
		WHERE 1=1
	`
	mapelID, _ := strconv.Atoi(mapelFilter)
	if mapelID <= 0 {
		mapelID = 5 // default ke mapel pertama guru
	}

	args := []interface{}{mapelID, mapelID, mapelID}

	if kelasFilter != "" {
		if kID, err := strconv.Atoi(kelasFilter); err == nil && kID > 0 {
			query += " AND s.kelas_id = ?"
			args = append(args, kID)
		}
	}
	query += " ORDER BY s.nama_lengkap"

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []NilaiSiswaItem{}
	for rows.Next() {
		var n NilaiSiswaItem
		if err := rows.Scan(&n.ID, &n.SiswaID, &n.NIS, &n.NamaSiswa, &n.KelasID, &n.NamaKelas, &n.MapelID, &n.NamaMapel, &n.NilaiTugas, &n.NilaiKuis, &n.NilaiUjian, &n.NilaiAkhir, &n.Catatan); err == nil {
			items = append(items, n)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func saveNilai(w http.ResponseWriter, r *http.Request, guruID int) {
	var input NilaiInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hitung bobot nilai akhir: Tugas (30%), Kuis (30%), Ujian (40%)
	nilaiAkhir := (input.NilaiTugas * 0.3) + (input.NilaiKuis * 0.3) + (input.NilaiUjian * 0.4)

	query := `
		INSERT INTO nilai_siswa (siswa_id, mata_pelajaran_id, kelas_id, guru_id, nilai_tugas, nilai_kuis, nilai_ujian, nilai_akhir, catatan)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			nilai_tugas = VALUES(nilai_tugas),
			nilai_kuis = VALUES(nilai_kuis),
			nilai_ujian = VALUES(nilai_ujian),
			nilai_akhir = VALUES(nilai_akhir),
			catatan = VALUES(catatan),
			guru_id = VALUES(guru_id)
	`

	_, err := config.DB.Exec(query, input.SiswaID, input.MapelID, input.KelasID, guruID, input.NilaiTugas, input.NilaiKuis, input.NilaiUjian, nilaiAkhir, input.Catatan)
	if err != nil {
		http.Error(w, "Gagal menyimpan nilai: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Nilai siswa berhasil disimpan",
		"nilai_akhir": nilaiAkhir,
	})
}
