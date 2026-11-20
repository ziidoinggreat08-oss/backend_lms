-- Schema Tambahan untuk Role Guru (Pembelajaran & Penilaian)
USE db_lms;

-- 1. Tabel Materi Pembelajaran
CREATE TABLE IF NOT EXISTS materi (
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
) ENGINE=InnoDB;

-- 2. Tabel Tugas
CREATE TABLE IF NOT EXISTS tugas (
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
) ENGINE=InnoDB;

-- 3. Tabel Pengumpulan Tugas Siswa
CREATE TABLE IF NOT EXISTS pengumpulan_tugas (
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
) ENGINE=InnoDB;

-- 4. Tabel Kuis & Ujian
CREATE TABLE IF NOT EXISTS kuis_ujian (
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
) ENGINE=InnoDB;

-- Soal dan opsi pilihan ganda untuk kuis/ujian. Dihapus otomatis bersama evaluasinya.
CREATE TABLE IF NOT EXISTS soal_kuis_ujian (
    id INT AUTO_INCREMENT PRIMARY KEY,
    kuis_ujian_id INT NOT NULL,
    pertanyaan TEXT NOT NULL,
    bobot INT NOT NULL DEFAULT 1,
    urutan INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (kuis_ujian_id) REFERENCES kuis_ujian(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS pilihan_soal_kuis_ujian (
    id INT AUTO_INCREMENT PRIMARY KEY,
    soal_id INT NOT NULL,
    teks TEXT NOT NULL,
    is_benar BOOLEAN NOT NULL DEFAULT FALSE,
    urutan INT NOT NULL,
    FOREIGN KEY (soal_id) REFERENCES soal_kuis_ujian(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- 5. Tabel Rekap Nilai Siswa
CREATE TABLE IF NOT EXISTS nilai_siswa (
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
) ENGINE=InnoDB;

