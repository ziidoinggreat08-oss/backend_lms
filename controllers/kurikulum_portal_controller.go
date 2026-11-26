package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"backend_lms/config"
	"github.com/golang-jwt/jwt/v5"
)

// Helper verifikasi token khusus Kurikulum / Admin / Kepala Sekolah
func requireKurikulumAccess(w http.ResponseWriter, r *http.Request) bool {
	auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if auth == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Token otorisasi diperlukan"})
		return false
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})
	if err != nil || !token.Valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Token kedaluwarsa atau tidak valid"})
		return false
	}
	if claims.Role != "kurikulum" && claims.Role != "admin" && claims.Role != "kepala_sekolah" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "Akses khusus Kurikulum / Pimpinan Akademik"})
		return false
	}
	return true
}

func HandleKurikulumPortal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if !requireKurikulumAccess(w, r) {
		return
	}

	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	// Format URL: api/kurikulum/{sub}/{id...}
	sub := ""
	idStr := ""
	if len(parts) >= 3 {
		sub = parts[2]
	}
	if len(parts) >= 4 {
		idStr = parts[3]
	}

	switch sub {
	case "dashboard":
		if r.Method == "GET" {
			getKurikulumDashboard(w, r)
		} else {
			methodNotAllowed(w)
		}
	case "akademik":
		if len(parts) >= 5 && parts[4] == "wali-kelas" && r.Method == "PUT" {
			updateWaliKelas(w, r, idStr)
		} else if r.Method == "GET" {
			getKurikulumAkademik(w, r)
		} else {
			methodNotAllowed(w)
		}
	case "pengampu":
		switch r.Method {
		case "GET":
			getKurikulumPengampu(w, r)
		case "POST":
			createKurikulumPengampu(w, r)
		case "PUT":
			updateKurikulumPengampu(w, r, idStr)
		case "DELETE":
			deleteKurikulumPengampu(w, r, idStr)
		default:
			methodNotAllowed(w)
		}
	case "monitoring":
		if len(parts) >= 5 && parts[3] == "evaluasi" && r.Method == "GET" {
			getDetailSoalEvaluasi(w, parts[4])
		} else if len(parts) >= 5 && parts[3] == "tugas" && r.Method == "GET" {
			getDetailPengumpulanTugas(w, parts[4])
		} else if r.Method == "GET" {
			getKurikulumMonitoring(w, r)
		} else {
			methodNotAllowed(w)
		}
	case "nilai":
		if r.Method == "GET" {
			getKurikulumNilai(w, r)
		} else {
			methodNotAllowed(w)
		}
	case "options":
		if r.Method == "GET" {
			getKurikulumOptions(w, r)
		} else {
			methodNotAllowed(w)
		}
	default:
		getKurikulumDashboard(w, r)
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]string{"message": "Metode tidak diizinkan"})
}

// -------------------------------------------------------------
// Options Dropdown Helper (Guru, Mapel, Kelas, Tahun Ajaran)
// -------------------------------------------------------------
func getKurikulumOptions(w http.ResponseWriter, r *http.Request) {
	type OptionItem struct {
		ID   int    `json:"id"`
		Nama string `json:"nama"`
		Info string `json:"info"`
	}

	guruList := []OptionItem{}
	gRows, err := config.DB.Query("SELECT id, nama_lengkap, COALESCE(nip, '-') FROM guru WHERE status = 'aktif' ORDER BY nama_lengkap")
	if err == nil {
		defer gRows.Close()
		for gRows.Next() {
			var o OptionItem
			if err := gRows.Scan(&o.ID, &o.Nama, &o.Info); err == nil {
				guruList = append(guruList, o)
			}
		}
	}

	mapelList := []OptionItem{}
	mRows, err := config.DB.Query("SELECT id, nama_mapel, kode_mapel FROM mata_pelajaran ORDER BY nama_mapel")
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var o OptionItem
			if err := mRows.Scan(&o.ID, &o.Nama, &o.Info); err == nil {
				mapelList = append(mapelList, o)
			}
		}
	}

	kelasList := []OptionItem{}
	kRows, err := config.DB.Query("SELECT k.id, k.nama_kelas, j.kode_jurusan FROM kelas k JOIN jurusan j ON j.id = k.jurusan_id ORDER BY k.tingkat, k.nama_kelas")
	if err == nil {
		defer kRows.Close()
		for kRows.Next() {
			var o OptionItem
			if err := kRows.Scan(&o.ID, &o.Nama, &o.Info); err == nil {
				kelasList = append(kelasList, o)
			}
		}
	}

	var taID int
	var taNama string
	_ = config.DB.QueryRow("SELECT id, CONCAT(tahun, ' (', semester, ')') FROM tahun_ajaran WHERE status_aktif = TRUE LIMIT 1").Scan(&taID, &taNama)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"guru":  guruList,
		"mapel": mapelList,
		"kelas": kelasList,
		"tahun_ajaran": map[string]any{
			"id":   taID,
			"nama": taNama,
		},
	})
}

// -------------------------------------------------------------
// 1. Dashboard Statistik Kurikulum
// -------------------------------------------------------------
func getKurikulumDashboard(w http.ResponseWriter, r *http.Request) {
	var totalGuru, totalSiswa, totalKelas, totalJurusan, totalMapel int
	var totalMateri, totalTugas, totalKuis, totalUjian int

	config.DB.QueryRow("SELECT COUNT(*) FROM guru WHERE status = 'aktif'").Scan(&totalGuru)
	config.DB.QueryRow("SELECT COUNT(*) FROM siswa WHERE status = 'aktif'").Scan(&totalSiswa)
	config.DB.QueryRow("SELECT COUNT(*) FROM kelas").Scan(&totalKelas)
	config.DB.QueryRow("SELECT COUNT(*) FROM jurusan").Scan(&totalJurusan)
	config.DB.QueryRow("SELECT COUNT(*) FROM mata_pelajaran").Scan(&totalMapel)

	// Aktivitas Pembelajaran
	config.DB.QueryRow("SELECT COUNT(*) FROM materi").Scan(&totalMateri)
	config.DB.QueryRow("SELECT COUNT(*) FROM tugas").Scan(&totalTugas)
	config.DB.QueryRow("SELECT COUNT(*) FROM kuis_ujian WHERE tipe = 'kuis'").Scan(&totalKuis)
	config.DB.QueryRow("SELECT COUNT(*) FROM kuis_ujian WHERE tipe = 'ujian'").Scan(&totalUjian)

	// Ringkasan per Jurusan
	type JurusanStat struct {
		ID          int    `json:"id"`
		KodeJurusan string `json:"kode_jurusan"`
		NamaJurusan string `json:"nama_jurusan"`
		TotalKelas  int    `json:"total_kelas"`
		TotalSiswa  int    `json:"total_siswa"`
	}
	jurusanStats := []JurusanStat{}
	jRows, err := config.DB.Query(`
		SELECT j.id, j.kode_jurusan, j.nama_jurusan,
		       COUNT(DISTINCT k.id) as total_kelas,
		       COUNT(DISTINCT s.id) as total_siswa
		FROM jurusan j
		LEFT JOIN kelas k ON k.jurusan_id = j.id
		LEFT JOIN siswa s ON s.jurusan_id = j.id AND s.status = 'aktif'
		GROUP BY j.id, j.kode_jurusan, j.nama_jurusan
		ORDER BY j.kode_jurusan
	`)
	if err == nil {
		defer jRows.Close()
		for jRows.Next() {
			var js JurusanStat
			if err := jRows.Scan(&js.ID, &js.KodeJurusan, &js.NamaJurusan, &js.TotalKelas, &js.TotalSiswa); err == nil {
				jurusanStats = append(jurusanStats, js)
			}
		}
	}

	// Materi dan Tugas Terbaru
	type AktivitasItem struct {
		Tipe      string `json:"tipe"`
		Judul     string `json:"judul"`
		Guru      string `json:"guru"`
		Mapel     string `json:"mapel"`
		Kelas     string `json:"kelas"`
		CreatedAt string `json:"created_at"`
	}
	aktivitas := []AktivitasItem{}
	aRows, err := config.DB.Query(`
		(SELECT 'Materi' as tipe, m.judul, g.nama_lengkap as guru, mp.nama_mapel as mapel, k.nama_kelas as kelas,
		        DATE_FORMAT(m.created_at, '%Y-%m-%d %H:%i') as created_at
		 FROM materi m
		 JOIN guru g ON g.id = m.guru_id
		 JOIN mata_pelajaran mp ON mp.id = m.mata_pelajaran_id
		 JOIN kelas k ON k.id = m.kelas_id
		 ORDER BY m.id DESC LIMIT 5)
		UNION ALL
		(SELECT 'Tugas' as tipe, t.judul, g.nama_lengkap as guru, mp.nama_mapel as mapel, k.nama_kelas as kelas,
		        DATE_FORMAT(t.created_at, '%Y-%m-%d %H:%i') as created_at
		 FROM tugas t
		 JOIN guru g ON g.id = t.guru_id
		 JOIN mata_pelajaran mp ON mp.id = t.mata_pelajaran_id
		 JOIN kelas k ON k.id = t.kelas_id
		 ORDER BY t.id DESC LIMIT 5)
		ORDER BY created_at DESC LIMIT 6
	`)
	if err == nil {
		defer aRows.Close()
		for aRows.Next() {
			var ai AktivitasItem
			if err := aRows.Scan(&ai.Tipe, &ai.Judul, &ai.Guru, &ai.Mapel, &ai.Kelas, &ai.CreatedAt); err == nil {
				aktivitas = append(aktivitas, ai)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"statistik": map[string]int{
			"total_guru":    totalGuru,
			"total_siswa":   totalSiswa,
			"total_kelas":   totalKelas,
			"total_jurusan": totalJurusan,
			"total_mapel":   totalMapel,
			"total_materi":  totalMateri,
			"total_tugas":   totalTugas,
			"total_kuis":    totalKuis,
			"total_ujian":   totalUjian,
		},
		"jurusan_stats": jurusanStats,
		"aktivitas":     aktivitas,
	})
}

// -------------------------------------------------------------
// 2. Data Terpadu Akademik (Kelas & Jurusan Ringkas) + Wali Kelas
// -------------------------------------------------------------
func getKurikulumAkademik(w http.ResponseWriter, r *http.Request) {
	type KelasRingkas struct {
		ID          int    `json:"id"`
		NamaKelas   string `json:"nama_kelas"`
		Tingkat     string `json:"tingkat"`
		JurusanID   int    `json:"jurusan_id"`
		KodeJurusan string `json:"kode_jurusan"`
		NamaJurusan string `json:"nama_jurusan"`
		WaliKelasID int    `json:"wali_kelas_id"`
		WaliKelas   string `json:"wali_kelas"`
		TotalSiswa  int    `json:"total_siswa"`
	}

	rows, err := config.DB.Query(`
		SELECT k.id, k.nama_kelas, k.tingkat, k.jurusan_id, j.kode_jurusan, j.nama_jurusan,
		       COALESCE(k.wali_kelas_id, 0),
		       COALESCE(g.nama_lengkap, 'Belum Ditentukan') as wali_kelas,
		       (SELECT COUNT(*) FROM siswa s WHERE s.kelas_id = k.id AND s.status = 'aktif') as total_siswa
		FROM kelas k
		JOIN jurusan j ON j.id = k.jurusan_id
		LEFT JOIN guru g ON g.id = k.wali_kelas_id
		ORDER BY k.tingkat, j.kode_jurusan, k.nama_kelas
	`)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}
	defer rows.Close()

	items := []KelasRingkas{}
	for rows.Next() {
		var item KelasRingkas
		if err := rows.Scan(&item.ID, &item.NamaKelas, &item.Tingkat, &item.JurusanID, &item.KodeJurusan, &item.NamaJurusan, &item.WaliKelasID, &item.WaliKelas, &item.TotalSiswa); err == nil {
			items = append(items, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func updateWaliKelas(w http.ResponseWriter, r *http.Request, kelasIDStr string) {
	kelasID, err := strconv.Atoi(kelasIDStr)
	if err != nil || kelasID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID Kelas tidak valid"})
		return
	}

	var req struct {
		WaliKelasID int `json:"wali_kelas_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Data tidak valid"})
		return
	}

	var updateErr error
	if req.WaliKelasID <= 0 {
		_, updateErr = config.DB.Exec("UPDATE kelas SET wali_kelas_id = NULL WHERE id = ?", kelasID)
	} else {
		// Validasi aturan 1 guru 1 walas
		var kelasNama, guruNama string
		errWalas := config.DB.QueryRow(`
			SELECT k.nama_kelas, COALESCE(g.nama_lengkap, 'Guru') 
			FROM kelas k 
			JOIN guru g ON g.id = k.wali_kelas_id 
			WHERE k.wali_kelas_id = ? AND k.id != ?
			LIMIT 1
		`, req.WaliKelasID, kelasID).Scan(&kelasNama, &guruNama)
		if errWalas == nil && kelasNama != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Guru " + guruNama + " sudah menjadi wali kelas di " + kelasNama + ". Satu guru hanya dapat menjadi wali kelas untuk 1 kelas.",
			})
			return
		}

		_, updateErr = config.DB.Exec("UPDATE kelas SET wali_kelas_id = ? WHERE id = ?", req.WaliKelasID, kelasID)
	}

	if updateErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui wali kelas: " + updateErr.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Wali kelas berhasil diperbarui"})
}

// -------------------------------------------------------------
// 3. CRUD Mata Pelajaran & Alokasi Guru Pengampu (guru_mengajar)
// -------------------------------------------------------------
type PengampuItem struct {
	ID            int    `json:"id"`
	GuruID        int    `json:"guru_id"`
	NamaGuru      string `json:"nama_guru"`
	NIP           string `json:"nip"`
	MapelID       int    `json:"mata_pelajaran_id"`
	KodeMapel     string `json:"kode_mapel"`
	NamaMapel     string `json:"nama_mapel"`
	Kelompok      string `json:"kelompok"`
	KelasID       int    `json:"kelas_id"`
	NamaKelas     string `json:"nama_kelas"`
	Tingkat       string `json:"tingkat"`
	TahunAjaranID int    `json:"tahun_ajaran_id"`
	TahunAjaran   string `json:"tahun_ajaran"`
}

func getKurikulumPengampu(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(`
		SELECT gm.id, g.id, g.nama_lengkap, COALESCE(g.nip, '-'),
		       mp.id, mp.kode_mapel, mp.nama_mapel, mp.kelompok,
		       k.id, k.nama_kelas, k.tingkat,
		       ta.id, CONCAT(ta.tahun, ' (', ta.semester, ')') as tahun_ajaran
		FROM guru_mengajar gm
		JOIN guru g ON g.id = gm.guru_id
		JOIN mata_pelajaran mp ON mp.id = gm.mata_pelajaran_id
		JOIN kelas k ON k.id = gm.kelas_id
		JOIN tahun_ajaran ta ON ta.id = gm.tahun_ajaran_id
		ORDER BY mp.kelompok, mp.nama_mapel, k.tingkat, k.nama_kelas
	`)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}
	defer rows.Close()

	items := []PengampuItem{}
	for rows.Next() {
		var item PengampuItem
		if err := rows.Scan(&item.ID, &item.GuruID, &item.NamaGuru, &item.NIP, &item.MapelID, &item.KodeMapel, &item.NamaMapel, &item.Kelompok, &item.KelasID, &item.NamaKelas, &item.Tingkat, &item.TahunAjaranID, &item.TahunAjaran); err == nil {
			items = append(items, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

type PengampuInput struct {
	GuruID        int `json:"guru_id"`
	MapelID       int `json:"mata_pelajaran_id"`
	KelasID       int `json:"kelas_id"`
	TahunAjaranID int `json:"tahun_ajaran_id"`
}

func createKurikulumPengampu(w http.ResponseWriter, r *http.Request) {
	var input PengampuInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format data tidak valid"})
		return
	}

	if input.GuruID <= 0 || input.MapelID <= 0 || input.KelasID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Guru, mata pelajaran, dan kelas wajib dipilih"})
		return
	}

	if input.TahunAjaranID <= 0 {
		_ = config.DB.QueryRow("SELECT id FROM tahun_ajaran WHERE status_aktif = TRUE LIMIT 1").Scan(&input.TahunAjaranID)
		if input.TahunAjaranID <= 0 {
			input.TahunAjaranID = 1
		}
	}

	// Cek duplikasi
	var exists int
	_ = config.DB.QueryRow("SELECT COUNT(*) FROM guru_mengajar WHERE guru_id = ? AND mata_pelajaran_id = ? AND kelas_id = ? AND tahun_ajaran_id = ?",
		input.GuruID, input.MapelID, input.KelasID, input.TahunAjaranID).Scan(&exists)
	if exists > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"message": "Penugasan guru pada mapel dan kelas ini sudah terdaftar"})
		return
	}

	res, err := config.DB.Exec(`
		INSERT INTO guru_mengajar (guru_id, mata_pelajaran_id, kelas_id, tahun_ajaran_id)
		VALUES (?, ?, ?, ?)
	`, input.GuruID, input.MapelID, input.KelasID, input.TahunAjaranID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menyimpan alokasi pengajar: " + err.Error()})
		return
	}

	newID, _ := res.LastInsertId()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "Alokasi guru pengampu berhasil ditambahkan",
		"id":      newID,
	})
}

func updateKurikulumPengampu(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID Alokasi tidak valid"})
		return
	}

	var input PengampuInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format data tidak valid"})
		return
	}

	if input.GuruID <= 0 || input.MapelID <= 0 || input.KelasID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Guru, mata pelajaran, dan kelas wajib dipilih"})
		return
	}

	if input.TahunAjaranID <= 0 {
		_ = config.DB.QueryRow("SELECT id FROM tahun_ajaran WHERE status_aktif = TRUE LIMIT 1").Scan(&input.TahunAjaranID)
		if input.TahunAjaranID <= 0 {
			input.TahunAjaranID = 1
		}
	}

	_, err = config.DB.Exec(`
		UPDATE guru_mengajar
		SET guru_id = ?, mata_pelajaran_id = ?, kelas_id = ?, tahun_ajaran_id = ?
		WHERE id = ?
	`, input.GuruID, input.MapelID, input.KelasID, input.TahunAjaranID, id)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui alokasi pengajar: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Alokasi guru pengampu berhasil diperbarui"})
}

func deleteKurikulumPengampu(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID Alokasi tidak valid"})
		return
	}

	_, err = config.DB.Exec("DELETE FROM guru_mengajar WHERE id = ?", id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menghapus alokasi pengajar: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Alokasi guru pengampu berhasil dihapus"})
}

// -------------------------------------------------------------
// 4. Monitoring Pembelajaran (Materi, Tugas, Kuis & Ujian) + Detail Modals
// -------------------------------------------------------------
func getKurikulumMonitoring(w http.ResponseWriter, r *http.Request) {
	// Materi
	type MateriMonitor struct {
		ID              int    `json:"id"`
		Judul           string `json:"judul"`
		Deskripsi       string `json:"deskripsi"`
		Guru            string `json:"guru"`
		Mapel           string `json:"mapel"`
		Kelas           string `json:"kelas"`
		FileURL         string `json:"file_url"`
		TautanEksternal string `json:"tautan_eksternal"`
		CreatedAt       string `json:"created_at"`
	}
	materiList := []MateriMonitor{}
	mRows, err := config.DB.Query(`
		SELECT m.id, m.judul, COALESCE(m.deskripsi, ''), g.nama_lengkap, mp.nama_mapel, k.nama_kelas,
		       COALESCE(m.file_url, ''), COALESCE(m.tautan_eksternal, ''),
		       DATE_FORMAT(m.created_at, '%Y-%m-%d %H:%i')
		FROM materi m
		JOIN guru g ON g.id = m.guru_id
		JOIN mata_pelajaran mp ON mp.id = m.mata_pelajaran_id
		JOIN kelas k ON k.id = m.kelas_id
		ORDER BY m.id DESC
	`)
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var m MateriMonitor
			if err := mRows.Scan(&m.ID, &m.Judul, &m.Deskripsi, &m.Guru, &m.Mapel, &m.Kelas, &m.FileURL, &m.TautanEksternal, &m.CreatedAt); err == nil {
				materiList = append(materiList, m)
			}
		}
	}

	// Tugas
	type TugasMonitor struct {
		ID             int    `json:"id"`
		Judul          string `json:"judul"`
		Deskripsi      string `json:"deskripsi"`
		Guru           string `json:"guru"`
		Mapel          string `json:"mapel"`
		Kelas          string `json:"kelas"`
		Deadline       string `json:"deadline"`
		TotalTerkumpul int    `json:"total_terkumpul"`
		CreatedAt      string `json:"created_at"`
	}
	tugasList := []TugasMonitor{}
	tRows, err := config.DB.Query(`
		SELECT t.id, t.judul, COALESCE(t.deskripsi, ''), g.nama_lengkap, mp.nama_mapel, k.nama_kelas,
		       DATE_FORMAT(t.deadline, '%Y-%m-%d %H:%i'),
		       (SELECT COUNT(*) FROM pengumpulan_tugas pt WHERE pt.tugas_id = t.id),
		       DATE_FORMAT(t.created_at, '%Y-%m-%d %H:%i')
		FROM tugas t
		JOIN guru g ON g.id = t.guru_id
		JOIN mata_pelajaran mp ON mp.id = t.mata_pelajaran_id
		JOIN kelas k ON k.id = t.kelas_id
		ORDER BY t.id DESC
	`)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var t TugasMonitor
			if err := tRows.Scan(&t.ID, &t.Judul, &t.Deskripsi, &t.Guru, &t.Mapel, &t.Kelas, &t.Deadline, &t.TotalTerkumpul, &t.CreatedAt); err == nil {
				tugasList = append(tugasList, t)
			}
		}
	}

	// Kuis & Ujian
	type EvaluasiMonitor struct {
		ID          int    `json:"id"`
		Judul       string `json:"judul"`
		Deskripsi   string `json:"deskripsi"`
		Tipe        string `json:"tipe"`
		Guru        string `json:"guru"`
		Mapel       string `json:"mapel"`
		Kelas       string `json:"kelas"`
		DurasiMenit int    `json:"durasi_menit"`
		Status      string `json:"status"`
		JumlahSoal  int    `json:"jumlah_soal"`
		CreatedAt   string `json:"created_at"`
	}
	evaluasiList := []EvaluasiMonitor{}
	eRows, err := config.DB.Query(`
		SELECT ku.id, ku.judul, COALESCE(ku.deskripsi, ''), ku.tipe, g.nama_lengkap, mp.nama_mapel, k.nama_kelas,
		       ku.durasi_menit, ku.status,
		       (SELECT COUNT(*) FROM soal_kuis_ujian s WHERE s.kuis_ujian_id = ku.id),
		       DATE_FORMAT(ku.created_at, '%Y-%m-%d %H:%i')
		FROM kuis_ujian ku
		JOIN guru g ON g.id = ku.guru_id
		JOIN mata_pelajaran mp ON mp.id = ku.mata_pelajaran_id
		JOIN kelas k ON k.id = ku.kelas_id
		ORDER BY ku.id DESC
	`)
	if err == nil {
		defer eRows.Close()
		for eRows.Next() {
			var e EvaluasiMonitor
			if err := eRows.Scan(&e.ID, &e.Judul, &e.Deskripsi, &e.Tipe, &e.Guru, &e.Mapel, &e.Kelas, &e.DurasiMenit, &e.Status, &e.JumlahSoal, &e.CreatedAt); err == nil {
				evaluasiList = append(evaluasiList, e)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"materi":   materiList,
		"tugas":    tugasList,
		"evaluasi": evaluasiList,
	})
}

// Detail Butir Soal Kuis / Ujian untuk Modal Preview
func getDetailSoalEvaluasi(w http.ResponseWriter, evaluasiIDStr string) {
	evalID, err := strconv.Atoi(evaluasiIDStr)
	if err != nil || evalID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID evaluasi tidak valid"})
		return
	}

	type PilihanItem struct {
		ID     int    `json:"id"`
		Teks   string `json:"teks"`
		IsBenar bool   `json:"is_benar"`
	}
	type SoalItem struct {
		ID         int           `json:"id"`
		Pertanyaan string        `json:"pertanyaan"`
		Bobot      int           `json:"bobot"`
		Pilihan    []PilihanItem `json:"pilihan"`
	}

	soalList := []SoalItem{}
	sRows, err := config.DB.Query("SELECT id, pertanyaan, bobot FROM soal_kuis_ujian WHERE kuis_ujian_id = ? ORDER BY id", evalID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}
	defer sRows.Close()

	for sRows.Next() {
		var s SoalItem
		if err := sRows.Scan(&s.ID, &s.Pertanyaan, &s.Bobot); err == nil {
			s.Pilihan = []PilihanItem{}
			pRows, pErr := config.DB.Query("SELECT id, pilihan_teks, is_benar FROM pilihan_soal_kuis_ujian WHERE soal_id = ? ORDER BY id", s.ID)
			if pErr == nil {
				for pRows.Next() {
					var p PilihanItem
					if err := pRows.Scan(&p.ID, &p.Teks, &p.IsBenar); err == nil {
						s.Pilihan = append(s.Pilihan, p)
					}
				}
				pRows.Close()
			}
			soalList = append(soalList, s)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(soalList)
}

// Detail Pengumpulan Siswa untuk Modal Preview Tugas
func getDetailPengumpulanTugas(w http.ResponseWriter, tugasIDStr string) {
	tugasID, err := strconv.Atoi(tugasIDStr)
	if err != nil || tugasID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID tugas tidak valid"})
		return
	}

	type PengumpulanItem struct {
		ID           int     `json:"id"`
		NIS          string  `json:"nis"`
		NamaSiswa    string  `json:"nama_siswa"`
		NamaKelas    string  `json:"nama_kelas"`
		FileURL      string  `json:"file_url"`
		WaktuKumpul  string  `json:"waktu_kumpul"`
		Nilai        float64 `json:"nilai"`
		CatatanGuru  string  `json:"catatan_guru"`
	}

	rows, err := config.DB.Query(`
		SELECT pt.id, s.nis, s.nama_lengkap, k.nama_kelas,
		       COALESCE(pt.file_url, ''),
		       DATE_FORMAT(pt.waktu_kumpul, '%Y-%m-%d %H:%i'),
		       COALESCE(pt.nilai, 0),
		       COALESCE(pt.catatan_guru, '')
		FROM pengumpulan_tugas pt
		JOIN siswa s ON s.id = pt.siswa_id
		JOIN kelas k ON k.id = s.kelas_id
		WHERE pt.tugas_id = ?
		ORDER BY pt.waktu_kumpul DESC
	`, tugasID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}
	defer rows.Close()

	items := []PengumpulanItem{}
	for rows.Next() {
		var item PengumpulanItem
		var nilaiVal sql.NullFloat64
		if err := rows.Scan(&item.ID, &item.NIS, &item.NamaSiswa, &item.NamaKelas, &item.FileURL, &item.WaktuKumpul, &nilaiVal, &item.CatatanGuru); err == nil {
			item.Nilai = nilaiVal.Float64
			items = append(items, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// -------------------------------------------------------------
// 5. Rekap Capaian Nilai Siswa Terpadu
// -------------------------------------------------------------
func getKurikulumNilai(w http.ResponseWriter, r *http.Request) {
	kelasFilter := r.URL.Query().Get("kelas_id")
	mapelFilter := r.URL.Query().Get("mapel_id")

	query := `
		SELECT ns.id, s.nis, s.nama_lengkap as nama_siswa, k.nama_kelas, mp.nama_mapel,
		       g.nama_lengkap as nama_guru,
		       COALESCE(ns.nilai_tugas, 0), COALESCE(ns.nilai_kuis, 0), COALESCE(ns.nilai_ujian, 0),
		       COALESCE(ns.nilai_akhir, 0), COALESCE(ns.catatan, '')
		FROM nilai_siswa ns
		JOIN siswa s ON s.id = ns.siswa_id
		JOIN kelas k ON k.id = ns.kelas_id
		JOIN mata_pelajaran mp ON mp.id = ns.mata_pelajaran_id
		JOIN guru g ON g.id = ns.guru_id
		WHERE 1=1
	`
	args := []interface{}{}

	if kelasFilter != "" {
		if kID, err := strconv.Atoi(kelasFilter); err == nil && kID > 0 {
			query += " AND ns.kelas_id = ?"
			args = append(args, kID)
		}
	}
	if mapelFilter != "" {
		if mID, err := strconv.Atoi(mapelFilter); err == nil && mID > 0 {
			query += " AND ns.mata_pelajaran_id = ?"
			args = append(args, mID)
		}
	}

	query += " ORDER BY k.nama_kelas, s.nama_lengkap"

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}
	defer rows.Close()

	type NilaiItem struct {
		ID         int     `json:"id"`
		NIS        string  `json:"nis"`
		NamaSiswa  string  `json:"nama_siswa"`
		NamaKelas  string  `json:"nama_kelas"`
		NamaMapel  string  `json:"nama_mapel"`
		NamaGuru   string  `json:"nama_guru"`
		NilaiTugas float64 `json:"nilai_tugas"`
		NilaiKuis  float64 `json:"nilai_kuis"`
		NilaiUjian float64 `json:"nilai_ujian"`
		NilaiAkhir float64 `json:"nilai_akhir"`
		Predikat   string  `json:"predikat"`
		Catatan    string  `json:"catatan"`
	}

	items := []NilaiItem{}
	for rows.Next() {
		var item NilaiItem
		var nTugas, nKuis, nUjian, nAkhir sql.NullFloat64
		if err := rows.Scan(&item.ID, &item.NIS, &item.NamaSiswa, &item.NamaKelas, &item.NamaMapel, &item.NamaGuru, &nTugas, &nKuis, &nUjian, &nAkhir, &item.Catatan); err == nil {
			item.NilaiTugas = nTugas.Float64
			item.NilaiKuis = nKuis.Float64
			item.NilaiUjian = nUjian.Float64
			item.NilaiAkhir = nAkhir.Float64

			// Predikat
			if item.NilaiAkhir >= 90 {
				item.Predikat = "A (Amat Baik)"
			} else if item.NilaiAkhir >= 80 {
				item.Predikat = "B (Baik)"
			} else if item.NilaiAkhir >= 70 {
				item.Predikat = "C (Cukup)"
			} else if item.NilaiAkhir > 0 {
				item.Predikat = "D (Kurang)"
			} else {
				item.Predikat = "-"
			}

			items = append(items, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
