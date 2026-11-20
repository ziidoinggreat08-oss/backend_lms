USE db_lms;

-- Seed Akun Kurikulum Default jika belum ada
INSERT INTO users (username, email, password, role)
SELECT 'kurikulum', 'kurikulum@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'kurikulum'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'kurikulum' OR email = 'kurikulum@educlass.com');

-- Profile Kurikulum
INSERT INTO kurikulum (user_id, nip, nama_lengkap, no_telepon)
SELECT u.id, '197505102000032001', 'Dra. Hj. Siti Aminah, M.Pd.', '081234567800'
FROM users u
WHERE u.username = 'kurikulum'
AND NOT EXISTS (SELECT 1 FROM kurikulum k WHERE k.user_id = u.id);
