# 📚 EDU CLASS — DOKUMENTASI PROJEK LMS (LEARNING MANAGEMENT SYSTEM)

Dokumen ini merupakan panduan komprehensif seluruh arsitektur, peran pengguna (*roles*), rincian fitur yang telah selesai dibangun, peta rute sistem (*roadmap*), skema basis data, rute API, dan panduan menjalankan projek **Edu Class**.

---

## 1. 📌 Gambaran Umum Projek

**Edu Class** adalah platform *Learning Management System* (LMS) berbasis web modern untuk institusi sekolah kejuruan/menengah (SMK/SMA/MA). Sistem ini menerapkan konsep **Role-Based Access Control (RBAC)** di mana seluruh pengguna masuk melalui satu pintu gerbang login (*single gateway*) dan secara otomatis diarahkan ke dasbor serta hak akses sesuai dengan perannya.

### 🛠️ Tech Stack & Ekosistem
- **Frontend**: Next.js 16 (App Router), TypeScript, Tailwind CSS, Lucide Icons / Heroicons, Fetch API.
- **Backend**: Go (Golang) REST API murni menggunakan pustaka standar `net/http` yang tangguh dan ringan, JWT Authentication (`golang-jwt/jwt/v5`), Bcrypt Password Hashing (`golang.org/x/crypto/bcrypt`), serta CORS middleware.
- **Database**: MySQL 8.x / MariaDB (Laragon / Standalone) — Nama Database: `db_lms` dengan relasi ketat (*InnoDB Foreign Keys*).
- **Desain UI/UX**: Clean, Modern, Elegant School Theme (Figma-inspired, dominan warna Navy `#003399` dan Slate).

---

## 2. 👥 Status & Daftar Role Pengguna

Sistem Edu Class dirancang untuk mengakomodasi **6 peran pengguna**, dengan **5 peran utama telah selesai diimplementasikan 100% secara *full-stack***:

| No | Role | Status Implementasi | Sifat Hak Akses | Deskripsi Singkat |
|:---:|:---|:---:|:---:|:---|
| 1 | **Admin** | ✅ **Selesai (Lengkap)** | Full CRUD Data Master | Pengelola fondasi data akademik sekolah: siswa, guru, kelas, jurusan, mapel, dan staf pimpinan. |
| 2 | **Guru** | ✅ **Selesai (Lengkap)** | KBM & Penilaian | Pembuat modul materi ajar, penugasan, kuis harian, ujian semester, koreksi tugas, dan rekapitulasi nilai. |
| 3 | **Siswa** | ✅ **Selesai (Lengkap)** | Pembelajaran & Ujian | Mengakses materi, mengunggah berkas tugas, mengerjakan kuis interaktif bertimer, ujian semester, serta kelola profil. |
| 4 | **Kurikulum** | ✅ **Selesai (Lengkap)** | Monitoring & Alokasi | Pengawas mutu pembelajaran, alokasi guru pengampu (`guru_mengajar`), penetapan wali kelas, dan monitoring rekap nilai. |
| 5 | **Kepala Sekolah** | ✅ **Selesai (Lengkap)** | View-Only Eksekutif | Pemantauan makro tingkat tinggi (*Executive Dashboard*), monitoring KBM, laporan resmi siap cetak (*print-friendly*), dan direktori sekolah. |
| 6 | **Master Admin** | 📝 **Rencana / Konsep** | Super Owner | Pemilik sistem tingkat tertinggi untuk audit log global, multi-sekolah/tenant, dan manajemen akun admin sekolah. |

---

## 3. 🔍 Rincian Fitur yang Sudah Dibuat per Role

### A. Role Admin (✅ Selesai)
*Route Frontend: `/admin/*` | Endpoint API: `/api/*`*

Admin bertugas mengelola seluruh fondasi data master akademik dan akun pengguna sekolah:
1. **Dasbor Statistik Admin (`/admin/dashboard`)**:
   - Kartu statistik total data: Total Siswa, Guru, Kelas Rombel, Jurusan, Mapel, Staf Kurikulum, dan Kepala Sekolah.
   - Pintasan navigasi cepat ke setiap menu manajemen data.
2. **Manajemen Siswa (`/admin/siswa`)**:
   - Tambah, lihat, ubah, dan hapus data siswa (CRUD).
   - Form komprehensif: NIS, NISN, Nama Lengkap, Jenis Kelamin, Tempat & Tanggal Lahir, No. Telepon, Alamat, Rombel/Kelas, Jurusan, dan Status Aktif/Nonaktif.
   - Pembuatan akun login sistem secara otomatis dengan enkripsi Bcrypt.
3. **Manajemen Guru (`/admin/guru`)**:
   - CRUD data guru pendidik lengkap: NIP, Nama Lengkap beserta Gelar, Jenis Kelamin, No. Telepon, Alamat, Status Pendidik (Aktif/Nonaktif).
   - Akun pengguna terhubung langsung ke data pengajar.
4. **Manajemen Struktur Kelas Rombel (`/admin/kelas`)**:
   - CRUD rombel kelas per tingkat (Tingkat X, XI, XII).
   - Relasi langsung ke jurusan keahlian dan penunjukan Wali Kelas dari daftar guru terdaftar.
5. **Manajemen Jurusan Keahlian (`/admin/jurusan`)**:
   - CRUD Konsentrasi Keahlian (misal: PPLG, TJKT, AKL, DKV) dengan kode jurusan unik dan deskripsi program.
6. **Manajemen Mata Pelajaran (`/admin/mata-pelajaran`)**:
   - CRUD Mata Pelajaran, Kode Mapel unik, Nama Mapel, Deskripsi, serta Kelompok Mapel (Umum, Kejuruan, Muatan Lokal).
7. **Manajemen Akun Staf Pimpinan (`/admin/kurikulum` & `/admin/kepala-sekolah`)**:
   - Pembuatan dan pengelolaan akun khusus penanggung jawab Bidang Kurikulum.
   - Pembuatan dan pengelolaan akun resmi Kepala Sekolah (NIP, Nama, Masa Jabatan, Kontak).

---

### B. Role Guru (✅ Selesai)
*Route Frontend: `/guru/*` | Endpoint API: `/api/guru/*`*

Guru berfokus penuh pada aktivitas belajar mengajar di kelas dan mata pelajaran yang diampu:
1. **Dasbor Guru (`/guru/dashboard`)**:
   - Statistik kelas ajar yang diampu, jadwal pelajaran, total materi dipublikasikan, dan tugas aktif.
   - Ringkasan aktivitas pengumpulan tugas siswa terbaru.
2. **Modul Materi Belajar (`/guru/materi`)**:
   - Publikasi materi ajar digital per rombel kelas dan mata pelajaran.
   - Fitur unggah berkas modul materi (PDF/dokumen) dan penyertaan tautan referensi web eksternal/video pembelajaran.
3. **Penugasan Siswa (`/guru/tugas`)**:
   - Pembuatan tugas dengan judul, instruksi terstruktur, batas waktu (*deadline*), dan lampiran berkas pendukung.
   - Peninjauan berkas tugas yang dikumpulkan siswa secara langsung.
   - Fitur pemberian nilai (0-100) dan catatan evaluasi/umpan balik dari guru kepada siswa.
4. **Kuis Harian Digital (`/guru/kuis`)**:
   - Pengelolaan kuis formatif terpisah berbatas waktu pengerjaan (durasi menit).
   - Pembuatan butir soal pilihan ganda langsung pada antarmuka (Opsi A, B, C, D), penentuan kunci jawaban benar, dan bobot nilai per butir.
   - Pengaturan status kuis (*draft*, aktif, selesai).
5. **Ujian Semester (`/guru/ujian`)**:
   - Pengelolaan Ujian Tengah/Akhir Semester (PTS/PAS) terstruktur dengan penjadwalan waktu mulai dan selesai.
   - Bank soal pilihan ganda dengan bobot nilai terintegrasi.
6. **Rekapitulasi Nilai Siswa (`/guru/nilai`)**:
   - Pemilihan filter berdasarkan kelas dan mata pelajaran yang diampu.
   - Input nilai tugas, nilai kuis, dan nilai ujian.
   - Kalkulasi otomatis **Nilai Akhir** berbasis pembobotan standar:
     $$\text{Nilai Akhir} = (30\% \times \text{Tugas}) + (30\% \times \text{Kuis}) + (40\% \times \text{Ujian})$$
   - Penentuan otomatis predikat kelulusan huruf (**A**, **B**, **C**, **D**) serta kolom catatan evaluasi pendidik.

---

### C. Role Siswa (✅ Selesai)
*Route Frontend: `/siswa/*` | Endpoint API: `/api/siswa/*`*

Siswa menggunakan portal untuk belajar mandiri, mengerjakan tugas, dan evaluasi daring:
1. **Login Fleksibel Multi-Identitas**:
   - Siswa dapat masuk menggunakan **NIS**, **Username**, atau **Email** terdaftar dengan verifikasi sandi aman.
   - Proteksi status aktif (siswa non-aktif dicegah masuk secara otomatis).
2. **Dasbor Belajar Siswa (`/siswa/dashboard`)**:
   - Tampilan progres tugas berjalan, pengingat tenggat waktu terdekat (*upcoming deadlines*), modul materi baru, dan status evaluasi.
3. **Akses Materi Belajar (`/siswa/materi`)**:
   - Membaca deskripsi materi sesuai tingkatan kelas dan mata pelajarannya.
   - Mengunduh modul PDF materi atau membuka tautan eksternal yang dibagikan guru.
4. **Pengumpulan Tugas (`/siswa/tugas`)**:
   - Meninjau instruksi tugas dan lampiran berkas dari guru.
   - Mengunggah berkas pengumpulan tugas (PDF/Word/ZIP) beserta catatan pengantar siswa.
   - Pelacakan status pengumpulan: tepat waktu / terlambat (*late submission*), status penilaian, serta catatan koreksi dari guru.
5. **Riwayat Tugas (`/siswa/riwayat-tugas`)**:
   - Arsip seluruh tugas yang pernah dikerjakan dan dikumpulkan.
6. **Kuis Interaktif Digital (`/siswa/quiz`)**:
   - Mengerjakan kuis interaktif secara daring dengan penghitung waktu mundur (*countdown timer*).
   - Pemilihan opsi jawaban pilihan ganda, tombol navigasi soal, dan konfirmasi pengiriman (*submit*).
7. **Ujian Semester Daring (`/siswa/ujian`)**:
   - Halaman khusus pengerjaan Ujian Tengah/Akhir Semester secara resmi dan terjadwal.
8. **Jadwal Pelajaran (`/siswa/jadwal`)**:
   - Tampilan jadwal mata pelajaran siswa.
9. **Pusat Notifikasi (`/siswa/notifikasi`)**:
   - Pemberitahuan kegiatan akademik sekolah dan evaluasi baru.
10. **Profil Siswa Mandiri (`/siswa/profil`)**:
    - Melihat informasi akademik resmi: NIS, NISN, Nama Lengkap, Rombel Kelas, dan Jurusan.
    - Pembaruan mandiri data kontak pribadi (Nomor Telepon dan Alamat Domisili).
    *(Catatan: Siswa tidak melihat tabel nilai mentah secara bebas demi privasi dan kebijakan rapor sekolah).*
11. **Pengaturan Akun (`/siswa/pengaturan`)**:
    - Antarmuka preferensi dan keamanan akun siswa.

---

### D. Role Kurikulum (✅ Selesai)
*Route Frontend: `/kurikulum/*` | Endpoint API: `/api/kurikulum/*`*

Bidang Kurikulum bertindak sebagai pengawas mutu kegiatan belajar mengajar sekolah:
1. **Dasbor Monitoring Akademik (`/kurikulum/dashboard`)**:
   - 8 kartu metrik terpadu data riil: Total Guru, Siswa, Kelas, Jurusan, Mapel, Materi Terbit, Tugas Aktif, dan Total Evaluasi.
   - Distribusi visual rombel kelas per jurusan dan riwayat *feed* unggahan materi/tugas guru secara *real-time*.
2. **Data Akademik & Penetapan Wali Kelas (`/kurikulum/akademik`)**:
   - Struktur rombel kelas per jurusan dengan filter tingkat (X, XI, XII).
   - Fitur **Penetapan / Pergantian Wali Kelas secara instan** langsung dari antarmuka kurikulum.
3. **Mapel & Alokasi Pengampu (`/kurikulum/pengampu`)**:
   - Manajemen *plotting* alokasi guru pengajar (`guru_mengajar`).
   - Penetapan guru mana mengajar mata pelajaran apa di rombel kelas mana.
4. **Monitoring Pembelajaran Terpadu 3 Pilar (`/kurikulum/monitoring`)**:
   - **Tab Materi**: Pratinjau detail materi, unduhan modul dokumen guru, dan informasi tanggal rilis.
   - **Tab Penugasan**: Supervisi daftar tugas guru, pemantauan nama-nama siswa yang telah mengumpulkan tugas, berkas tugas, dan nilai yang diberikan.
   - **Tab Evaluasi (Kuis & Ujian)**: Pratinjau butir pertanyaan soal pilihan ganda A/B/C/D, kunci jawaban benar, durasi kuis/ujian, dan statusnya.
5. **Rekapitulasi Capaian Nilai Siswa (`/kurikulum/nilai`)**:
   - Pemantauan nilai seluruh siswa lintas rombel dan mata pelajaran.
   - Statistik rata-rata nilai kelas dan persentase ketuntasan belajar (kriteria KKM $\ge 75$).
   - Pratinjau cetak (*Print/Export Report*) rekapitulasi capaian akademik.

---

### E. Role Kepala Sekolah (✅ Selesai)
*Route Frontend: `/kepala-sekolah/*` | Endpoint API: `/api/kepala-sekolah-portal/*`*

Kepala Sekolah memiliki hak akses **Eksekutif Manajerial** dengan sifat akses **Strictly Read-Only** (hanya metode `GET` di portal API, mencegah manipulasi data demi validitas audit):
1. **Executive Dashboard (`/kepala-sekolah/dashboard`)**:
   - Kartu metrik makro lembaga: Total Siswa Aktif, Total Guru Pengajar, Rombel Kelas, Jurusan, Mata Pelajaran, Total Materi Terbit, Tugas Aktif, Kuis, Ujian, dan Jumlah Tugas Dikumpulkan.
   - Grafik distribusi rasio siswa per jurusan keahlian.
   - Tabel pemetaan rombel kelas beserta nama wali kelas yang bertugas dan jumlah siswa per kelas.
   - Ringkasan evaluasi sekolah: Rata-rata nilai tugas, kuis, ujian, estimasi nilai akhir sekolah, serta tingkat ketuntasan KKM.
   - Linimasa aktivitas terbaru LMS (*live feed* publikasi materi dan penugasan).
2. **Monitoring Akademik Eksekutif (`/kepala-sekolah/monitoring`)**:
   - Filter pemantauan berdasarkan tingkat kelas dan mata pelajaran.
   - **Sub-tab Nilai**: Pengawasan nilai tugas, kuis, ujian, nilai akhir, dan predikat siswa se-sekolah.
   - **Sub-tab Evaluasi**: Pengawasan keterlaksanaan kuis harian dan ujian semester beserta durasi dan status pengerjaannya.
   - **Sub-tab Kepatuhan Tugas**: Monitoring rasio siswa yang mengumpulkan tugas dan kedisiplinan koreksi nilai guru.
3. **Laporan Resmi Sekolah Siap Cetak (`/kepala-sekolah/laporan`)**:
   - Antarmuka dokumen resmi pimpinan (*Print-friendly Layout* dengan CSS `@media print`).
   - Rekapitulasi per rombel kelas lengkap (tingkat, jurusan, wali kelas, jumlah siswa, dan status operasional).
   - Rekapitulasi per jurusan keahlian (total rombel, total siswa, dan persentase siswa).
   - Metrik rasio guru-ke-siswa dan persentase ketuntasan akademik sekolah untuk bahan rapat pimpinan atau pelaporan ke dinas/yayasan pendidikan.
   - Tombol cetak langsung (*Print to PDF / Printer fisik*).
4. **Direktori Data Siswa Eksekutif (`/kepala-sekolah/siswa`)**:
   - Pencarian dan filter data siswa aktif/nonaktif per rombel dan jurusan (*View-Only*).
5. **Direktori Data Guru Eksekutif (`/kepala-sekolah/guru`)**:
   - Direktori profil pengajar sekolah: NIP, kontak, mapel binaan, rombel ajar, dan status (*View-Only*).
6. **Data Kelas & Jurusan (`/kepala-sekolah/kelas-jurusan`)**:
   - Tinjauan struktur rombel kelas per tingkat dan konsentrasi keahlian jurusan (*View-Only*).
7. **Direktori Mata Pelajaran (`/kepala-sekolah/mata-pelajaran`)**:
   - Katalog kurikulum mata pelajaran (Umum, Kejuruan, Muatan Lokal) dan guru pengampunya (*View-Only*).
8. **Profil Akun Pimpinan (`/kepala-sekolah/profil`)**:
   - Informasi profil Kepala Sekolah terverifikasi dan kendali keluar sesi (*logout*).

---

### F. Role Master Admin (📝 Konsep / Rencana Tahap Berikutnya)
Peran khusus pemilik infrastruktur sistem (*Super Administrator*):
1. **Owner Authority**: Berada di level tertinggi untuk proteksi infrastruktur multi-sekolah atau multi-instansi.
2. **Manajemen Akun Admin**: Membuat, mengaudit, dan menonaktifkan akun Admin sekolah.
3. **Audit Log & Jejak Sistem**: Pencatatan riwayat autentikasi, rekam jejak modifikasi basis data sensitif, dan log aktivitas API.
4. **System Maintenance**: Utilitas pencadangan (*backup/restore*) basis data dan konfigurasi global.

---

## 4. 🌐 Struktur Routing Frontend & Endpoint API Backend

### A. Rute Halaman Frontend (Next.js App Router)
```text
frontend_lms/app/
├── login/                        # Single Gateway Login Page
├── admin/                        # Portal Administrator
│   ├── dashboard/                # Ringkasan data master
│   ├── siswa/                    # CRUD Siswa
│   ├── guru/                     # CRUD Guru
│   ├── kelas/                    # CRUD Kelas & Rombel
│   ├── jurusan/                  # CRUD Jurusan Keahlian
│   ├── mata-pelajaran/           # CRUD Mata Pelajaran
│   ├── kurikulum/                # Manajemen Staf Kurikulum
│   └── kepala-sekolah/           # Manajemen Akun Kepala Sekolah
├── guru/                         # Portal Guru Pengajar
│   ├── dashboard/                # Dasbor guru & statistik KBM
│   ├── materi/                   # CRUD Modul materi & file PDF
│   ├── tugas/                    # CRUD Tugas & koreksi tugas siswa
│   ├── kuis/                     # Kuis harian & butir soal pilihan ganda
│   ├── ujian/                    # Ujian semester PTS/PAS terstruktur
│   └── nilai/                    # Rekapitulasi & kalkulasi nilai akhir
├── siswa/                        # Portal Siswa
│   ├── dashboard/                # Dasbor belajar & deadline terdekat
│   ├── materi/                   # Baca & unduh modul ajar
│   ├── tugas/                    # Unggah berkas tugas & catatan
│   ├── riwayat-tugas/            # Arsip tugas selesai
│   ├── quiz/                     # Pengerjaan kuis bertimer interaktif
│   ├── ujian/                    # Pengerjaan ujian semester daring
│   ├── jadwal/                   # Jadwal pelajaran
│   ├── notifikasi/               # Notifikasi akademik
│   ├── profil/                   # Profil siswa & edit kontak pribadi
│   └── pengaturan/               # Pengaturan akun
├── kurikulum/                    # Portal Bidang Kurikulum
│   ├── dashboard/                # Dasbor monitoring akademik 8 kartu
│   ├── akademik/                 # Data rombel & pergantian wali kelas
│   ├── pengampu/                 # Plotting alokasi guru_mengajar
│   ├── monitoring/               # Supervisi 3 pilar: Materi, Tugas, Kuis/Ujian
│   └── nilai/                    # Rekap capaian nilai & KKM se-sekolah
└── kepala-sekolah/               # Portal Kepala Sekolah (Read-Only)
    ├── dashboard/                # Executive Dashboard makro sekolah
    ├── siswa/                    # Direktori data siswa
    ├── guru/                     # Direktori data guru
    ├── kelas-jurusan/            # Direktori kelas & jurusan
    ├── mata-pelajaran/           # Direktori kurikulum mapel
    ├── monitoring/               # Monitoring terpadu nilai, evaluasi, tugas
    ├── laporan/                  # Laporan resmi siap cetak / PDF
    └── profil/                   # Profil akun pimpinan
```

### B. Endpoint API Backend (Go REST API)
```text
POST   /api/login                               # Otentikasi JWT (Email/Username/NIS)
GET    /api/dashboard                           # Statistik data master admin
GET    /api/jurusan, POST, PUT, DELETE          # CRUD Jurusan
GET    /api/kelas, POST, PUT, DELETE            # CRUD Kelas & Wali Kelas
GET    /api/mata-pelajaran, POST, PUT, DELETE   # CRUD Mata Pelajaran
GET    /api/guru, POST, PUT, DELETE             # CRUD Guru
GET    /api/siswa, POST, PUT, DELETE            # CRUD Siswa
GET    /api/kurikulum, POST, PUT, DELETE        # CRUD Akun Kurikulum
GET    /api/kepala-sekolah, POST, PUT, DELETE   # CRUD Akun Kepala Sekolah

# Portal Guru:
GET    /api/guru/dashboard                      # Dasbor aktivitas guru
GET    /api/guru/materi, POST, PUT, DELETE      # Kelola modul materi ajar
GET    /api/guru/tugas, POST, PUT, DELETE       # Kelola tugas & nilai pengumpulan
GET    /api/guru/kuis, POST, PUT, DELETE        # Kelola kuis & butir soal
GET    /api/guru/ujian, POST, PUT, DELETE       # Kelola ujian semester & butir soal
GET    /api/guru/nilai, POST                    # Rekap & simpan nilai siswa

# Portal Siswa:
GET    /api/siswa/dashboard                     # Dasbor belajar siswa
GET    /api/siswa/materi                        # Akses materi sesuai rombel siswa
GET    /api/siswa/tugas, POST                   # Daftar tugas & kirim berkas tugas
GET    /api/siswa/quiz, POST                    # Daftar kuis & simpan jawaban
GET    /api/siswa/ujian, POST                   # Daftar ujian & submit pengerjaan
GET    /api/siswa/profil, PUT                   # Lihat & ubah kontak profil siswa

# Portal Kurikulum:
GET    /api/kurikulum/dashboard                 # Metrik monitoring kurikulum
GET    /api/kurikulum/akademik, PUT             # Data kelas & update wali kelas
GET    /api/kurikulum/pengampu, POST, DELETE    # Alokasi guru_mengajar
GET    /api/kurikulum/monitoring                # Supervisi materi, tugas, kuis/ujian
GET    /api/kurikulum/nilai                     # Rekap nilai & ketuntasan KKM
GET    /api/kurikulum/options                   # Data opsi filter kelas & mapel

# Portal Kepala Sekolah (Strictly Read-Only / GET):
GET    /api/kepala-sekolah-portal/dashboard     # Data eksekutif makro sekolah
GET    /api/kepala-sekolah-portal/laporan       # Data laporan lengkap siap cetak
GET    /api/kepala-sekolah-portal/profil        # Profil pimpinan
```

---

## 5. 🗄️ Struktur Basis Data (`db_lms`)

Database menggunakan MySQL dengan *engine* InnoDB dan relasi *Foreign Key* yang ketat:

```text
users ──────────┬── admins
                ├── guru ───────────┬── guru_mengajar (alokasi pengampu)
                ├── siswa           ├── materi
                ├── kurikulum       ├── tugas ────── pengumpulan_tugas ── riwayat_pengumpulan_tugas
                └── kepala_sekolah  ├── kuis_ujian ── soal_kuis_ujian ── pilihan_soal_kuis_ujian
                                    │               └── pengerjaan_kuis_ujian ── jawaban_kuis_ujian
                                    └── nilai_siswa
tahun_ajaran ───┬── kelas
jurusan ────────┘
mata_pelajaran
```

### Ringkasan Entitas Tabel:
1. `users`: Tabel induk autentikasi (`username`, `email`, `password` bcrypt, `role` ENUM: `admin`, `guru`, `siswa`, `kurikulum`, `kepala_sekolah`).
2. `admins`: Profil penanggung jawab administrasi sistem.
3. `guru`: Profil guru pendidik, NIP, jenis kelamin, kontak, alamat, dan status keaktifan.
4. `siswa`: Profil peserta didik, NIS (indeks unik), NISN, tempat/tanggal lahir, relasi ke `kelas_id` dan `jurusan_id`.
5. `kurikulum`: Profil staf bidang kurikulum sekolah.
6. `kepala_sekolah`: Profil pimpinan lembaga / kepala sekolah.
7. `tahun_ajaran`: Periode tahun ajaran dan semester aktif.
8. `jurusan`: Konsentrasi keahlian (`kode_jurusan`, `nama_jurusan`, `deskripsi`).
9. `mata_pelajaran`: Kurikulum mapel (`kode_mapel`, `nama_mapel`, kelompok `umum`/`kejuruan`/`muatan_lokal`).
10. `kelas`: Rombel kelas (`nama_kelas`, tingkat `X`/`XI`/`XII`, `jurusan_id`, `wali_kelas_id`).
11. `guru_mengajar`: Pemetaan alokasi pengampu (`guru_id`, `mata_pelajaran_id`, `kelas_id`).
12. `materi`: Modul digital guru (`judul`, `deskripsi`, `file_url`, `tautan_eksternal`).
13. `tugas`: Penugasan guru (`judul`, `deskripsi`, `deadline`, `file_lampiran`).
14. `pengumpulan_tugas`: Submisi berkas siswa, catatan siswa, nilai tugas, dan catatan evaluasi guru.
15. `riwayat_pengumpulan_tugas`: Arsip revisi berkas pengumpulan tugas siswa.
16. `kuis_ujian`: Evaluasi kuis/ujian (`tipe` kuis/ujian, `durasi_menit`, jadwal waktu, status).
17. `soal_kuis_ujian`: Butir pertanyaan soal pilihan ganda, urutan, dan bobot.
18. `pilihan_soal_kuis_ujian`: Opsi jawaban (teks pilihan, `is_benar`).
19. `pengerjaan_kuis_ujian`: Sesi pengerjaan evaluasi oleh siswa dan status pengiriman.
20. `jawaban_kuis_ujian`: Pilihan opsi yang disimpan siswa per butir soal.
21. `nilai_siswa`: Rekapitulasi nilai akhir (`nilai_tugas`, `nilai_kuis`, `nilai_ujian`, `nilai_akhir`, `catatan`).

---

## 6. 🔑 Kredensial Akun Pengujian Bawaan (*Default Seeds*)

Semua akun *dummy* bawaan menggunakan kata sandi standar: **`admin123`**

| Role | Username / Identifier | Email | Password | Hak Akses Utama |
|:---|:---|:---|:---|:---|
| **Admin** | `admin` | `admin@educlass.com` | `admin123` | Kelola seluruh Data Master & Pengguna |
| **Guru** | `budi` | `budi@educlass.com` | `admin123` | Kelola Materi, Tugas, Kuis, Ujian, Nilai |
| **Siswa** | `dewi` *(atau NIS: `112233`)* | `dewi@educlass.com` | `admin123` | Belajar, Unduh Modul, Kirim Tugas, Kuis & Ujian |
| **Kurikulum** | `kurikulum` | `kurikulum@educlass.com` | `admin123` | Alokasi Pengampu, Wali Kelas, Monitoring KBM |
| **Kepala Sekolah** | `kepala_sekolah` | `kepala.sekolah@educlass.com` | `admin123` | Executive Dashboard, Monitoring, Cetak Laporan |

---

## 7. 🚀 Panduan Menjalankan Projek

### 1. Prasyarat Sistem
- **Go**: Versi 1.21 ke atas.
- **Node.js**: Versi 18.x atau 20.x ke atas.
- **MySQL / MariaDB**: Berjalan di port 3306 (melalui Laragon / XAMPP / Native MySQL) dengan database bernama `db_lms`.

### 2. Menjalankan Backend (Go REST API)
```bash
# Masuk ke direktori backend
cd backend_lms

# Menjalankan langsung via Go:
go run main.go
```
*Backend berjalan di: `http://localhost:8080`*  
*Catatan: Backend secara otomatis menjalankan **Auto-Migration** (memastikan database `db_lms` ada, membuat seluruh tabel skema, dan mengisi seed data akun bawaan). Sistem bersifat idempotent sehingga aman dijalankan berkali-kali tanpa risiko duplikasi atau tabel ganda.*

### 3. Menjalankan Frontend (Next.js)
```bash
# Masuk ke direktori frontend
cd frontend_lms

# Install dependensi (jika belum):
npm install

# Menjalankan server pengembangan:
npm run dev
```
*Frontend berjalan di: `http://localhost:3000`*
