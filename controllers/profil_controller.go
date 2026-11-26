package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend_lms/config"
	"github.com/golang-jwt/jwt/v5"
)

// HandleGetMyProfile mengembalikan data profil lengkap pengguna yang sedang login (semua role)
func HandleGetMyProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Metode tidak diizinkan"})
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Token otorisasi diperlukan"})
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})

	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Token kedaluwarsa atau tidak valid"})
		return
	}

	userID := claims.UserID

	// Query data dasar users
	var username, email, userRole string
	err = config.DB.QueryRow("SELECT username, email, role FROM users WHERE id = ?", userID).Scan(&username, &email, &userRole)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Data akun pengguna tidak ditemukan"})
		return
	}

	profileData := map[string]interface{}{
		"user_id":  userID,
		"username": username,
		"email":    email,
		"role":     userRole,
		"status":   "aktif",
	}

	switch userRole {
	case "admin":
		var namaLengkap, noTelepon, foto string
		_ = config.DB.QueryRow(`
			SELECT COALESCE(a.nama_lengkap, ''), COALESCE(a.no_telepon, ''), COALESCE(a.foto, '')
			FROM admins a
			RIGHT JOIN users u ON u.id = a.user_id
			WHERE a.user_id = ? OR u.username = ? OR u.email = ?
			LIMIT 1
		`, userID, username, email).Scan(&namaLengkap, &noTelepon, &foto)

		if namaLengkap == "" {
			namaLengkap = username
		}
		profileData["nama_lengkap"] = namaLengkap
		profileData["no_telepon"] = noTelepon
		profileData["foto"] = foto
		profileData["jabatan"] = "Administrator Utama"
		profileData["deskripsi_tugas"] = "Pengelola kendali penuh sistem, master data, dan akun pengguna LMS Edu Class."

	case "guru":
		var guruID int
		var nip, namaLengkap, jenisKelamin, noTelepon, alamat, foto, status string
		err = config.DB.QueryRow(`
			SELECT g.id, COALESCE(NULLIF(g.nip, ''), '-'), g.nama_lengkap, 
			       CASE WHEN g.jenis_kelamin = 'P' THEN 'Perempuan' 
			            WHEN g.jenis_kelamin = 'L' THEN 'Laki-laki' 
			            ELSE COALESCE(NULLIF(g.jenis_kelamin, ''), '-') END,
			       COALESCE(NULLIF(g.no_telepon, ''), '-'), 
			       COALESCE(NULLIF(g.alamat, ''), '-'), 
			       COALESCE(g.foto, ''), 
			       COALESCE(g.status, 'aktif')
			FROM guru g
			LEFT JOIN users u ON u.id = g.user_id
			WHERE g.user_id = ? OR u.username = ? OR u.email = ?
			LIMIT 1
		`, userID, username, email).Scan(&guruID, &nip, &namaLengkap, &jenisKelamin, &noTelepon, &alamat, &foto, &status)

		if err == nil && guruID > 0 {
			_, _ = config.DB.Exec("UPDATE guru SET user_id = ? WHERE id = ? AND (user_id IS NULL OR user_id != ?)", userID, guruID, userID)
		} else {
			namaLengkap = username
		}

		// Ambil mata pelajaran yang diampu
		mapelList := []string{}
		rows, err := config.DB.Query(`
			SELECT mp.nama_mapel
			FROM guru_mata_pelajaran gmp
			JOIN mata_pelajaran mp ON mp.id = gmp.mata_pelajaran_id
			WHERE gmp.guru_id = ?
			ORDER BY mp.nama_mapel ASC
		`, guruID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var mapel string
				if err := rows.Scan(&mapel); err == nil {
					mapelList = append(mapelList, mapel)
				}
			}
		}

		profileData["guru_id"] = guruID
		profileData["nip"] = nip
		profileData["nama_lengkap"] = namaLengkap
		profileData["jenis_kelamin"] = jenisKelamin
		profileData["no_telepon"] = noTelepon
		profileData["alamat"] = alamat
		profileData["foto"] = foto
		profileData["status"] = status
		profileData["mata_pelajaran"] = mapelList
		profileData["jabatan"] = "Tenaga Pendidik / Guru Mata Pelajaran"
		profileData["deskripsi_tugas"] = "Pengajar mata pelajaran, penyusun materi ajar, penugasan, dan evaluasi siswa."

	case "siswa":
		var siswaID int
		var nis, nisn, namaLengkap, jenisKelamin, tempatLahir, tanggalLahir, noTelepon, alamat, foto, status string
		var namaKelas, namaJurusan string

		err = config.DB.QueryRow(`
			SELECT s.id, s.nis, COALESCE(NULLIF(s.nisn, ''), '-'), s.nama_lengkap, 
			       CASE WHEN s.jenis_kelamin = 'P' THEN 'Perempuan' 
			            WHEN s.jenis_kelamin = 'L' THEN 'Laki-laki' 
			            ELSE COALESCE(NULLIF(s.jenis_kelamin, ''), '-') END,
			       COALESCE(NULLIF(s.tempat_lahir, ''), '-'), 
			       COALESCE(DATE_FORMAT(s.tanggal_lahir, '%d %M %Y'), '-'),
			       COALESCE(NULLIF(s.no_telepon, ''), '-'), 
			       COALESCE(NULLIF(s.alamat, ''), '-'), 
			       COALESCE(s.foto, ''), 
			       COALESCE(s.status, 'aktif'),
			       COALESCE(k.nama_kelas, '-'), 
			       COALESCE(j.nama_jurusan, '-')
			FROM siswa s
			LEFT JOIN users u ON u.id = s.user_id
			LEFT JOIN kelas k ON k.id = s.kelas_id
			LEFT JOIN jurusan j ON j.id = s.jurusan_id
			WHERE s.user_id = ? OR u.username = ? OR u.email = ?
			LIMIT 1
		`, userID, username, email).Scan(&siswaID, &nis, &nisn, &namaLengkap, &jenisKelamin, &tempatLahir, &tanggalLahir, &noTelepon, &alamat, &foto, &status, &namaKelas, &namaJurusan)

		if err == nil && siswaID > 0 {
			// Auto-sinkronisasi user_id pada tabel siswa jika ada selisih ID
			_, _ = config.DB.Exec("UPDATE siswa SET user_id = ? WHERE id = ? AND (user_id IS NULL OR user_id != ?)", userID, siswaID, userID)
		} else {
			namaLengkap = username
		}

		profileData["siswa_id"] = siswaID
		profileData["nis"] = nis
		profileData["nisn"] = nisn
		profileData["nama_lengkap"] = namaLengkap
		profileData["jenis_kelamin"] = jenisKelamin
		profileData["tempat_lahir"] = tempatLahir
		profileData["tanggal_lahir"] = tanggalLahir
		profileData["no_telepon"] = noTelepon
		profileData["alamat"] = alamat
		profileData["foto"] = foto
		profileData["status"] = status
		profileData["kelas"] = namaKelas
		profileData["jurusan"] = namaJurusan
		profileData["jabatan"] = "Peserta Didik / Siswa Aktif"
		profileData["deskripsi_tugas"] = "Mengikuti kegiatan belajar mengajar, pengerjaan penugasan, kuis, dan ujian sekolah."

	case "kurikulum":
		var kurikulumID int
		var nip, namaLengkap, noTelepon, foto string

		err = config.DB.QueryRow(`
			SELECT k.id, COALESCE(NULLIF(k.nip, ''), '-'), k.nama_lengkap, COALESCE(NULLIF(k.no_telepon, ''), '-'), COALESCE(k.foto, '')
			FROM kurikulum k
			LEFT JOIN users u ON u.id = k.user_id
			WHERE k.user_id = ? OR u.username = ? OR u.email = ?
			LIMIT 1
		`, userID, username, email).Scan(&kurikulumID, &nip, &namaLengkap, &noTelepon, &foto)

		if err == nil && kurikulumID > 0 {
			_, _ = config.DB.Exec("UPDATE kurikulum SET user_id = ? WHERE id = ? AND (user_id IS NULL OR user_id != ?)", userID, kurikulumID, userID)
		} else {
			namaLengkap = username
		}

		profileData["kurikulum_id"] = kurikulumID
		profileData["nip"] = nip
		profileData["nama_lengkap"] = namaLengkap
		profileData["no_telepon"] = noTelepon
		profileData["foto"] = foto
		profileData["jabatan"] = "Tim Pengembang & Pengawas Kurikulum"
		profileData["deskripsi_tugas"] = "Pengawasan pembelajaran, monitoring kelengkapan materi, kuis, dan analisis rekap nilai siswa."

	case "kepala_sekolah":
		var ksID int
		var nip, namaLengkap, noTelepon, foto string

		err = config.DB.QueryRow(`
			SELECT ks.id, COALESCE(NULLIF(ks.nip, ''), '-'), ks.nama_lengkap, COALESCE(NULLIF(ks.no_telepon, ''), '-'), COALESCE(ks.foto, '')
			FROM kepala_sekolah ks
			LEFT JOIN users u ON u.id = ks.user_id
			WHERE ks.user_id = ? OR u.username = ? OR u.email = ?
			LIMIT 1
		`, userID, username, email).Scan(&ksID, &nip, &namaLengkap, &noTelepon, &foto)

		if err == nil && ksID > 0 {
			_, _ = config.DB.Exec("UPDATE kepala_sekolah SET user_id = ? WHERE id = ? AND (user_id IS NULL OR user_id != ?)", userID, ksID, userID)
		} else {
			namaLengkap = username
		}

		profileData["kepala_sekolah_id"] = ksID
		profileData["nip"] = nip
		profileData["nama_lengkap"] = namaLengkap
		profileData["no_telepon"] = noTelepon
		profileData["foto"] = foto
		profileData["jabatan"] = "Kepala Sekolah (Pimpinan Eksekutif Lembaga)"
		profileData["deskripsi_tugas"] = "Penanggung jawab utama mutu pendidikan, evaluasi tenaga pendidik, dan pelaporan sekolah."

	default:
		profileData["nama_lengkap"] = username
		profileData["jabatan"] = userRole
	}

	json.NewEncoder(w).Encode(profileData)
}
