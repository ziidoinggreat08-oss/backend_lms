USE db_lms;

-- Riwayat revisi berkas tugas; pengumpulan_tugas tetap menyimpan berkas terakhir.
CREATE TABLE IF NOT EXISTS riwayat_pengumpulan_tugas (
    id INT AUTO_INCREMENT PRIMARY KEY,
    pengumpulan_id INT NOT NULL,
    file_url VARCHAR(255) NOT NULL,
    waktu_kumpul TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pengumpulan_id) REFERENCES pengumpulan_tugas(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS pengerjaan_kuis_ujian (
    id INT AUTO_INCREMENT PRIMARY KEY,
    kuis_ujian_id INT NOT NULL,
    siswa_id INT NOT NULL,
    status ENUM('dikerjakan', 'terkirim') NOT NULL DEFAULT 'dikerjakan',
    dimulai_pada TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    dikirim_pada DATETIME NULL,
    UNIQUE KEY unik_evaluasi_siswa (kuis_ujian_id, siswa_id),
    FOREIGN KEY (kuis_ujian_id) REFERENCES kuis_ujian(id) ON DELETE CASCADE,
    FOREIGN KEY (siswa_id) REFERENCES siswa(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS jawaban_kuis_ujian (
    id INT AUTO_INCREMENT PRIMARY KEY,
    pengerjaan_id INT NOT NULL,
    soal_id INT NOT NULL,
    pilihan_id INT NOT NULL,
    UNIQUE KEY unik_jawaban_soal (pengerjaan_id, soal_id),
    FOREIGN KEY (pengerjaan_id) REFERENCES pengerjaan_kuis_ujian(id) ON DELETE CASCADE,
    FOREIGN KEY (soal_id) REFERENCES soal_kuis_ujian(id) ON DELETE CASCADE,
    FOREIGN KEY (pilihan_id) REFERENCES pilihan_soal_kuis_ujian(id) ON DELETE RESTRICT
) ENGINE=InnoDB;
