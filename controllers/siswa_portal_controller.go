package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"backend_lms/config"
)

// HandleSiswaPortal hanya mengembalikan data pembelajaran. Tidak ada nilai,
// skor, jawaban benar, atau umpan balik penilaian di respons mana pun.
func HandleSiswaPortal(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method == http.MethodOptions {
		return
	}
	claims, ok := getClaimsFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Token login diperlukan")
		return
	}
	if claims.Role != "siswa" {
		writeError(w, http.StatusForbidden, "Akses hanya untuk siswa")
		return
	}
	var siswaID, kelasID int
	err := config.DB.QueryRow("SELECT id, kelas_id FROM siswa WHERE user_id=? AND status='aktif'", claims.UserID).Scan(&siswaID, &kelasID)
	if err != nil {
		writeError(w, http.StatusForbidden, "Akun siswa tidak aktif atau profil tidak ditemukan")
		return
	}
	path := strings.TrimPrefix(strings.Trim(r.URL.Path, "/"), "api/siswa/")
	parts := strings.Split(path, "/")
	switch parts[0] {
	case "dashboard":
		if r.Method == http.MethodGet {
			siswaDashboard(w, siswaID, kelasID)
		} else {
			writeError(w, 405, "Metode tidak diizinkan")
		}
	case "materi":
		siswaMateri(w, r, siswaID, kelasID, parts)
	case "tugas":
		siswaTugas(w, r, siswaID, kelasID, parts)
	case "riwayat-tugas":
		if r.Method == http.MethodGet {
			daftarTugas(w, siswaID, kelasID)
		} else {
			writeError(w, 405, "Metode tidak diizinkan")
		}
	case "quiz":
		siswaEvaluasi(w, r, siswaID, kelasID, parts, "kuis")
	case "ujian":
		siswaEvaluasi(w, r, siswaID, kelasID, parts, "ujian")
	case "profil":
		siswaProfil(w, r, siswaID, claims.UserID)
	case "jadwal", "notifikasi":
		json.NewEncoder(w).Encode([]any{}) // modul sumber belum tersedia; tidak membuat data palsu.
	default:
		writeError(w, http.StatusNotFound, "Endpoint siswa tidak ditemukan")
	}
}

func siswaDashboard(w http.ResponseWriter, siswaID, kelasID int) {
	out := map[string]any{}
	var nama, nis, kelas, jurusan string
	if err := config.DB.QueryRow(`SELECT s.nama_lengkap,s.nis,k.nama_kelas,j.nama_jurusan FROM siswa s JOIN kelas k ON k.id=s.kelas_id JOIN jurusan j ON j.id=s.jurusan_id WHERE s.id=?`, siswaID).Scan(&nama, &nis, &kelas, &jurusan); err != nil {
		writeError(w, 500, "Gagal memuat dashboard")
		return
	}
	out["profil"] = map[string]string{"nama_lengkap": nama, "nis": nis, "kelas": kelas, "jurusan": jurusan}
	var belum, sudah, terlambat int
	config.DB.QueryRow(`SELECT COUNT(*) FROM tugas t LEFT JOIN pengumpulan_tugas p ON p.tugas_id=t.id AND p.siswa_id=? WHERE t.kelas_id=? AND p.id IS NULL AND t.deadline>=NOW()`, siswaID, kelasID).Scan(&belum)
	config.DB.QueryRow(`SELECT COUNT(*) FROM pengumpulan_tugas p JOIN tugas t ON t.id=p.tugas_id WHERE p.siswa_id=? AND t.kelas_id=?`, siswaID, kelasID).Scan(&sudah)
	config.DB.QueryRow(`SELECT COUNT(*) FROM pengumpulan_tugas p JOIN tugas t ON t.id=p.tugas_id WHERE p.siswa_id=? AND t.kelas_id=? AND p.waktu_kumpul>t.deadline`, siswaID, kelasID).Scan(&terlambat)
	out["ringkasan_tugas"] = map[string]int{"belum_dikerjakan": belum, "sudah_dikumpulkan": sudah, "terlambat": terlambat}
	rows, err := config.DB.Query(`SELECT t.id,t.judul,mp.nama_mapel,DATE_FORMAT(t.deadline,'%Y-%m-%d %H:%i') FROM tugas t JOIN mata_pelajaran mp ON mp.id=t.mata_pelajaran_id WHERE t.kelas_id=? AND t.deadline>=NOW() ORDER BY t.deadline LIMIT 5`, kelasID)
	if err == nil {
		defer rows.Close()
		dekat := []map[string]any{}
		for rows.Next() {
			var id int
			var judul, mapel, deadline string
			rows.Scan(&id, &judul, &mapel, &deadline)
			dekat = append(dekat, map[string]any{"id": id, "judul": judul, "mata_pelajaran": mapel, "deadline": deadline})
		}
		out["tugas_terdekat"] = dekat
	}
	json.NewEncoder(w).Encode(out)
}

func siswaMateri(w http.ResponseWriter, r *http.Request, _, kelasID int, p []string) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "Metode tidak diizinkan")
		return
	}

	// Helper muat lampiran multi-file
	loadLampiranForMateri := func(materiIDs []int) map[int][]map[string]any {
		res := make(map[int][]map[string]any)
		if len(materiIDs) == 0 {
			return res
		}
		ph := make([]string, len(materiIDs))
		args := make([]interface{}, len(materiIDs))
		for i, id := range materiIDs {
			ph[i] = "?"
			args[i] = id
		}
		rows, err := config.DB.Query(fmt.Sprintf(`SELECT materi_id, file_url, nama_file, ukuran_file, COALESCE(tipe_file, 'dokumen') FROM materi_lampiran WHERE materi_id IN (%s) ORDER BY id ASC`, strings.Join(ph, ",")), args...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var mID int
				var fURL, fName, fType string
				var fSize int64
				if err := rows.Scan(&mID, &fURL, &fName, &fSize, &fType); err == nil {
					res[mID] = append(res[mID], map[string]any{
						"file_url":    fURL,
						"nama_file":   fName,
						"ukuran_file": fSize,
						"tipe_file":   fType,
					})
				}
			}
		}
		return res
	}

	if len(p) > 1 {
		id, err := strconv.Atoi(p[1])
		if err != nil {
			writeError(w, 400, "ID materi tidak valid")
			return
		}
		var judul, deskripsi, file, tautan, mapel, guru, created string
		err = config.DB.QueryRow(`SELECT m.judul,COALESCE(m.deskripsi,''),COALESCE(m.file_url,''),COALESCE(m.tautan_eksternal,''),mp.nama_mapel,g.nama_lengkap,DATE_FORMAT(m.created_at,'%Y-%m-%d %H:%i') FROM materi m JOIN mata_pelajaran mp ON mp.id=m.mata_pelajaran_id JOIN guru g ON g.id=m.guru_id WHERE m.id=? AND m.kelas_id=?`, id, kelasID).Scan(&judul, &deskripsi, &file, &tautan, &mapel, &guru, &created)
		if err == sql.ErrNoRows {
			writeError(w, 404, "Materi tidak ditemukan")
			return
		}
		if err != nil {
			writeError(w, 500, "Gagal memuat materi")
			return
		}
		lampiranList := loadLampiranForMateri([]int{id})[id]
		if len(lampiranList) == 0 && file != "" {
			lampiranList = []map[string]any{{
				"file_url":    file,
				"nama_file":   filepath.Base(file),
				"ukuran_file": 0,
				"tipe_file":   "dokumen",
			}}
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id":               id,
			"judul":            judul,
			"deskripsi":        deskripsi,
			"file_url":         file,
			"tautan_eksternal": tautan,
			"mata_pelajaran":   mapel,
			"guru":             guru,
			"created_at":       created,
			"lampiran":         lampiranList,
		})
		return
	}

	rows, err := config.DB.Query(`SELECT m.id,m.judul,COALESCE(m.deskripsi,''),COALESCE(m.file_url,''),COALESCE(m.tautan_eksternal,''),mp.nama_mapel,g.nama_lengkap,DATE_FORMAT(m.created_at,'%Y-%m-%d %H:%i') FROM materi m JOIN mata_pelajaran mp ON mp.id=m.mata_pelajaran_id JOIN guru g ON g.id=m.guru_id WHERE m.kelas_id=? ORDER BY m.id DESC`, kelasID)
	if err != nil {
		writeError(w, 500, "Gagal memuat materi")
		return
	}
	defer rows.Close()

	out := []map[string]any{}
	materiIDs := []int{}
	for rows.Next() {
		var id int
		var a, b, c, d, e, f, g string
		rows.Scan(&id, &a, &b, &c, &d, &e, &f, &g)
		out = append(out, map[string]any{"id": id, "judul": a, "deskripsi": b, "file_url": c, "tautan_eksternal": d, "mata_pelajaran": e, "guru": f, "created_at": g, "lampiran": []map[string]any{}})
		materiIDs = append(materiIDs, id)
	}

	if len(materiIDs) > 0 {
		lampiranMap := loadLampiranForMateri(materiIDs)
		for i := range out {
			mID := out[i]["id"].(int)
			fURL := out[i]["file_url"].(string)
			if l, ok := lampiranMap[mID]; ok && len(l) > 0 {
				out[i]["lampiran"] = l
			} else if fURL != "" {
				out[i]["lampiran"] = []map[string]any{{
					"file_url":    fURL,
					"nama_file":   filepath.Base(fURL),
					"ukuran_file": 0,
					"tipe_file":   "dokumen",
				}}
			}
		}
	}

	json.NewEncoder(w).Encode(out)
}

func siswaEvaluasi(w http.ResponseWriter, r *http.Request, siswaID, kelasID int, p []string, tipe string) {
	if r.Method == http.MethodGet && len(p) == 1 {
		daftarEvaluasi(w, siswaID, kelasID, tipe)
		return
	}
	if len(p) >= 2 {
		id, err := strconv.Atoi(p[1])
		if err != nil {
			writeError(w, 400, "ID evaluasi tidak valid")
			return
		}
		if r.Method == http.MethodGet {
			detailEvaluasi(w, siswaID, kelasID, id, tipe)
			return
		}
		if r.Method == http.MethodPost && len(p) == 3 && p[2] == "submit" {
			submitEvaluasi(w, r, siswaID, kelasID, id, tipe)
			return
		}
	}
	writeError(w, 405, "Metode tidak diizinkan")
}
func daftarEvaluasi(w http.ResponseWriter, siswaID, kelasID int, tipe string) {
	rows, err := config.DB.Query(`SELECT k.id,k.judul,COALESCE(k.deskripsi,''),mp.nama_mapel,k.durasi_menit,k.status,COALESCE(DATE_FORMAT(k.waktu_mulai,'%Y-%m-%d %H:%i'),''),COALESCE(DATE_FORMAT(k.waktu_selesai,'%Y-%m-%d %H:%i'),''),CASE WHEN p.status='terkirim' THEN 'selesai' ELSE 'tersedia' END FROM kuis_ujian k JOIN mata_pelajaran mp ON mp.id=k.mata_pelajaran_id LEFT JOIN pengerjaan_kuis_ujian p ON p.kuis_ujian_id=k.id AND p.siswa_id=? WHERE k.kelas_id=? AND k.tipe=? AND k.status='aktif' ORDER BY k.id DESC`, siswaID, kelasID, tipe)
	if err != nil {
		writeError(w, 500, "Gagal memuat evaluasi")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, durasi int
		var judul, deskripsi, mapel, status, mulai, selesai, progress string
		rows.Scan(&id, &judul, &deskripsi, &mapel, &durasi, &status, &mulai, &selesai, &progress)
		out = append(out, map[string]any{"id": id, "judul": judul, "deskripsi": deskripsi, "mata_pelajaran": mapel, "durasi_menit": durasi, "status": status, "waktu_mulai": mulai, "waktu_selesai": selesai, "progress": progress})
	}
	json.NewEncoder(w).Encode(out)
}
func detailEvaluasi(w http.ResponseWriter, siswaID, kelasID, id int, tipe string) {
	var judul, deskripsi, mapel string
	var durasi int
	err := config.DB.QueryRow(`SELECT k.judul,COALESCE(k.deskripsi,''),mp.nama_mapel,k.durasi_menit FROM kuis_ujian k JOIN mata_pelajaran mp ON mp.id=k.mata_pelajaran_id WHERE k.id=? AND k.kelas_id=? AND k.tipe=? AND k.status='aktif'`, id, kelasID, tipe).Scan(&judul, &deskripsi, &mapel, &durasi)
	if err != nil {
		writeError(w, 404, "Evaluasi tidak tersedia")
		return
	}
	var submitted int
	config.DB.QueryRow("SELECT COUNT(*) FROM pengerjaan_kuis_ujian WHERE kuis_ujian_id=? AND siswa_id=? AND status='terkirim'", id, siswaID).Scan(&submitted)
	if submitted > 0 {
		writeError(w, 409, "Jawaban sudah dikumpulkan")
		return
	}
	rows, err := config.DB.Query(`SELECT s.id,s.pertanyaan,s.urutan,p.id,p.teks,p.urutan FROM soal_kuis_ujian s JOIN pilihan_soal_kuis_ujian p ON p.soal_id=s.id WHERE s.kuis_ujian_id=? ORDER BY s.urutan,p.urutan`, id)
	if err != nil {
		writeError(w, 500, "Gagal memuat soal")
		return
	}
	defer rows.Close()
	soal := []map[string]any{}
	var last int
	var current map[string]any
	for rows.Next() {
		var sid, pid, urutan, purutan int
		var pertanyaan, teks string
		rows.Scan(&sid, &pertanyaan, &urutan, &pid, &teks, &purutan)
		if sid != last {
			current = map[string]any{"id": sid, "pertanyaan": pertanyaan, "urutan": urutan, "pilihan": []map[string]any{}}
			soal = append(soal, current)
			last = sid
		}
		current["pilihan"] = append(current["pilihan"].([]map[string]any), map[string]any{"id": pid, "teks": teks})
	}
	json.NewEncoder(w).Encode(map[string]any{"id": id, "judul": judul, "deskripsi": deskripsi, "mata_pelajaran": mapel, "durasi_menit": durasi, "soal": soal})
}
func submitEvaluasi(w http.ResponseWriter, r *http.Request, siswaID, kelasID, id int, tipe string) {
	var input struct {
		Jawaban []struct {
			SoalID    int `json:"soal_id"`
			PilihanID int `json:"pilihan_id"`
		} `json:"jawaban"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || len(input.Jawaban) == 0 {
		writeError(w, 400, "Jawaban tidak valid")
		return
	}
	var count int
	config.DB.QueryRow("SELECT COUNT(*) FROM kuis_ujian WHERE id=? AND kelas_id=? AND tipe=? AND status='aktif'", id, kelasID, tipe).Scan(&count)
	if count != 1 {
		writeError(w, 404, "Evaluasi tidak tersedia")
		return
	}
	tx, err := config.DB.Begin()
	if err != nil {
		writeError(w, 500, "Gagal menyimpan jawaban")
		return
	}
	defer tx.Rollback()
	res, err := tx.Exec("INSERT INTO pengerjaan_kuis_ujian(kuis_ujian_id,siswa_id,status,dikirim_pada) VALUES(?,?,'terkirim',NOW())", id, siswaID)
	if err != nil {
		writeError(w, 409, "Jawaban sudah dikumpulkan")
		return
	}
	pengerjaan, _ := res.LastInsertId()
	for _, j := range input.Jawaban {
		var valid int
		tx.QueryRow("SELECT COUNT(*) FROM pilihan_soal_kuis_ujian p JOIN soal_kuis_ujian s ON s.id=p.soal_id WHERE s.id=? AND p.id=? AND s.kuis_ujian_id=?", j.SoalID, j.PilihanID, id).Scan(&valid)
		if valid != 1 {
			writeError(w, 400, "Pilihan jawaban tidak valid")
			return
		}
		if _, err = tx.Exec("INSERT INTO jawaban_kuis_ujian(pengerjaan_id,soal_id,pilihan_id) VALUES(?,?,?)", pengerjaan, j.SoalID, j.PilihanID); err != nil {
			writeError(w, 400, "Jawaban duplikat atau tidak valid")
			return
		}
	}
	if err = tx.Commit(); err != nil {
		writeError(w, 500, "Gagal mengirim jawaban")
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Jawaban berhasil dikumpulkan"})
}
func siswaProfil(w http.ResponseWriter, r *http.Request, siswaID, userID int) {
	if r.Method == http.MethodGet {
		var nama, nis, nisn, username, email, kelas, jurusan, jk, tmpLahir, tglLahir, phone, alamat string
		err := config.DB.QueryRow(`
			SELECT s.nama_lengkap, s.nis, COALESCE(NULLIF(s.nisn, ''), '-'),
			       u.username, u.email,
			       COALESCE(k.nama_kelas, '-'), COALESCE(j.nama_jurusan, '-'),
			       CASE WHEN s.jenis_kelamin = 'P' THEN 'Perempuan' 
			            WHEN s.jenis_kelamin = 'L' THEN 'Laki-laki' 
			            ELSE COALESCE(NULLIF(s.jenis_kelamin, ''), '-') END,
			       COALESCE(NULLIF(s.tempat_lahir, ''), '-'),
			       COALESCE(DATE_FORMAT(s.tanggal_lahir, '%d %M %Y'), '-'),
			       COALESCE(NULLIF(s.no_telepon, ''), '-'),
			       COALESCE(NULLIF(s.alamat, ''), '-')
			FROM siswa s 
			JOIN users u ON u.id = s.user_id 
			LEFT JOIN kelas k ON k.id = s.kelas_id 
			LEFT JOIN jurusan j ON j.id = s.jurusan_id 
			WHERE s.id = ?
		`, siswaID).Scan(&nama, &nis, &nisn, &username, &email, &kelas, &jurusan, &jk, &tmpLahir, &tglLahir, &phone, &alamat)
		if err != nil {
			writeError(w, 500, "Gagal memuat profil: "+err.Error())
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"nama_lengkap":  nama,
			"nis":           nis,
			"nisn":          nisn,
			"username":      username,
			"email":         email,
			"kelas":         kelas,
			"jurusan":       jurusan,
			"jenis_kelamin": jk,
			"tempat_lahir":  tmpLahir,
			"tanggal_lahir": tglLahir,
			"no_telepon":    phone,
			"alamat":        alamat,
		})
		return
	}
	if r.Method == http.MethodPut {
		var in struct {
			NoTelepon string `json:"no_telepon"`
			Alamat    string `json:"alamat"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			writeError(w, 400, "Data profil tidak valid")
			return
		}
		_, err := config.DB.Exec("UPDATE siswa SET no_telepon=NULLIF(?,''),alamat=NULLIF(?, '') WHERE id=? AND user_id=(SELECT id FROM users WHERE id=?)", strings.TrimSpace(in.NoTelepon), strings.TrimSpace(in.Alamat), siswaID, userID)
		if err != nil {
			writeError(w, 500, "Gagal memperbarui profil")
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Profil berhasil diperbarui"})
		return
	}
	writeError(w, 405, "Metode tidak diizinkan")
}

func daftarTugas(w http.ResponseWriter, siswaID, kelasID int) {
	rows, err := config.DB.Query(`SELECT 
		t.id,
		t.judul,
		COALESCE(t.deskripsi, ''),
		mp.nama_mapel,
		g.nama_lengkap,
		DATE_FORMAT(t.created_at,'%Y-%m-%d %H:%i'),
		DATE_FORMAT(t.deadline,'%Y-%m-%d %H:%i'),
		COALESCE(DATE_FORMAT(p.waktu_kumpul,'%Y-%m-%d %H:%i'),''),
		CASE 
			WHEN p.id IS NULL AND t.deadline<NOW() THEN 'terlambat' 
			WHEN p.id IS NULL THEN 'belum_dikerjakan' 
			WHEN p.waktu_kumpul>t.deadline THEN 'terlambat' 
			ELSE 'sudah_dikumpulkan' 
		END,
		COALESCE(t.file_lampiran, ''),
		COALESCE(p.file_url, ''),
		COALESCE(p.catatan_siswa, ''),
		COALESCE(p.status, 'belum_dinilai'),
		COALESCE(p.nilai, 0),
		COALESCE(p.catatan_guru, ''),
		COALESCE(p.id, 0)
	FROM tugas t 
	JOIN mata_pelajaran mp ON mp.id=t.mata_pelajaran_id 
	JOIN guru g ON g.id=t.guru_id 
	LEFT JOIN pengumpulan_tugas p ON p.tugas_id=t.id AND p.siswa_id=? 
	WHERE t.kelas_id=? 
	ORDER BY t.deadline DESC`, siswaID, kelasID)
	if err != nil {
		writeError(w, 500, "Gagal memuat tugas")
		return
	}
	defer rows.Close()

	out := []map[string]any{}
	tugasIDs := []int{}
	pengumpulanIDs := []int{}

	for rows.Next() {
		var id int
		var judul, deskripsi, mapel, guru, tanggalDiberikan, deadline, waktuKumpul, status string
		var fileLampiran, fileUrl, catatanSiswa, statusPenilaian, catatanGuru string
		var nilai float64
		var pID int
		if err := rows.Scan(&id, &judul, &deskripsi, &mapel, &guru, &tanggalDiberikan, &deadline, &waktuKumpul, &status, &fileLampiran, &fileUrl, &catatanSiswa, &statusPenilaian, &nilai, &catatanGuru, &pID); err != nil {
			continue
		}
		tugasIDs = append(tugasIDs, id)
		if pID > 0 {
			pengumpulanIDs = append(pengumpulanIDs, pID)
		}
		out = append(out, map[string]any{
			"id":                 id,
			"pengumpulan_id":     pID,
			"judul":              judul,
			"deskripsi":          deskripsi,
			"mata_pelajaran":     mapel,
			"guru":               guru,
			"tanggal_diberikan":  tanggalDiberikan,
			"deadline":           deadline,
			"waktu_kumpul":       waktuKumpul,
			"status":             status,
			"file_lampiran":      fileLampiran,
			"file_url":           fileUrl,
			"catatan_siswa":      catatanSiswa,
			"status_penilaian":   statusPenilaian,
			"nilai":              nilai,
			"catatan_guru":       catatanGuru,
			"lampiran_guru":      []map[string]any{},
			"berkas_pengumpulan": []map[string]any{},
		})
	}

	// 1. Muat lampiran guru
	if len(tugasIDs) > 0 {
		ph := make([]string, len(tugasIDs))
		args := make([]interface{}, len(tugasIDs))
		for i, id := range tugasIDs {
			ph[i] = "?"
			args[i] = id
		}
		tRows, err := config.DB.Query(fmt.Sprintf(`SELECT tugas_id, file_url, nama_file, ukuran_file, COALESCE(tipe_file, 'dokumen') FROM tugas_lampiran WHERE tugas_id IN (%s) ORDER BY id ASC`, strings.Join(ph, ",")), args...)
		if err == nil {
			defer tRows.Close()
			tLampiranMap := make(map[int][]map[string]any)
			for tRows.Next() {
				var tID int
				var fURL, fName, fType string
				var fSize int64
				if err := tRows.Scan(&tID, &fURL, &fName, &fSize, &fType); err == nil {
					tLampiranMap[tID] = append(tLampiranMap[tID], map[string]any{
						"file_url":    fURL,
						"nama_file":   fName,
						"ukuran_file": fSize,
						"tipe_file":   fType,
					})
				}
			}
			for i := range out {
				tID := out[i]["id"].(int)
				if l, ok := tLampiranMap[tID]; ok && len(l) > 0 {
					out[i]["lampiran_guru"] = l
				} else if fl := out[i]["file_lampiran"].(string); fl != "" {
					out[i]["lampiran_guru"] = []map[string]any{{
						"file_url":    fl,
						"nama_file":   filepath.Base(fl),
						"ukuran_file": 0,
						"tipe_file":   "dokumen",
					}}
				}
			}
		}
	}

	// 2. Muat berkas pengumpulan siswa
	if len(pengumpulanIDs) > 0 {
		ph := make([]string, len(pengumpulanIDs))
		args := make([]interface{}, len(pengumpulanIDs))
		for i, id := range pengumpulanIDs {
			ph[i] = "?"
			args[i] = id
		}
		pRows, err := config.DB.Query(fmt.Sprintf(`SELECT pengumpulan_id, file_url, nama_file, ukuran_file, COALESCE(tipe_file, 'dokumen') FROM pengumpulan_tugas_file WHERE pengumpulan_id IN (%s) ORDER BY id ASC`, strings.Join(ph, ",")), args...)
		if err == nil {
			defer pRows.Close()
			pFilesMap := make(map[int][]map[string]any)
			for pRows.Next() {
				var pID int
				var fURL, fName, fType string
				var fSize int64
				if err := pRows.Scan(&pID, &fURL, &fName, &fSize, &fType); err == nil {
					pFilesMap[pID] = append(pFilesMap[pID], map[string]any{
						"file_url":    fURL,
						"nama_file":   fName,
						"ukuran_file": fSize,
						"tipe_file":   fType,
					})
				}
			}
			for i := range out {
				pID := out[i]["pengumpulan_id"].(int)
				if bf, ok := pFilesMap[pID]; ok && len(bf) > 0 {
					out[i]["berkas_pengumpulan"] = bf
				} else if fu := out[i]["file_url"].(string); fu != "" {
					out[i]["berkas_pengumpulan"] = []map[string]any{{
						"file_url":    fu,
						"nama_file":   filepath.Base(fu),
						"ukuran_file": 0,
						"tipe_file":   "dokumen",
					}}
				}
			}
		}
	}

	json.NewEncoder(w).Encode(out)
}
func siswaTugas(w http.ResponseWriter, r *http.Request, siswaID, kelasID int, p []string) {
	if len(p) == 1 && r.Method == http.MethodGet {
		daftarTugas(w, siswaID, kelasID)
		return
	}
	if len(p) == 3 && p[2] == "pengumpulan" && r.Method == http.MethodPost {
		unggahTugas(w, r, siswaID, kelasID, p[1])
		return
	}
	writeError(w, 405, "Metode tidak diizinkan")
}
func unggahTugas(w http.ResponseWriter, r *http.Request, siswaID, kelasID int, idText string) {
	id, err := strconv.Atoi(idText)
	if err != nil {
		writeError(w, 400, "ID tugas tidak valid")
		return
	}
	var deadline time.Time
	err = config.DB.QueryRow("SELECT deadline FROM tugas WHERE id=? AND kelas_id=?", id, kelasID).Scan(&deadline)
	if err != nil {
		writeError(w, 404, "Tugas tidak ditemukan")
		return
	}
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeError(w, 400, "Ukuran berkas maksimal 50 MB")
		return
	}

	catatanSiswa := strings.TrimSpace(r.FormValue("catatan_siswa"))

	allowed := map[string]string{
		".pdf":  "pdf",
		".doc":  "doc",
		".docx": "docx",
		".xls":  "excel",
		".xlsx": "excel",
		".ppt":  "powerpoint",
		".pptx": "powerpoint",
		".png":  "gambar",
		".jpg":  "gambar",
		".jpeg": "gambar",
		".webp": "gambar",
		".zip":  "zip",
		".rar":  "zip",
	}

	// Kumpulkan seluruh berkas dari field "files" atau "file"
	var allHeaders []*multipart.FileHeader
	if r.MultipartForm != nil {
		if files, ok := r.MultipartForm.File["files"]; ok {
			allHeaders = append(allHeaders, files...)
		}
		if fileSingle, ok := r.MultipartForm.File["file"]; ok && len(allHeaders) == 0 {
			allHeaders = append(allHeaders, fileSingle...)
		}
	}

	dir := filepath.Join("uploads", "tugas", strconv.Itoa(siswaID))
	if err := os.MkdirAll(dir, 0750); err != nil {
		writeError(w, 500, "Gagal menyiapkan penyimpanan")
		return
	}

	type savedFile struct {
		fileURL  string
		fileName string
		fileSize int64
		fileType string
	}
	var savedFiles []savedFile

	for _, header := range allHeaders {
		ext := strings.ToLower(filepath.Ext(header.Filename))
		fileType, ok := allowed[ext]
		if !ok {
			writeError(w, 400, fmt.Sprintf("Format berkas '%s' tidak didukung", header.Filename))
			return
		}
		if header.Size > 50<<20 {
			writeError(w, 400, "Ukuran berkas terlalu besar")
			return
		}

		f, err := header.Open()
		if err != nil {
			continue
		}

		cleanName := filepath.Base(header.Filename)
		uniqueName := fmt.Sprintf("%d_%d%s", id, time.Now().UnixNano(), ext)
		dstPath := filepath.Join(dir, uniqueName)

		dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
		if err != nil {
			f.Close()
			writeError(w, 500, "Gagal menyimpan berkas")
			return
		}

		_, err = io.Copy(dst, f)
		dst.Close()
		f.Close()

		if err != nil {
			writeError(w, 500, "Gagal menyimpan berkas")
			return
		}

		url := "/uploads/tugas/" + strconv.Itoa(siswaID) + "/" + uniqueName
		savedFiles = append(savedFiles, savedFile{
			fileURL:  url,
			fileName: cleanName,
			fileSize: header.Size,
			fileType: fileType,
		})
	}

	tx, err := config.DB.Begin()
	if err != nil {
		writeError(w, 500, "Gagal menyimpan pengumpulan")
		return
	}
	defer tx.Rollback()

	var pengumpulanID int64
	var existingFileURL string
	err = tx.QueryRow("SELECT id, COALESCE(file_url, '') FROM pengumpulan_tugas WHERE tugas_id=? AND siswa_id=?", id, siswaID).Scan(&pengumpulanID, &existingFileURL)

	mainURL := existingFileURL
	if len(savedFiles) > 0 {
		mainURL = savedFiles[0].fileURL
	}

	if err == sql.ErrNoRows {
		if len(savedFiles) == 0 {
			writeError(w, 400, "Berkas tugas wajib diunggah")
			return
		}
		res, e := tx.Exec("INSERT INTO pengumpulan_tugas(tugas_id,siswa_id,file_url,catatan_siswa,status) VALUES(?,?,?,?,'belum_dinilai')", id, siswaID, mainURL, catatanSiswa)
		err = e
		pengumpulanID, _ = res.LastInsertId()
	} else if err == nil {
		_, err = tx.Exec("UPDATE pengumpulan_tugas SET file_url=?,catatan_siswa=?,waktu_kumpul=NOW() WHERE id=?", mainURL, catatanSiswa, pengumpulanID)
	}

	if err != nil {
		writeError(w, 500, "Gagal menyimpan pengumpulan")
		return
	}

	// Simpan berkas-berkas ke pengumpulan_tugas_file
	for _, sf := range savedFiles {
		_, err = tx.Exec(`
			INSERT INTO pengumpulan_tugas_file (pengumpulan_id, file_url, nama_file, ukuran_file, tipe_file)
			VALUES (?, ?, ?, ?, ?)
		`, pengumpulanID, sf.fileURL, sf.fileName, sf.fileSize, sf.fileType)
		if err != nil {
			writeError(w, 500, "Gagal menyimpan berkas pengumpulan")
			return
		}
		// Riwayat pengumpulan
		tx.Exec("INSERT INTO riwayat_pengumpulan_tugas(pengumpulan_id,file_url) VALUES(?,?)", pengumpulanID, sf.fileURL)
	}

	if err = tx.Commit(); err != nil {
		writeError(w, 500, "Gagal menyimpan pengumpulan")
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Tugas berhasil dikumpulkan"})
}
