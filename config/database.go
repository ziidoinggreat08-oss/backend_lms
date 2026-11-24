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
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	if name == "" {
		name = "db_lms"
	}

	// 1. Pastikan database MySQL sudah dibuat sebelum membuka koneksi ke DB_NAME (mencegah error "Unknown database")
	if err := EnsureDatabaseExists(host, port, user, password, name); err != nil {
		log.Printf("[Peringatan] %v", err)
	}

	// 2. Hubungkan ke database target
	// DSN (Data Source Name): username:password@protocol(address)/dbname
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Gagal membuka koneksi ke database: %v", err)
	}

	// Tes Ping untuk memastikan koneksi ke database berhasil
	if err = db.Ping(); err != nil {
		log.Fatalf("Gagal terhubung ke MySQL (Pastikan MySQL berjalan di %s:%s): %v", host, port, err)
	}

	DB = db
	fmt.Printf("Koneksi ke database '%s' berhasil!\n", name)

	// 3. Jalankan Auto-Migration (idempotent, skema dan seed otomatis, tidak ada duplikasi tabel/data)
	if err := RunAutoMigrations(db); err != nil {
		log.Fatalf("Gagal menjalankan auto-migration: %v", err)
	}
}
