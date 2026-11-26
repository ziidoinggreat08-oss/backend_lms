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

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// 1. Memuat konfigurasi dari file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan variabel sistem.")
	}

	// 2. Menghubungkan ke Database MySQL
	config.ConnectDB()

	// 3. Routing Endpoint API
	http.HandleFunc("/api/login", controllers.Login)
	http.HandleFunc("/api/dashboard", controllers.GetDashboardStats)
	http.HandleFunc("/api/dashboard/", controllers.GetDashboardStats)
	http.HandleFunc("/api/dashboard/stats", controllers.GetDashboardStats)
	http.HandleFunc("/api/admin/dashboard", controllers.GetDashboardStats)
	http.HandleFunc("/api/admin/dashboard/", controllers.GetDashboardStats)

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

	// Routing Portal Kepala Sekolah (Pimpinan Eksekutif - Read-Only)
	http.HandleFunc("/api/kepala-sekolah-portal/dashboard", controllers.HandleKepalaSekolahPortal)
	http.HandleFunc("/api/kepala-sekolah-portal/laporan", controllers.HandleKepalaSekolahPortal)
	http.HandleFunc("/api/kepala-sekolah-portal/profil", controllers.HandleKepalaSekolahPortal)
	http.HandleFunc("/api/kepala-sekolah-portal/", controllers.HandleKepalaSekolahPortal)

	// Routing Kurikulum Portal (Monitoring & Manajemen Akademik)
	http.HandleFunc("/api/kurikulum/dashboard", controllers.HandleKurikulumPortal)
	http.HandleFunc("/api/kurikulum/akademik", controllers.HandleKurikulumPortal)
	http.HandleFunc("/api/kurikulum/akademik/", controllers.HandleKurikulumPortal)
	http.HandleFunc("/api/kurikulum/pengampu", controllers.HandleKurikulumPortal)
	http.HandleFunc("/api/kurikulum/pengampu/", controllers.HandleKurikulumPortal)
	http.HandleFunc("/api/kurikulum/monitoring", controllers.HandleKurikulumPortal)
	http.HandleFunc("/api/kurikulum/monitoring/", controllers.HandleKurikulumPortal)
	http.HandleFunc("/api/kurikulum/nilai", controllers.HandleKurikulumPortal)
	http.HandleFunc("/api/kurikulum/options", controllers.HandleKurikulumPortal)

	// CRUD Kurikulum oleh Admin
	http.HandleFunc("/api/kurikulum", controllers.HandleKurikulum)
	http.HandleFunc("/api/kurikulum/", controllers.HandleKurikulum)

	// Routing Khusus Role Guru (Pembelajaran & Penilaian)
	http.HandleFunc("/api/guru/dashboard", controllers.HandleGuruDashboard)
	http.HandleFunc("/api/guru/mata-pelajaran", controllers.HandleGuruMataPelajaran)
	http.HandleFunc("/api/guru/materi", controllers.HandleGuruMateri)
	http.HandleFunc("/api/guru/materi/", controllers.HandleGuruMateri)
	http.HandleFunc("/api/guru/tugas", controllers.HandleGuruTugas)
	http.HandleFunc("/api/guru/tugas/", controllers.HandleGuruTugas)
	http.HandleFunc("/api/guru/kuis-ujian", controllers.HandleGuruKuisUjian)
	http.HandleFunc("/api/guru/kuis-ujian/", controllers.HandleGuruKuisUjian)
	// Endpoint terpisah agar UI dan izin Kuis/Ujian tidak tercampur.
	http.HandleFunc("/api/guru/kuis", controllers.HandleGuruKuisUjian)
	http.HandleFunc("/api/guru/kuis/", controllers.HandleGuruKuisUjian)
	http.HandleFunc("/api/guru/ujian", controllers.HandleGuruKuisUjian)
	http.HandleFunc("/api/guru/ujian/", controllers.HandleGuruKuisUjian)
	http.HandleFunc("/api/guru/nilai", controllers.HandleGuruNilai)
	http.HandleFunc("/api/guru/upload", controllers.HandleFileUpload)
	http.HandleFunc("/api/upload", controllers.HandleFileUpload)

	// Routing Siswa Custom
	http.HandleFunc("/api/siswa", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodOptions {
			controllers.CreateSiswa(w, r)
			return
		}
		controllers.ListSiswa(w, r)
	})
	http.HandleFunc("/api/siswa/", controllers.StudentByID)
	// Portal pembelajaran siswa dipisahkan dari CRUD siswa milik Admin.
	http.HandleFunc("/api/siswa/dashboard", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/materi", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/materi/", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/tugas", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/tugas/", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/riwayat-tugas", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/quiz", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/quiz/", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/ujian", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/ujian/", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/profil", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/jadwal", controllers.HandleSiswaPortal)
	http.HandleFunc("/api/siswa/notifikasi", controllers.HandleSiswaPortal)

	// Routing Profil Terpadu Seluruh Role (View-Only)
	http.HandleFunc("/api/profil/me", controllers.HandleGetMyProfile)

	// Static file server untuk berkas uploads (materi, tugas, screenshot)
	_ = os.MkdirAll("uploads", 0755)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

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
