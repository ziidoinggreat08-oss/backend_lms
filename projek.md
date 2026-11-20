# 📚 EDU CLASS — DOKUMENTASI PROJEK LMS (LEARNING MANAGEMENT SYSTEM)

Dokumen ini merupakan panduan komprehensif seluruh arsitektur, peran pengguna (*roles*), fitur yang sudah selesai dibangun, fitur yang belum dibuat (*roadmap*), serta struktur teknis projek **Edu Class**.

---

## 1. 📌 Gambaran Umum Projek

**Edu Class** adalah platform *Learning Management System* (LMS) berbasis web modern untuk institusi sekolah (SMK/SMA/MA). Sistem ini menerapkan konsep **Role-Based Access Control (RBAC)** di mana setiap pengguna login melalui satu gerbang login (*single gateway*) dan secara otomatis diarahkan ke dasbor serta hak akses yang sesuai dengan perannya.

### 🛠️ Tech Stack & Ekosistem
- **Frontend**: Next.js 16 (App Router), TypeScript, Tailwind CSS, Lucide Icons.
- **Backend**: Go (Golang) REST API, Go Standard Library + Chi/Mux architecture, JWT Authentication, Bcrypt Password Hashing.
- **Database**: MySQL 8.x (Laragon / MariaDB) — Database Name: `db_lms`.
- **Desain UI/UX**: Clean, Modern, Elegant School Theme (Figma-inspired).

---

## 2. 👥 Status & Daftar Role Pengguna

Sistem Edu Class dirancang untuk mengakomodasi **6 peran pengguna**:

| No | Role | Status Implementasi | Sifat Hak Akses | Deskripsi Singkat |
|:---:|:---|:---:|:---:|:---|
| 1 | **Admin** | ✅ **Selesai (Lengkap)** | Full CRUD Data Master | Pengelola data master sekolah, kelas, rombel, akun guru, siswa, dan staf. |
| 2 | **Guru** | ✅ **Selesai (Lengkap)** | Pembelajaran & Penilaian | Pembuat materi modul, penugasan, pembuat butir soal kuis & ujian, serta input nilai. |
| 3 | **Siswa** | ✅ **Selesai (Lengkap)** | Pembelajaran & Tugas | Menerima materi ajar, mengunggah berkas tugas, mengerjakan kuis & ujian digital. |
| 4 | **Kurikulum** | ✅ **Selesai (Lengkap)** | Monitoring & Alokasi | Pengawas mutu pembelajaran, alokasi guru pengampu (`guru_mengajar`), dan penetapan wali kelas. |
| 5 | **Kepala Sekolah** | ✅ **Selesai (Lengkap)** | View-Only Eksekutif | Pemantauan tingkat tinggi (*Executive Dashboard*), statistik grafik, monitoring nilai, dan laporan resmi siap cetak. |
| 6 | **Master Admin** | 📝 **Rencana / Konsep** | Super Owner | Pemilik sistem tingkat tertinggi untuk audit log sistem & manajemen akun Admin. |

---

## 3. 🔍 Rincian Isi & Fitur Tiap Role

### A. Role Admin (✅ Selesai)
Admin bertugas mengelola seluruh fondasi data akademik sekolah:
1. **Dasbor Statistik Admin**: Ringkasan data master (Total Siswa, Total Guru, Total Kelas, Total Jurusan, Total Mata Pelajaran) dan navigasi cepat.
2. **Manajemen Siswa**: CRUD Siswa, penempatan rombel kelas, jurusan, NIS, NISN, kontak, dan status aktif/non-aktif.
3. **Manajemen Guru**: CRUD Guru, NIP, kontak, gelar, jenis kelamin, dan status pendidik.
4. **Manajemen Kelas**: CRUD Rombel, penentuan tingkat (X, XI, XII), relasi jurusan (dengan kode jurusan otomatis), dan penunjukan wali kelas.
5. **Manajemen Jurusan**: CRUD Konsentrasi keahlian (contoh: PPLG, TJKT, AKL) beserta kode jurusan dan deskripsi.
6. **Manajemen Mata Pelajaran**: CRUD Mapel, kode mapel, kelompok (Umum, Kejuruan, Muatan Lokal).
7. **Manajemen Staf Pimpinan**: CRUD akun penanggung jawab Kurikulum dan Kepala Sekolah.

---

### B. Role Guru (✅ Selesai)
Guru berfokus penuh pada aktivitas belajar mengajar di kelas yang diampu:
1. **Dasbor Guru**: Statistik rombel ajar, jadwal kelas, tugas aktif, dan materi terbaru.
2. **Modul Materi Belajar**: Publikasi modul ajar digital, upload berkas dokumen/PDF, tautan referensi web eksternal per kelas & mapel.
3. **Penugasan Siswa**: CRUD tugas dengan penentuan batas waktu (*deadline*), instruksi penugasan, dan pemantauan berkas tugas siswa.
4. **Kuis Harian (Terpisah)**: Pengelolaan kuis formatif harian berbatas waktu (menit) dengan pembuatan langsung butir pertanyaan dan opsi pilihan ganda A, B, C, D beserta penentuan kunci jawaban.
5. **Ujian Semester (Terpisah)**: Pengelolaan Ujian Tengah/Akhir Semester (UTS/UAS) terjadwal dengan butir soal terstruktur dan bobot nilai per butir.
6. **Rekapitulasi Nilai Siswa**: Input nilai tugas, kuis, ujian, kalkulasi bobot otomatis (30% tugas + 30% kuis + 40% ujian), predikat huruf (A/B/C/D), dan catatan evaluasi pendidik.

---

### C. Role Siswa (✅ Selesai)
Siswa menggunakan portal untuk belajar mandiri dan mengerjakan kewajiban sekolah:
1. **Login Fleksibel**: Dapat masuk menggunakan **NIS**, **Username**, atau **Email** sekolah dengan password terverifikasi.
2. **Dasbor Pembelajaran Siswa**: Kartu progres tugas, pengingat tenggat (*deadline* terdekat), dan ringkasan kelas.
3. **E-Learning Materi**: Membaca materi ajar guru, mengunduh modul PDF, dan membuka referensi pelajaran.
4. **Pengumpulan Tugas**: Mengunggah berkas penugasan (PDF/Word/ZIP), melihat catatan guru, dan status pengumpulan tepat waktu / terlambat.
5. **Quiz Digital**: Mengerjakan kuis interaktif dengan timer berjalan dan sistem simpan jawaban.
6. **Ujian Online**: Mengikuti ujian sekolah terintegrasi secara daring.
7. **Profil Siswa**: Melihat biodata akademik, NIS, kelas, jurusan, serta memperbarui nomor kontak dan alamat pribadi. *(Catatan kebijakan: Siswa tidak melihat nilai/raport secara mentah demi privasi & ketertiban akademik sekolah).*

---

### D. Role Kurikulum (✅ Selesai)
Bidang Kurikulum mengawasi keterlaksanaan kegiatan belajar mengajar secara objektif:
1. **Dasbor Monitoring Akademik**: 8 cardbox interaktif terintegrasi data riil, pemetaan rombel per jurusan, dan feed aktivitas upload materi/tugas guru secara *real-time*.
2. **Data Akademik (Kelas & Jurusan Ringkas)**: Struktur terpadu rombel kelas per jurusan, filter tingkat (X, XI, XII), serta fitur **Penetapan / Pergantian Wali Kelas** langsung dari antarmuka.
3. **Plotting Guru Pengampu (`guru_mengajar`)**: Full CRUD alokasi guru pengajar (menetapkan guru mana mengajar mata pelajaran apa di rombel kelas mana).
4. **Monitoring Pembelajaran Terpadu**:
   - Tab 1: **Materi Ajar** — Modal pratinjau isi materi dan akses modul PDF guru.
   - Tab 2: **Penugasan Siswa** — Modal pantau daftar siswa yang mengumpulkan tugas beserta berkas & nilainya.
   - Tab 3: **Kuis & Ujian** — Modal pratinjau butir pertanyaan soal pilihan ganda A/B/C/D beserta kunci jawaban benar.
5. **Rekapitulasi Capaian Nilai**: Pemantauan rekap nilai se-sekolah, statistik rata-rata nilai kelas, persentase ketuntasan (KKM ≥ 75), dan cetak/ekspor laporan.

---

### E. Role Kepala Sekolah (⏳ Belum Dibuat / Tahap Berikutnya)
Fitur yang direncanakan untuk Kepala Sekolah (*Executive Level*):
1. **Executive Dashboard**: Ringkasan makro sekolah (kehadiran pendidik, produktivitas materi, ketercapaian kurikulum semester).
2. **Laporan Mutu Akademik**: Analisis grafik capaian rata-rata nilai siswa per jurusan dan rombel kelas.
3. **Kinerja Pengajar**: Pemantauan keaktifan guru dalam mengunggah materi, memberikan tugas, dan ketepatan koreksi nilai.
4. **Arsip Laporan Sekolah**: Unduh rekapitulasi laporan semester untuk dinas/yayasan pendidikan dalam format cetak/PDF resmi.
5. **Akses View-Only Seluruh Data**: Tidak melakukan CRUD data master, murni audit dan pengawasan manajerial.

---

### F. Role Master Admin (📝 Konsep / Rencana Tambahan)
Peran khusus pemilik sistem (*Super Administrator*):
1. **Owner Authority**: Berada di atas Admin biasa untuk proteksi infrastruktur.
2. **Manajemen Akun Admin**: Membuat, mengedit, menonaktifkan, dan menghapus akun Admin sekolah.
3. **Audit Log & Riwayat Aktivitas**: Pencatatan riwayat login/logout, waktu perubahan data vital sekolah, dan rekam jejak pengguna.
4. **System Maintenance**: Cadangan (*backup*) basis data dan pengaturan konfigurasi sistem global.

---

## 4. 🗄️ Struktur Basis Data (`db_lms`)

Database menggunakan MySQL dengan *engine* InnoDB dan relasi *Foreign Key* yang ketat:

```text
users ──────────┬── admins
                ├── guru ───────────┬── guru_mengajar (alokasi pengampu)
                ├── siswa           ├── materi
                ├── kurikulum       ├── tugas ────── pengumpulan_tugas
                └── kepala_sekolah  ├── kuis_ujian ── soal_kuis_ujian ── pilihan_soal
                                    └── nilai_siswa
tahun_ajaran ───┬── kelas
jurusan ────────┘
mata_pelajaran
```

---

## 5. 🔑 Kredensial Pengujian Bawaan (*Default Seeds*)

Semua akun dummy menggunakan password standar: **`admin123`**

| Role | Username / Identifier | Email | Password |
|:---|:---|:---|:---|
| **Admin** | `admin` | `admin@educlass.com` | `admin123` |
| **Guru** | `budi` | `budi@educlass.com` | `admin123` |
| **Siswa** | `dewi` *(atau NIS: `112233`)* | `dewi@educlass.com` | `admin123` |
| **Kurikulum** | `kurikulum` | `kurikulum@educlass.com` | `admin123` |
| **Kepala Sekolah** | `kepala_sekolah` | `kepala.sekolah@educlass.com` | `admin123` |

---

## 6. 🚀 Panduan Menjalankan Projek

### 1. Menjalankan Backend (Go)
```powershell
cd c:\Users\Dewi\OneDrive\Documents\LMS\backend_lms
# Menjalankan binary yang sudah di-build:
.\server.exe
# Atau menjalankan kode sumber langsung:
go run main.go
```
*Backend berjalan di: `http://localhost:8080`*

### 2. Menjalankan Frontend (Next.js)
```powershell
cd c:\Users\Dewi\OneDrive\Documents\LMS\frontend_lms
npm run dev
```
*Frontend berjalan di: `http://localhost:3000`*
