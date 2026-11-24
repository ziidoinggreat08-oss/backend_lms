package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend_lms/config"
	"golang.org/x/crypto/bcrypt"
)

type Guru struct {
	ID               int      `json:"id"`
	UserID           int      `json:"user_id"`
	NIP              string   `json:"nip"`
	NamaLengkap      string   `json:"nama_lengkap"`
	JenisKelamin     string   `json:"jenis_kelamin"`
	NoTelepon        string   `json:"no_telepon"`
	Alamat           string   `json:"alamat"`
	Status           string   `json:"status"`
	Email            string   `json:"email"`
	Username         string   `json:"username"`
	MataPelajaranIDs []int    `json:"mata_pelajaran_ids"`
	MataPelajaran    []string `json:"mata_pelajaran"`
}

type GuruInput struct {
	NIP              string `json:"nip"`
	NamaLengkap      string `json:"nama_lengkap"`
	JenisKelamin     string `json:"jenis_kelamin"`
	NoTelepon        string `json:"no_telepon"`
	Alamat           string `json:"alamat"`
	Status           string `json:"status"`
	Email            string `json:"email"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	MataPelajaranIDs []int  `json:"mata_pelajaran_ids"`
}

func HandleGuru(w http.ResponseWriter, r *http.Request) {
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
			getGurus(w, r)
		} else {
			getGuruByID(w, id)
		}
	case "POST":
		createGuru(w, r)
	case "PUT":
		updateGuru(w, r, id)
	case "DELETE":
		deleteGuru(w, r, id)
	}
}

func getGurus(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(`
		SELECT g.id, g.user_id, COALESCE(g.nip,''), g.nama_lengkap, g.jenis_kelamin,
		       COALESCE(g.no_telepon,''), COALESCE(g.alamat,''), COALESCE(g.status,'aktif'),
		       COALESCE(u.email,''), COALESCE(u.username,'')
		FROM guru g
		LEFT JOIN users u ON u.id = g.user_id
		ORDER BY g.nama_lengkap
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Map relasi mata pelajaran diampu
	mpMapIDs := make(map[int][]int)
	mpMapNames := make(map[int][]string)

	mpRows, err := config.DB.Query(`
		SELECT gmp.guru_id, mp.id, mp.nama_mapel
		FROM guru_mata_pelajaran gmp
		JOIN mata_pelajaran mp ON mp.id = gmp.mata_pelajaran_id
		ORDER BY mp.nama_mapel
	`)
	if err == nil {
		defer mpRows.Close()
		for mpRows.Next() {
			var gid, mpid int
			var mName string
			if err := mpRows.Scan(&gid, &mpid, &mName); err == nil {
				mpMapIDs[gid] = append(mpMapIDs[gid], mpid)
				mpMapNames[gid] = append(mpMapNames[gid], mName)
			}
		}
	}

	items := []Guru{}
	for rows.Next() {
		var item Guru
		if err := rows.Scan(&item.ID, &item.UserID, &item.NIP, &item.NamaLengkap, &item.JenisKelamin,
			&item.NoTelepon, &item.Alamat, &item.Status, &item.Email, &item.Username); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		item.MataPelajaranIDs = mpMapIDs[item.ID]
		if item.MataPelajaranIDs == nil {
			item.MataPelajaranIDs = []int{}
		}
		item.MataPelajaran = mpMapNames[item.ID]
		if item.MataPelajaran == nil {
			item.MataPelajaran = []string{}
		}

		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func getGuruByID(w http.ResponseWriter, id string) {
	var item Guru
	err := config.DB.QueryRow(`
		SELECT g.id, g.user_id, COALESCE(g.nip,''), g.nama_lengkap, g.jenis_kelamin,
		       COALESCE(g.no_telepon,''), COALESCE(g.alamat,''), COALESCE(g.status,'aktif'),
		       COALESCE(u.email,''), COALESCE(u.username,'')
		FROM guru g
		LEFT JOIN users u ON u.id = g.user_id
		WHERE g.id = ?
	`, id).Scan(&item.ID, &item.UserID, &item.NIP, &item.NamaLengkap, &item.JenisKelamin,
		&item.NoTelepon, &item.Alamat, &item.Status, &item.Email, &item.Username)

	if err != nil {
		http.Error(w, "Guru tidak ditemukan", http.StatusNotFound)
		return
	}

	item.MataPelajaranIDs = []int{}
	item.MataPelajaran = []string{}

	mpRows, err := config.DB.Query(`
		SELECT mp.id, mp.nama_mapel
		FROM guru_mata_pelajaran gmp
		JOIN mata_pelajaran mp ON mp.id = gmp.mata_pelajaran_id
		WHERE gmp.guru_id = ?
		ORDER BY mp.nama_mapel
	`, id)
	if err == nil {
		defer mpRows.Close()
		for mpRows.Next() {
			var mpid int
			var mName string
			if err := mpRows.Scan(&mpid, &mName); err == nil {
				item.MataPelajaranIDs = append(item.MataPelajaranIDs, mpid)
				item.MataPelajaran = append(item.MataPelajaran, mName)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func createGuru(w http.ResponseWriter, r *http.Request) {
	var item GuruInput
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(item.NamaLengkap) == "" || strings.TrimSpace(item.Username) == "" || strings.TrimSpace(item.Email) == "" {
		http.Error(w, "Nama lengkap, username, dan email wajib diisi", http.StatusBadRequest)
		return
	}

	if len(item.MataPelajaranIDs) == 0 {
		http.Error(w, "Minimal satu mata pelajaran yang diampu wajib dipilih", http.StatusBadRequest)
		return
	}

	pwd := strings.TrimSpace(item.Password)
	if pwd == "" {
		pwd = "guru123" // default password jika kosong
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Gagal mengenkripsi password", http.StatusInternalServerError)
		return
	}

	if item.Status == "" {
		item.Status = "aktif"
	}
	if item.JenisKelamin == "" {
		item.JenisKelamin = "L"
	}

	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, "Gagal memulai transaksi: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, 'guru')",
		item.Username, item.Email, string(hash))
	if err != nil {
		http.Error(w, "Username atau email sudah terdaftar: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := res.LastInsertId()

	resGuru, err := tx.Exec("INSERT INTO guru (user_id, nip, nama_lengkap, jenis_kelamin, no_telepon, alamat, status) VALUES (?, NULLIF(?, ''), ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?)",
		userID, item.NIP, item.NamaLengkap, item.JenisKelamin, item.NoTelepon, item.Alamat, item.Status)
	if err != nil {
		http.Error(w, "Gagal menyimpan profil guru: "+err.Error(), http.StatusInternalServerError)
		return
	}

	guruID, _ := resGuru.LastInsertId()

	// Simpan relasi mata pelajaran diampu
	for _, mpID := range item.MataPelajaranIDs {
		if mpID > 0 {
			_, err = tx.Exec("INSERT IGNORE INTO guru_mata_pelajaran (guru_id, mata_pelajaran_id) VALUES (?, ?)", guruID, mpID)
			if err != nil {
				http.Error(w, "Gagal mengaitkan mata pelajaran: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Gagal menyimpan data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Guru berhasil ditambahkan"})
}

func updateGuru(w http.ResponseWriter, r *http.Request, id string) {
	var item GuruInput
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(item.NamaLengkap) == "" || strings.TrimSpace(item.Username) == "" || strings.TrimSpace(item.Email) == "" {
		http.Error(w, "Nama lengkap, username, dan email wajib diisi", http.StatusBadRequest)
		return
	}

	if len(item.MataPelajaranIDs) == 0 {
		http.Error(w, "Minimal satu mata pelajaran yang diampu wajib dipilih", http.StatusBadRequest)
		return
	}

	var userID int
	err := config.DB.QueryRow("SELECT user_id FROM guru WHERE id = ?", id).Scan(&userID)
	if err != nil {
		http.Error(w, "Guru tidak ditemukan", http.StatusNotFound)
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, "Gagal memulai transaksi", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update profil guru
	_, err = tx.Exec("UPDATE guru SET nip=NULLIF(?, ''), nama_lengkap=?, jenis_kelamin=?, no_telepon=NULLIF(?, ''), alamat=NULLIF(?, ''), status=? WHERE id=?",
		item.NIP, item.NamaLengkap, item.JenisKelamin, item.NoTelepon, item.Alamat, item.Status, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update user info
	_, err = tx.Exec("UPDATE users SET email=?, username=? WHERE id=?", item.Email, item.Username, userID)
	if err != nil {
		http.Error(w, "Username atau email sudah digunakan: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Jika password diisi, update password
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

	// Sinkronisasi mata pelajaran diampu
	_, err = tx.Exec("DELETE FROM guru_mata_pelajaran WHERE guru_id = ?", id)
	if err != nil {
		http.Error(w, "Gagal memperbarui relasi mata pelajaran: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, mpID := range item.MataPelajaranIDs {
		if mpID > 0 {
			_, err = tx.Exec("INSERT IGNORE INTO guru_mata_pelajaran (guru_id, mata_pelajaran_id) VALUES (?, ?)", id, mpID)
			if err != nil {
				http.Error(w, "Gagal menyimpan relasi mata pelajaran: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Gagal memperbarui data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Guru berhasil diubah"})
}

func deleteGuru(w http.ResponseWriter, r *http.Request, id string) {
	var userID int
	err := config.DB.QueryRow("SELECT user_id FROM guru WHERE id=?", id).Scan(&userID)
	if err != nil {
		http.Error(w, "Guru tidak ditemukan", http.StatusNotFound)
		return
	}

	// Menghapus user akan cascade ke tabel guru dan tabel guru_mata_pelajaran (FOREIGN KEY ... ON DELETE CASCADE)
	_, err = config.DB.Exec("DELETE FROM users WHERE id=?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Guru berhasil dihapus"})
}
