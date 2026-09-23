package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"backend_lms/config"
	"backend_lms/controllers"

	"github.com/joho/godotenv"
)

func main() {
	// Load file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan variabel sistem")
	}

	// Inisialisasi Koneksi Database
	config.ConnectDB()

	// Mendaftarkan Endpoint (Route) API
	http.HandleFunc("/api/login", controllers.Login)
	http.HandleFunc("/api/siswa", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodOptions {
			controllers.CreateSiswa(w, r)
			return
		}
		controllers.ListSiswa(w, r)
	})
	http.HandleFunc("/api/siswa/", controllers.StudentByID)
	http.HandleFunc("/api/kelas", controllers.ListKelas)

	// Route root untuk testing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Backend Edu Class Berjalan!\nKoneksi Database: Aktif")
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server berjalan di http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
