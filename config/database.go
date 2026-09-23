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
}
