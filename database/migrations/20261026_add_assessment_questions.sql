-- Migrasi non-destruktif untuk database db_lms yang sudah ada.
USE db_lms;

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
