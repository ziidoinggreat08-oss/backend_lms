package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDB() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	// DSN (Data Source Name): username:password@protocol(address)/dbname
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Gagal membuka koneksi ke database: %v", err)
	}

	// Tes Ping untuk memastikan koneksi ke database berhasil
	if err = db.Ping(); err != nil {
		log.Fatalf("Gagal terhubung ke MySQL (Pastikan Laragon berjalan): %v", err)
	}

	DB = db
	fmt.Printf("Koneksi ke database '%s' berhasil!\n", name)

	// Pastikan akun default untuk role kurikulum dan kepala_sekolah selalu siap
	ensureDefaultUsers(db)
}

func ensureDefaultUsers(db *sql.DB) {
	// 1. Pastikan user Kurikulum ada
	var kCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'kurikulum'").Scan(&kCount)
	if kCount == 0 {
		res, err := db.Exec("INSERT INTO users (username, email, password, role) VALUES ('kurikulum', 'kurikulum@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'kurikulum')")
		if err == nil {
			uid, _ := res.LastInsertId()
			_, _ = db.Exec("INSERT INTO kurikulum (user_id, nip, nama_lengkap, no_telepon) VALUES (?, '197505102000032001', 'Dra. Hj. Siti Aminah, M.Pd.', '081234567800')", uid)
		}
	}

	// 2. Pastikan user Kepala Sekolah ada
	var ksCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE username IN ('kepala_sekolah', 'kepsek')").Scan(&ksCount)
	if ksCount == 0 {
		res, err := db.Exec("INSERT INTO users (username, email, password, role) VALUES ('kepala_sekolah', 'kepala.sekolah@educlass.com', '$2a$10$NIVpMEXSSvfHUTBlYaXt1uQgv63sZhIWH9M6MnbX0nvQPKRXVPDlm', 'kepala_sekolah')")
		if err == nil {
			uid, _ := res.LastInsertId()
			_, _ = db.Exec("INSERT INTO kepala_sekolah (user_id, nip, nama_lengkap, no_telepon) VALUES (?, '196803121995121001', 'Drs. H. Mulyadi, M.M.', '081234567899')", uid)
		}
	}
}
