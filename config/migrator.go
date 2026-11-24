package config

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// EnsureDatabaseExists memastikan basis data MySQL sudah dibuat sebelum aplikasi terhubung ke DB_NAME.
func EnsureDatabaseExists(host, port, user, password, dbName string) error {
	// Koneksi ke server MySQL tanpa memilih nama database spesifik
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true", user, password, host, port)
	serverDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("gagal membuka koneksi ke server MySQL: %w", err)
	}
	defer serverDB.Close()

	if err := serverDB.Ping(); err != nil {
		return fmt.Errorf("gagal terhubung ke MySQL server di %s:%s. Pastikan MySQL sudah berjalan: %w", host, port, err)
	}

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)
	if _, err := serverDB.Exec(query); err != nil {
		return fmt.Errorf("gagal memastikan database '%s' ada: %w", dbName, err)
	}

	return nil
}

type migrationItem struct {
	version     string
	description string
	queries     []string
}

// RunAutoMigrations menjalankan migrasi database secara berurutan dan idempotent.
// Riwayat eksekusi disimpan di tabel schema_migrations sehingga aman dijalankan berulang kali.
func RunAutoMigrations(db *sql.DB) error {
	// 1. Inisialisasi tabel tracking migrasi
	initTableQuery := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(100) PRIMARY KEY,
		description VARCHAR(255) NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB;`
	if _, err := db.Exec(initTableQuery); err != nil {
		return fmt.Errorf("gagal menginisialisasi tabel schema_migrations: %w", err)
	}

	// 2. Baca daftar migrasi yang sudah pernah diterapkan
	applied := make(map[string]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var v string
			if err := rows.Scan(&v); err == nil {
				applied[v] = true
			}
		}
	}

	// 3. Daftar seluruh urutan migrasi skema dan seed
	migrations := getMigrationList()

	appliedCount := 0
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}

		// Jalankan setiap statement dalam migrasi
		for _, q := range m.queries {
			trimmed := strings.TrimSpace(q)
			if trimmed == "" {
				continue
			}
			if _, err := db.Exec(trimmed); err != nil {
				return fmt.Errorf("gagal menjalankan migrasi '%s' pada query [%s]: %w", m.version, trimmed, err)
			}
		}

		// Catat riwayat migrasi yang sukses
		_, err := db.Exec("INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, ?)",
			m.version, m.description, time.Now())
		if err != nil {
			return fmt.Errorf("gagal mencatat riwayat migrasi '%s': %w", m.version, err)
		}

		log.Printf("[Migration] Berhasil menerapkan: %s (%s)", m.version, m.description)
		appliedCount++
	}

	if appliedCount == 0 {
		log.Printf("[Migration] Skema database sudah mutakhir (%d migrasi terverifikasi).", len(migrations))
	} else {
		log.Printf("[Migration] Total %d migrasi baru berhasil diterapkan.", appliedCount)
	}

	return nil
}

func getMigrationList() []migrationItem {
	return []migrationItem{
		// -------------------------------------------------------------------------
		// MIGRASI 1: Core Master Schema
		// -------------------------------------------------------------------------
		{
			version:     "20261123_01_core_schema",
			description: "Tabel Master: Users, Admins, Tahun Ajaran, Jurusan, Mapel, Guru, Kelas, Siswa, Kurikulum, Kepsek, Guru Mengajar",
			queries: []string{
				`CREATE TABLE IF NOT EXISTS users (
					id INT AUTO_INCREMENT PRIMARY KEY,
					username VARCHAR(50) UNIQUE NOT NULL,
					email VARCHAR(100) UNIQUE NOT NULL,
					password VARCHAR(255) NOT NULL,
					role ENUM('admin', 'guru', 'siswa', 'kepala_sekolah', 'kurikulum') NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS admins (
					id INT AUTO_INCREMENT PRIMARY KEY,
					user_id INT UNIQUE NOT NULL,
					nama_lengkap VARCHAR(100) NOT NULL,
					no_telepon VARCHAR(20) NULL,
					foto VARCHAR(255) NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS tahun_ajaran (
					id INT AUTO_INCREMENT PRIMARY KEY,
					tahun VARCHAR(20) NOT NULL,
					semester ENUM('ganjil', 'genap') NOT NULL,
					status_aktif BOOLEAN DEFAULT FALSE,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS jurusan (
					id INT AUTO_INCREMENT PRIMARY KEY,
					kode_jurusan VARCHAR(20) UNIQUE NOT NULL,
					nama_jurusan VARCHAR(100) NOT NULL,
					deskripsi TEXT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS mata_pelajaran (
					id INT AUTO_INCREMENT PRIMARY KEY,
					kode_mapel VARCHAR(20) UNIQUE NOT NULL,
					nama_mapel VARCHAR(100) NOT NULL,
					kelompok ENUM('umum', 'kejuruan', 'muatan_lokal') DEFAULT 'kejuruan',
					deskripsi TEXT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS guru (
					id INT AUTO_INCREMENT PRIMARY KEY,
					user_id INT UNIQUE NOT NULL,
					nip VARCHAR(30) UNIQUE NULL,
					nama_lengkap VARCHAR(100) NOT NULL,
					jenis_kelamin ENUM('L', 'P') NOT NULL,
					no_telepon VARCHAR(20) NULL,
					alamat TEXT NULL,
					foto VARCHAR(255) NULL,
					status ENUM('aktif', 'nonaktif') DEFAULT 'aktif',
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS kelas (
					id INT AUTO_INCREMENT PRIMARY KEY,
					nama_kelas VARCHAR(50) NOT NULL,
					tingkat ENUM('X', 'XI', 'XII') NOT NULL,
					jurusan_id INT NOT NULL,
					tahun_ajaran_id INT NOT NULL,
					wali_kelas_id INT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					FOREIGN KEY (jurusan_id) REFERENCES jurusan(id) ON DELETE RESTRICT,
					FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id) ON DELETE RESTRICT,
					FOREIGN KEY (wali_kelas_id) REFERENCES guru(id) ON DELETE SET NULL
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS siswa (
					id INT AUTO_INCREMENT PRIMARY KEY,
					user_id INT UNIQUE NOT NULL,
					nis VARCHAR(30) UNIQUE NOT NULL,
					nisn VARCHAR(30) UNIQUE NULL,
					nama_lengkap VARCHAR(100) NOT NULL,
					jenis_kelamin ENUM('L', 'P') NOT NULL,
					tempat_lahir VARCHAR(50) NULL,
					tanggal_lahir DATE NULL,
					no_telepon VARCHAR(20) NULL,
					alamat TEXT NULL,
					foto VARCHAR(255) NULL,
					kelas_id INT NOT NULL,
					jurusan_id INT NOT NULL,
					status ENUM('aktif', 'nonaktif', 'lulus', 'keluar') DEFAULT 'aktif',
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
					FOREIGN KEY (kelas_id) REFERENCES kelas(id) ON DELETE RESTRICT,
					FOREIGN KEY (jurusan_id) REFERENCES jurusan(id) ON DELETE RESTRICT
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS guru_mengajar (
					id INT AUTO_INCREMENT PRIMARY KEY,
					guru_id INT NOT NULL,
					mata_pelajaran_id INT NOT NULL,
					kelas_id INT NOT NULL,
					tahun_ajaran_id INT NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (guru_id) REFERENCES guru(id) ON DELETE CASCADE,
					FOREIGN KEY (mata_pelajaran_id) REFERENCES mata_pelajaran(id) ON DELETE RESTRICT,
					FOREIGN KEY (kelas_id) REFERENCES kelas(id) ON DELETE RESTRICT,
					FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id) ON DELETE RESTRICT
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS kepala_sekolah (
					id INT AUTO_INCREMENT PRIMARY KEY,
					user_id INT UNIQUE NOT NULL,
					nip VARCHAR(30) UNIQUE NULL,
					nama_lengkap VARCHAR(100) NOT NULL,
					no_telepon VARCHAR(20) NULL,
					foto VARCHAR(255) NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS kurikulum (
					id INT AUTO_INCREMENT PRIMARY KEY,
					user_id INT UNIQUE NOT NULL,
					nip VARCHAR(30) UNIQUE NULL,
					nama_lengkap VARCHAR(100) NOT NULL,
					no_telepon VARCHAR(20) NULL,
					foto VARCHAR(255) NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,
			},
		},

		// -------------------------------------------------------------------------
		// MIGRASI 2: Skema Pembelajaran Guru & Nilai
		// -------------------------------------------------------------------------
		{
			version:     "20261123_02_learning_schema",
			description: "Tabel Pembelajaran: Materi, Tugas, Pengumpulan Tugas, Kuis Ujian, Nilai Siswa",
			queries: []string{
				`CREATE TABLE IF NOT EXISTS materi (
					id INT AUTO_INCREMENT PRIMARY KEY,
					guru_id INT NOT NULL,
					mata_pelajaran_id INT NOT NULL,
					kelas_id INT NOT NULL,
					judul VARCHAR(150) NOT NULL,
					deskripsi TEXT NULL,
					file_url VARCHAR(255) NULL,
					tautan_eksternal VARCHAR(255) NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					FOREIGN KEY (guru_id) REFERENCES guru(id) ON DELETE CASCADE,
					FOREIGN KEY (mata_pelajaran_id) REFERENCES mata_pelajaran(id) ON DELETE CASCADE,
					FOREIGN KEY (kelas_id) REFERENCES kelas(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS tugas (
					id INT AUTO_INCREMENT PRIMARY KEY,
					guru_id INT NOT NULL,
					mata_pelajaran_id INT NOT NULL,
					kelas_id INT NOT NULL,
					judul VARCHAR(150) NOT NULL,
					deskripsi TEXT NOT NULL,
					deadline DATETIME NOT NULL,
					file_lampiran VARCHAR(255) NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					FOREIGN KEY (guru_id) REFERENCES guru(id) ON DELETE CASCADE,
					FOREIGN KEY (mata_pelajaran_id) REFERENCES mata_pelajaran(id) ON DELETE CASCADE,
					FOREIGN KEY (kelas_id) REFERENCES kelas(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS pengumpulan_tugas (
					id INT AUTO_INCREMENT PRIMARY KEY,
					tugas_id INT NOT NULL,
					siswa_id INT NOT NULL,
					file_url VARCHAR(255) NULL,
					catatan_siswa TEXT NULL,
					waktu_kumpul TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					nilai DECIMAL(5,2) NULL,
					catatan_guru TEXT NULL,
					status ENUM('belum_dinilai', 'sudah_dinilai') DEFAULT 'belum_dinilai',
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					UNIQUE KEY unik_tugas_siswa (tugas_id, siswa_id),
					FOREIGN KEY (tugas_id) REFERENCES tugas(id) ON DELETE CASCADE,
					FOREIGN KEY (siswa_id) REFERENCES siswa(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS kuis_ujian (
					id INT AUTO_INCREMENT PRIMARY KEY,
					guru_id INT NOT NULL,
					mata_pelajaran_id INT NOT NULL,
					kelas_id INT NOT NULL,
					judul VARCHAR(150) NOT NULL,
					tipe ENUM('kuis', 'ujian') NOT NULL DEFAULT 'kuis',
					durasi_menit INT NOT NULL DEFAULT 60,
					waktu_mulai DATETIME NULL,
					waktu_selesai DATETIME NULL,
					deskripsi TEXT NULL,
					status ENUM('draft', 'aktif', 'selesai') DEFAULT 'draft',
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					FOREIGN KEY (guru_id) REFERENCES guru(id) ON DELETE CASCADE,
					FOREIGN KEY (mata_pelajaran_id) REFERENCES mata_pelajaran(id) ON DELETE CASCADE,
					FOREIGN KEY (kelas_id) REFERENCES kelas(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS nilai_siswa (
					id INT AUTO_INCREMENT PRIMARY KEY,
					siswa_id INT NOT NULL,
					mata_pelajaran_id INT NOT NULL,
					kelas_id INT NOT NULL,
					guru_id INT NOT NULL,
					nilai_tugas DECIMAL(5,2) DEFAULT 0,
					nilai_kuis DECIMAL(5,2) DEFAULT 0,
					nilai_ujian DECIMAL(5,2) DEFAULT 0,
					nilai_akhir DECIMAL(5,2) DEFAULT 0,
					catatan TEXT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					UNIQUE KEY unik_nilai_siswa_mapel (siswa_id, mata_pelajaran_id),
					FOREIGN KEY (siswa_id) REFERENCES siswa(id) ON DELETE CASCADE,
					FOREIGN KEY (mata_pelajaran_id) REFERENCES mata_pelajaran(id) ON DELETE CASCADE,
					FOREIGN KEY (kelas_id) REFERENCES kelas(id) ON DELETE CASCADE,
					FOREIGN KEY (guru_id) REFERENCES guru(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,
			},
		},

		// -------------------------------------------------------------------------
		// MIGRASI 3: Butir Soal & Pilihan Evaluasi
		// -------------------------------------------------------------------------
		{
			version:     "20261123_03_evaluasi_questions",
			description: "Tabel Butir Soal & Pilihan Jawaban Kuis/Ujian",
			queries: []string{
				`CREATE TABLE IF NOT EXISTS soal_kuis_ujian (
					id INT AUTO_INCREMENT PRIMARY KEY,
					kuis_ujian_id INT NOT NULL,
					pertanyaan TEXT NOT NULL,
					bobot INT NOT NULL DEFAULT 1,
					urutan INT NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (kuis_ujian_id) REFERENCES kuis_ujian(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS pilihan_soal_kuis_ujian (
					id INT AUTO_INCREMENT PRIMARY KEY,
					soal_id INT NOT NULL,
					teks TEXT NOT NULL,
					is_benar BOOLEAN NOT NULL DEFAULT FALSE,
					urutan INT NOT NULL,
					FOREIGN KEY (soal_id) REFERENCES soal_kuis_ujian(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,
			},
		},

		// -------------------------------------------------------------------------
		// MIGRASI 4: Pengerjaan Kuis Siswa & Riwayat Tugas
		// -------------------------------------------------------------------------
		{
			version:     "20261123_04_siswa_evaluasi_history",
			description: "Tabel Riwayat Tugas, Sesi Pengerjaan Kuis Siswa, dan Jawaban Kuis",
			queries: []string{
				`CREATE TABLE IF NOT EXISTS riwayat_pengumpulan_tugas (
					id INT AUTO_INCREMENT PRIMARY KEY,
					pengumpulan_id INT NOT NULL,
					file_url VARCHAR(255) NOT NULL,
					waktu_kumpul TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (pengumpulan_id) REFERENCES pengumpulan_tugas(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS pengerjaan_kuis_ujian (
					id INT AUTO_INCREMENT PRIMARY KEY,
					kuis_ujian_id INT NOT NULL,
					siswa_id INT NOT NULL,
					status ENUM('dikerjakan', 'terkirim') NOT NULL DEFAULT 'dikerjakan',
					dimulai_pada TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					dikirim_pada DATETIME NULL,
					UNIQUE KEY unik_evaluasi_siswa (kuis_ujian_id, siswa_id),
					FOREIGN KEY (kuis_ujian_id) REFERENCES kuis_ujian(id) ON DELETE CASCADE,
					FOREIGN KEY (siswa_id) REFERENCES siswa(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				`CREATE TABLE IF NOT EXISTS jawaban_kuis_ujian (
					id INT AUTO_INCREMENT PRIMARY KEY,
					pengerjaan_id INT NOT NULL,
					soal_id INT NOT NULL,
					pilihan_id INT NOT NULL,
					UNIQUE KEY unik_jawaban_soal (pengerjaan_id, soal_id),
					FOREIGN KEY (pengerjaan_id) REFERENCES pengerjaan_kuis_ujian(id) ON DELETE CASCADE,
					FOREIGN KEY (soal_id) REFERENCES soal_kuis_ujian(id) ON DELETE CASCADE,
					FOREIGN KEY (pilihan_id) REFERENCES pilihan_soal_kuis_ujian(id) ON DELETE RESTRICT
				) ENGINE=InnoDB;`,
			},
		},

		// -------------------------------------------------------------------------
		// MIGRASI 5: Data Seed Bawaan (Idempotent - Aman Dijalankan Berulang Kali)
		// -------------------------------------------------------------------------
		{
			version:     "20261123_05_default_seeds",
			description: "Seeding Akun Utama (Admin, Guru, Siswa, Kurikulum, Kepsek) dan Data Master Default",
			queries: []string{
				// 1. Tahun Ajaran
				`INSERT INTO tahun_ajaran (id, tahun, semester, status_aktif)
				VALUES (1, '2025/2026', 'ganjil', TRUE), (2, '2025/2026', 'genap', FALSE)
				ON DUPLICATE KEY UPDATE tahun=VALUES(tahun);`,

				// 2. Jurusan
				`INSERT INTO jurusan (id, kode_jurusan, nama_jurusan, deskripsi) VALUES
				(1, 'PPLG', 'Pengembangan Perangkat Lunak dan Gim', 'Fokus pada pemrograman perangkat lunak, web, dan gim.'),
				(2, 'TJKT', 'Teknik Jaringan Komputer dan Telekomunikasi', 'Fokus pada jaringan komputer dan infrastruktur jaringan.'),
				(3, 'AKL', 'Akuntansi dan Keuangan Lembaga', 'Fokus pada bidang keuangan dan akuntansi.')
				ON DUPLICATE KEY UPDATE nama_jurusan=VALUES(nama_jurusan);`,

				// 3. Mata Pelajaran
				`INSERT INTO mata_pelajaran (id, kode_mapel, nama_mapel, kelompok, deskripsi) VALUES
				(1, 'UM-01', 'Pendidikan Agama dan Budi Pekerti', 'umum', 'Mata pelajaran umum keagamaan'),
				(2, 'UM-02', 'Pendidikan Pancasila dan Kewarganegaraan', 'umum', 'Mata pelajaran umum kenegaraan'),
				(3, 'UM-03', 'Bahasa Indonesia', 'umum', 'Mata pelajaran umum bahasa'),
				(4, 'UM-04', 'Matematika', 'umum', 'Mata pelajaran matematika dasar'),
				(5, 'KJ-PPLG-01', 'Pemrograman Web dan Perangkat Bergerak', 'kejuruan', 'Kejuruan khusus PPLG'),
				(6, 'KJ-PPLG-02', 'Basis Data', 'kejuruan', 'Kejuruan khusus PPLG'),
				(7, 'KJ-PPLG-03', 'Pemrograman Berorientasi Objek', 'kejuruan', 'Kejuruan khusus PPLG')
				ON DUPLICATE KEY UPDATE nama_mapel=VALUES(nama_mapel);`,

				// 4. Akun Admin (Password: admin123)
				`INSERT INTO users (id, username, email, password, role)
				SELECT 1, 'admin', 'admin@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'admin'
				WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin' OR email = 'admin@educlass.com');`,

				`INSERT INTO admins (user_id, nama_lengkap, no_telepon)
				SELECT u.id, 'Administrator Edu Class', '081234567890'
				FROM users u WHERE u.username = 'admin'
				AND NOT EXISTS (SELECT 1 FROM admins a WHERE a.user_id = u.id);`,

				// 5. Akun Guru Budi (Password: admin123)
				`INSERT INTO users (id, username, email, password, role)
				SELECT 2, 'budi', 'budi@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'guru'
				WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'budi' OR email = 'budi@educlass.com');`,

				`INSERT INTO guru (id, user_id, nip, nama_lengkap, jenis_kelamin, no_telepon, status)
				SELECT 1, u.id, '198001012005011001', 'Budi Santoso, S.Kom.', 'L', '081299998888', 'aktif'
				FROM users u WHERE u.username = 'budi'
				AND NOT EXISTS (SELECT 1 FROM guru g WHERE g.user_id = u.id);`,

				// 6. Kelas Rombel
				`INSERT INTO kelas (id, nama_kelas, tingkat, jurusan_id, tahun_ajaran_id, wali_kelas_id) VALUES
				(1, 'X PPLG 1', 'X', 1, 1, 1),
				(2, 'XI PPLG 1', 'XI', 1, 1, 1),
				(3, 'XI PPLG 2', 'XI', 1, 1, NULL),
				(4, 'XII PPLG 1', 'XII', 1, 1, NULL)
				ON DUPLICATE KEY UPDATE nama_kelas=VALUES(nama_kelas);`,

				// 7. Akun Siswa Dewi (Password: admin123, NIS: 112233)
				`INSERT INTO users (id, username, email, password, role)
				SELECT 3, 'dewi', 'dewi@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'siswa'
				WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'dewi' OR email = 'dewi@educlass.com');`,

				`INSERT INTO siswa (id, user_id, nis, nisn, nama_lengkap, jenis_kelamin, tempat_lahir, tanggal_lahir, kelas_id, jurusan_id, status)
				SELECT 1, u.id, '112233', '0051234567', 'Dewi Nur Azizah', 'P', 'Jakarta', '2005-08-17', 3, 1, 'aktif'
				FROM users u WHERE u.username = 'dewi'
				AND NOT EXISTS (SELECT 1 FROM siswa s WHERE s.user_id = u.id);`,

				// 8. Akun Kurikulum (Password: admin123)
				`INSERT INTO users (username, email, password, role)
				SELECT 'kurikulum', 'kurikulum@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'kurikulum'
				WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'kurikulum' OR email = 'kurikulum@educlass.com');`,

				`INSERT INTO kurikulum (user_id, nip, nama_lengkap, no_telepon)
				SELECT u.id, '197505102000032001', 'Dra. Hj. Siti Aminah, M.Pd.', '081234567800'
				FROM users u WHERE u.username = 'kurikulum'
				AND NOT EXISTS (SELECT 1 FROM kurikulum k WHERE k.user_id = u.id);`,

				// 9. Akun Kepala Sekolah (Password: admin123)
				`INSERT INTO users (username, email, password, role)
				SELECT 'kepala_sekolah', 'kepala.sekolah@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'kepala_sekolah'
				WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'kepala_sekolah' OR email = 'kepala.sekolah@educlass.com');`,

				`INSERT INTO kepala_sekolah (user_id, nip, nama_lengkap, no_telepon)
				SELECT u.id, '196803121995121001', 'Drs. H. Mulyadi, M.M.', '081234567899'
				FROM users u WHERE u.username = 'kepala_sekolah'
				AND NOT EXISTS (SELECT 1 FROM kepala_sekolah ks WHERE ks.user_id = u.id);`,

				// 10. Guru Mengajar (Guru Budi mengajar Pemrograman Web & Basis Data di kelas XI PPLG 1 & 2)
				`INSERT INTO guru_mengajar (guru_id, mata_pelajaran_id, kelas_id, tahun_ajaran_id)
				SELECT 1, 5, 2, 1 WHERE NOT EXISTS (SELECT 1 FROM guru_mengajar WHERE guru_id=1 AND mata_pelajaran_id=5 AND kelas_id=2);`,
				`INSERT INTO guru_mengajar (guru_id, mata_pelajaran_id, kelas_id, tahun_ajaran_id)
				SELECT 1, 6, 2, 1 WHERE NOT EXISTS (SELECT 1 FROM guru_mengajar WHERE guru_id=1 AND mata_pelajaran_id=6 AND kelas_id=2);`,
				`INSERT INTO guru_mengajar (guru_id, mata_pelajaran_id, kelas_id, tahun_ajaran_id)
				SELECT 1, 5, 3, 1 WHERE NOT EXISTS (SELECT 1 FROM guru_mengajar WHERE guru_id=1 AND mata_pelajaran_id=5 AND kelas_id=3);`,
				`INSERT INTO guru_mengajar (guru_id, mata_pelajaran_id, kelas_id, tahun_ajaran_id)
				SELECT 1, 6, 3, 1 WHERE NOT EXISTS (SELECT 1 FROM guru_mengajar WHERE guru_id=1 AND mata_pelajaran_id=6 AND kelas_id=3);`,
			},
		},

		// -------------------------------------------------------------------------
		// MIGRASI 6: Tabel Relasi Guru dan Mata Pelajaran Diampu
		// -------------------------------------------------------------------------
		{
			version:     "20261123_06_guru_mata_pelajaran",
			description: "Tabel Pivot guru_mata_pelajaran dan sinkronisasi mapel diampu bawaan",
			queries: []string{
				`CREATE TABLE IF NOT EXISTS guru_mata_pelajaran (
					id INT AUTO_INCREMENT PRIMARY KEY,
					guru_id INT NOT NULL,
					mata_pelajaran_id INT NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					UNIQUE KEY unik_guru_mapel (guru_id, mata_pelajaran_id),
					FOREIGN KEY (guru_id) REFERENCES guru(id) ON DELETE CASCADE,
					FOREIGN KEY (mata_pelajaran_id) REFERENCES mata_pelajaran(id) ON DELETE CASCADE
				) ENGINE=InnoDB;`,

				// Default seed relasi untuk guru Budi (ID 1) mengampu Mapel 5 (Pemrograman Web) dan Mapel 6 (Basis Data)
				`INSERT INTO guru_mata_pelajaran (guru_id, mata_pelajaran_id)
				SELECT 1, 5 WHERE NOT EXISTS (SELECT 1 FROM guru_mata_pelajaran WHERE guru_id=1 AND mata_pelajaran_id=5);`,
				`INSERT INTO guru_mata_pelajaran (guru_id, mata_pelajaran_id)
				SELECT 1, 6 WHERE NOT EXISTS (SELECT 1 FROM guru_mata_pelajaran WHERE guru_id=1 AND mata_pelajaran_id=6);`,
			},
		},
	}
}
