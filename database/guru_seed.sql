-- Seed data pembelajaran untuk guru ID 1
USE db_lms;

-- Guru Mengajar (Guru ID 1 mengajar Mapel 5 di Kelas 3)
INSERT INTO guru_mengajar (guru_id, mata_pelajaran_id, kelas_id, tahun_ajaran_id)
VALUES (1, 5, 3, 1)
ON DUPLICATE KEY UPDATE tahun_ajaran_id=VALUES(tahun_ajaran_id);

-- Materi Contoh
INSERT INTO materi (guru_id, mata_pelajaran_id, kelas_id, judul, deskripsi, file_url, tautan_eksternal)
VALUES 
(1, 5, 3, 'Pengenalan Web & HTML5 Dasar', 'Mempelajari struktur semantik dasar dokumen HTML5, tag heading, paragraf, dan list.', 'https://example.com/modul-html5.pdf', 'https://developer.mozilla.org/en-US/docs/Web/HTML'),
(1, 5, 3, 'Styling Modern dengan Tailwind CSS', 'Pengenalan utility-first CSS framework untuk membangun antarmuka web responsif dengan cepat.', 'https://example.com/modul-css.pdf', 'https://tailwindcss.com/docs');

-- Tugas Contoh
INSERT INTO tugas (guru_id, mata_pelajaran_id, kelas_id, judul, deskripsi, deadline)
VALUES 
(1, 5, 3, 'Tugas 1: Membuat Halaman Profil Portofolio', 'Buatlah halaman web portofolio pribadi menggunakan HTML dan styling Tailwind CSS sesuai panduan di modul.', DATE_ADD(NOW(), INTERVAL 7 DAY));

-- Kuis & Ujian Contoh
INSERT INTO kuis_ujian (guru_id, mata_pelajaran_id, kelas_id, judul, tipe, durasi_menit, deskripsi, status)
VALUES 
(1, 5, 3, 'Kuis 1: Sintaks dan Semantik HTML', 'kuis', 30, 'Kuis singkat mengenai pemahaman tag HTML semantik.', 'aktif'),
(1, 5, 3, 'Ujian Tengah Semester: Praktik Web Dasar', 'ujian', 90, 'Ujian praktik pembuatan website mini.', 'draft');

-- Rekap Nilai Siswa (untuk siswa ID 1 & 3)
INSERT INTO nilai_siswa (siswa_id, mata_pelajaran_id, kelas_id, guru_id, nilai_tugas, nilai_kuis, nilai_ujian, nilai_akhir, catatan)
VALUES 
(1, 5, 3, 1, 88.50, 85.00, 90.00, 88.20, 'Sangat baik dalam pemahaman konsep'),
(3, 5, 3, 1, 80.00, 78.50, 82.00, 80.50, 'Cukup baik, perlu ditingkatkan dalam styling')
ON DUPLICATE KEY UPDATE nilai_tugas=VALUES(nilai_tugas), nilai_akhir=VALUES(nilai_akhir);

