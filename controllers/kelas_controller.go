package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"backend_lms/config"
)

type Kelas struct {
	ID            int    `json:"id"`
	NamaKelas     string `json:"nama_kelas"`
	Tingkat       string `json:"tingkat"`
	JurusanID     int    `json:"jurusan_id"`
	NamaJurusan   string `json:"nama_jurusan"`
	KodeJurusan   string `json:"kode_jurusan"`
	TahunAjaranID int    `json:"tahun_ajaran_id"`
	TahunAjaran   string `json:"tahun_ajaran"`
	WaliKelasID   *int   `json:"wali_kelas_id"`
	WaliKelasNama string `json:"wali_kelas_nama"`
	JumlahSiswa   int    `json:"jumlah_siswa"`
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
		} else {
			getKelasByID(w, id)
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
	jurusanIDQuery := r.URL.Query().Get("jurusan_id")
	tingkatQuery := r.URL.Query().Get("tingkat")

	baseQuery := `
		SELECT k.id, k.nama_kelas, k.tingkat, k.jurusan_id, 
		       COALESCE(j.nama_jurusan, ''), COALESCE(j.kode_jurusan, ''),
		       k.tahun_ajaran_id,
		       COALESCE(CONCAT(ta.tahun, ' - ', ta.semester), ''),
		       k.wali_kelas_id,
		       COALESCE(g.nama_lengkap, ''),
		       COUNT(DISTINCT s.id) as jumlah_siswa
		FROM kelas k
		LEFT JOIN jurusan j ON j.id = k.jurusan_id
		LEFT JOIN tahun_ajaran ta ON ta.id = k.tahun_ajaran_id
		LEFT JOIN guru g ON g.id = k.wali_kelas_id
		LEFT JOIN siswa s ON s.kelas_id = k.id AND s.status = 'aktif'
	`

	var conditions []string
	var args []interface{}

	if jurusanIDQuery != "" {
		if jID, err := strconv.Atoi(jurusanIDQuery); err == nil && jID > 0 {
			conditions = append(conditions, "k.jurusan_id = ?")
			args = append(args, jID)
		}
	}

	if tingkatQuery != "" {
		conditions = append(conditions, "k.tingkat = ?")
		args = append(args, tingkatQuery)
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += " GROUP BY k.id, k.nama_kelas, k.tingkat, k.jurusan_id, j.nama_jurusan, j.kode_jurusan, k.tahun_ajaran_id, ta.tahun, ta.semester, k.wali_kelas_id, g.nama_lengkap ORDER BY k.jurusan_id ASC, k.tingkat ASC, k.nama_kelas ASC"

	rows, err := config.DB.Query(baseQuery, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []Kelas{}
	for rows.Next() {
		var item Kelas
		var waliKelasID sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.NamaKelas,
			&item.Tingkat,
			&item.JurusanID,
			&item.NamaJurusan,
			&item.KodeJurusan,
			&item.TahunAjaranID,
			&item.TahunAjaran,
			&waliKelasID,
			&item.WaliKelasNama,
			&item.JumlahSiswa,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if waliKelasID.Valid {
			idVal := int(waliKelasID.Int64)
			item.WaliKelasID = &idVal
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func getKelasByID(w http.ResponseWriter, id string) {
	var item Kelas
	var waliKelasID sql.NullInt64
	err := config.DB.QueryRow(`
		SELECT k.id, k.nama_kelas, k.tingkat, k.jurusan_id, 
		       COALESCE(j.nama_jurusan, ''), COALESCE(j.kode_jurusan, ''),
		       k.tahun_ajaran_id,
		       COALESCE(CONCAT(ta.tahun, ' - ', ta.semester), ''),
		       k.wali_kelas_id,
		       COALESCE(g.nama_lengkap, ''),
		       COUNT(DISTINCT s.id) as jumlah_siswa
		FROM kelas k
		LEFT JOIN jurusan j ON j.id = k.jurusan_id
		LEFT JOIN tahun_ajaran ta ON ta.id = k.tahun_ajaran_id
		LEFT JOIN guru g ON g.id = k.wali_kelas_id
		LEFT JOIN siswa s ON s.kelas_id = k.id AND s.status = 'aktif'
		WHERE k.id = ?
		GROUP BY k.id, k.nama_kelas, k.tingkat, k.jurusan_id, j.nama_jurusan, j.kode_jurusan, k.tahun_ajaran_id, ta.tahun, ta.semester, k.wali_kelas_id, g.nama_lengkap
	`, id).Scan(
		&item.ID,
		&item.NamaKelas,
		&item.Tingkat,
		&item.JurusanID,
		&item.NamaJurusan,
		&item.KodeJurusan,
		&item.TahunAjaranID,
		&item.TahunAjaran,
		&waliKelasID,
		&item.WaliKelasNama,
		&item.JumlahSiswa,
	)

	if err != nil {
		http.Error(w, "Kelas tidak ditemukan", http.StatusNotFound)
		return
	}

	if waliKelasID.Valid {
		idVal := int(waliKelasID.Int64)
		item.WaliKelasID = &idVal
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func createKelas(w http.ResponseWriter, r *http.Request) {
	var item struct {
		NamaKelas     string `json:"nama_kelas"`
		Tingkat       string `json:"tingkat"`
		JurusanID     int    `json:"jurusan_id"`
		TahunAjaranID int    `json:"tahun_ajaran_id"`
		WaliKelasID   *int   `json:"wali_kelas_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item.NamaKelas = strings.TrimSpace(item.NamaKelas)
	if item.NamaKelas == "" {
		http.Error(w, "Nama kelas wajib diisi", http.StatusBadRequest)
		return
	}

	// Cegah duplikasi nama kelas / nomor rombel
	var exists int
	errCheck := config.DB.QueryRow("SELECT COUNT(*) FROM kelas WHERE LOWER(TRIM(nama_kelas)) = LOWER(?)", item.NamaKelas).Scan(&exists)
	if errCheck == nil && exists > 0 {
		http.Error(w, "Kelas '"+item.NamaKelas+"' sudah terdaftar. Nomor rombel tidak boleh duplikat.", http.StatusBadRequest)
		return
	}

	if item.TahunAjaranID <= 0 {
		item.TahunAjaranID = 1
	}

	// Validasi aturan 1 guru 1 walas: Cegah guru merangkap menjadi walas di kelas lain
	if item.WaliKelasID != nil && *item.WaliKelasID > 0 {
		var kelasNama, guruNama string
		errWalas := config.DB.QueryRow(`
			SELECT k.nama_kelas, COALESCE(g.nama_lengkap, 'Guru') 
			FROM kelas k 
			JOIN guru g ON g.id = k.wali_kelas_id 
			WHERE k.wali_kelas_id = ? 
			LIMIT 1
		`, *item.WaliKelasID).Scan(&kelasNama, &guruNama)
		if errWalas == nil && kelasNama != "" {
			http.Error(w, "Guru "+guruNama+" sudah ditugaskan sebagai wali kelas di kelas "+kelasNama+". Satu guru hanya dapat menjadi wali kelas untuk 1 kelas.", http.StatusBadRequest)
			return
		}
	}

	var err error
	if item.WaliKelasID != nil && *item.WaliKelasID > 0 {
		_, err = config.DB.Exec("INSERT INTO kelas (nama_kelas, tingkat, jurusan_id, tahun_ajaran_id, wali_kelas_id) VALUES (?, ?, ?, ?, ?)",
			item.NamaKelas, item.Tingkat, item.JurusanID, item.TahunAjaranID, *item.WaliKelasID)
	} else {
		_, err = config.DB.Exec("INSERT INTO kelas (nama_kelas, tingkat, jurusan_id, tahun_ajaran_id) VALUES (?, ?, ?, ?)",
			item.NamaKelas, item.Tingkat, item.JurusanID, item.TahunAjaranID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kelas berhasil ditambahkan"})
}

func updateKelas(w http.ResponseWriter, r *http.Request, id string) {
	var item struct {
		NamaKelas     string `json:"nama_kelas"`
		Tingkat       string `json:"tingkat"`
		JurusanID     int    `json:"jurusan_id"`
		TahunAjaranID int    `json:"tahun_ajaran_id"`
		WaliKelasID   *int   `json:"wali_kelas_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item.NamaKelas = strings.TrimSpace(item.NamaKelas)
	if item.NamaKelas == "" {
		http.Error(w, "Nama kelas wajib diisi", http.StatusBadRequest)
		return
	}

	// Cegah duplikasi saat edit ke nama kelas lain yang sudah dipakai
	var exists int
	errCheck := config.DB.QueryRow("SELECT COUNT(*) FROM kelas WHERE LOWER(TRIM(nama_kelas)) = LOWER(?) AND id != ?", item.NamaKelas, id).Scan(&exists)
	if errCheck == nil && exists > 0 {
		http.Error(w, "Kelas '"+item.NamaKelas+"' sudah digunakan oleh kelas lain. Nomor rombel tidak boleh duplikat.", http.StatusBadRequest)
		return
	}

	if item.TahunAjaranID <= 0 {
		item.TahunAjaranID = 1
	}

	// Validasi aturan 1 guru 1 walas: Cegah guru merangkap menjadi walas di kelas lain (kecuali kelas ini sendiri)
	if item.WaliKelasID != nil && *item.WaliKelasID > 0 {
		var kelasNama, guruNama string
		errWalas := config.DB.QueryRow(`
			SELECT k.nama_kelas, COALESCE(g.nama_lengkap, 'Guru') 
			FROM kelas k 
			JOIN guru g ON g.id = k.wali_kelas_id 
			WHERE k.wali_kelas_id = ? AND k.id != ?
			LIMIT 1
		`, *item.WaliKelasID, id).Scan(&kelasNama, &guruNama)
		if errWalas == nil && kelasNama != "" {
			http.Error(w, "Guru "+guruNama+" sudah ditugaskan sebagai wali kelas di kelas "+kelasNama+". Satu guru hanya dapat menjadi wali kelas untuk 1 kelas.", http.StatusBadRequest)
			return
		}
	}

	var err error
	if item.WaliKelasID != nil && *item.WaliKelasID > 0 {
		_, err = config.DB.Exec("UPDATE kelas SET nama_kelas=?, tingkat=?, jurusan_id=?, tahun_ajaran_id=?, wali_kelas_id=? WHERE id=?",
			item.NamaKelas, item.Tingkat, item.JurusanID, item.TahunAjaranID, *item.WaliKelasID, id)
	} else {
		_, err = config.DB.Exec("UPDATE kelas SET nama_kelas=?, tingkat=?, jurusan_id=?, tahun_ajaran_id=?, wali_kelas_id=NULL WHERE id=?",
			item.NamaKelas, item.Tingkat, item.JurusanID, item.TahunAjaranID, id)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kelas berhasil diubah"})
}

func deleteKelas(w http.ResponseWriter, r *http.Request, id string) {
	var count int
	_ = config.DB.QueryRow("SELECT COUNT(*) FROM siswa WHERE kelas_id = ?", id).Scan(&count)
	if count > 0 {
		http.Error(w, "Tidak dapat menghapus kelas karena masih memiliki siswa terdaftar", http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("DELETE FROM kelas WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Kelas berhasil dihapus"})
}

func ListKelas(w http.ResponseWriter, r *http.Request) {
	getKelas(w, r)
}
