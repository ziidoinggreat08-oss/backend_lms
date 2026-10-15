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

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

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
	http.HandleFunc("/api/dashboard/stats", controllers.GetDashboardStats)

	// Routing untuk CRUD yang digenerate
	http.HandleFunc("/api/jurusan", controllers.HandleJurusan)
	http.HandleFunc("/api/jurusan/", controllers.HandleJurusan)

	http.HandleFunc("/api/kelas", controllers.HandleKelas)
	http.HandleFunc("/api/kelas/", controllers.HandleKelas)

	http.HandleFunc("/api/mata-pelajaran", controllers.HandleMataPelajaran)
	http.HandleFunc("/api/mata-pelajaran/", controllers.HandleMataPelajaran)

	http.HandleFunc("/api/guru", controllers.HandleGuru)
	http.HandleFunc("/api/guru/", controllers.HandleGuru)

	http.HandleFunc("/api/kepala-sekolah", controllers.HandleKepalaSekolah)
	http.HandleFunc("/api/kepala-sekolah/", controllers.HandleKepalaSekolah)

	http.HandleFunc("/api/kurikulum", controllers.HandleKurikulum)
	http.HandleFunc("/api/kurikulum/", controllers.HandleKurikulum)

	// Routing Siswa Custom (yang sudah ada sebelumnya)
	http.HandleFunc("/api/siswa", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodOptions {
			controllers.CreateSiswa(w, r)
			return
		}
		controllers.ListSiswa(w, r)
	})
	http.HandleFunc("/api/siswa/", controllers.StudentByID)

	// Route root untuk testing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Backend Edu Class Berjalan!\nKoneksi Database: Aktif")
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server berjalan di http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, withCORS(http.DefaultServeMux)); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
