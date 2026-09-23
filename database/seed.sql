USE db_lms;

-- 1. Seeding Data Tahun Ajaran
INSERT INTO tahun_ajaran (tahun, semester, status_aktif) VALUES
('2025/2026', 'ganjil', TRUE),
('2025/2026', 'genap', FALSE);

-- 2. Seeding Data Jurusan
INSERT INTO jurusan (kode_jurusan, nama_jurusan, deskripsi) VALUES
('PPLG', 'Pengembangan Perangkat Lunak dan Gim', 'Fokus pada pemrograman perangkat lunak, web, dan gim.'),
('TJKT', 'Teknik Jaringan Komputer dan Telekomunikasi', 'Fokus pada jaringan komputer dan infrastruktur jaringan.'),
('AKL', 'Akuntansi dan Keuangan Lembaga', 'Fokus pada bidang keuangan dan akuntansi.');

-- 3. Seeding Data Mata Pelajaran
INSERT INTO mata_pelajaran (kode_mapel, nama_mapel, kelompok, deskripsi) VALUES
('UM-01', 'Pendidikan Agama dan Budi Pekerti', 'umum', 'Mata pelajaran umum keagamaan'),
('UM-02', 'Pendidikan Pancasila dan Kewarganegaraan', 'umum', 'Mata pelajaran umum kenegaraan'),
('UM-03', 'Bahasa Indonesia', 'umum', 'Mata pelajaran umum bahasa'),
('UM-04', 'Matematika', 'umum', 'Mata pelajaran matematika dasar'),
('KJ-PPLG-01', 'Pemrograman Web dan Perangkat Bergerak', 'kejuruan', 'Kejuruan khusus PPLG'),
('KJ-PPLG-02', 'Basis Data', 'kejuruan', 'Kejuruan khusus PPLG'),
('KJ-PPLG-03', 'Pemrograman Berorientasi Objek', 'kejuruan', 'Kejuruan khusus PPLG');

-- 4. Seeding Data Akun Admin Default
INSERT INTO users (username, email, password, role) VALUES
('admin', 'admin@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'admin');

-- Set user_id into admins profile table (assuming ID is 1 for the first user inserted)
INSERT INTO admins (user_id, nama_lengkap, no_telepon) VALUES
(1, 'Administrator Edu Class', '081234567890');

-- 5. Seeding Data Akun Guru (Dummy)
INSERT INTO users (username, email, password, role) VALUES
('budi', 'budi@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'guru');

INSERT INTO guru (user_id, nip, nama_lengkap, jenis_kelamin, no_telepon, status) VALUES
(2, '198001012005011001', 'Budi Santoso, S.Kom.', 'L', '081299998888', 'aktif');

-- 6. Seeding Data Kelas (menggunakan Jurusan PPLG=1, Tahun Ajaran Ganjil=1, Wali Kelas Budi=1)
INSERT INTO kelas (nama_kelas, tingkat, jurusan_id, tahun_ajaran_id, wali_kelas_id) VALUES
('X PPLG 1', 'X', 1, 1, 1),
('XI PPLG 1', 'XI', 1, 1, 1),
('XI PPLG 2', 'XI', 1, 1, NULL),
('XII PPLG 1', 'XII', 1, 1, NULL);

-- 7. Seeding Data Akun Siswa (Dummy)
INSERT INTO users (username, email, password, role) VALUES
('dewi', 'dewi@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'siswa');

-- Siswa 1 di kelas XI PPLG 2 (ID kelas = 3), Jurusan PPLG (ID = 1)
INSERT INTO siswa (user_id, nis, nisn, nama_lengkap, jenis_kelamin, tempat_lahir, tanggal_lahir, kelas_id, jurusan_id, status) VALUES
(3, '112233', '0051234567', 'Dewi Nur Azizah', 'P', 'Jakarta', '2005-08-17', 3, 1, 'aktif');

-- 8. Seeding Penugasan Mengajar (Guru Budi mengajar Pemrograman Web (5) & Basis Data (6) di kelas XI PPLG 1 & 2)
INSERT INTO guru_mengajar (guru_id, mata_pelajaran_id, kelas_id, tahun_ajaran_id) VALUES
(1, 5, 2, 1),
(1, 6, 2, 1),
(1, 5, 3, 1),
(1, 6, 3, 1);

