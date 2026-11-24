# 🤖 PANDUAN KOMUNIKASI & INSTRUKSI ASISTEN AI (GEMINI / AGENT)

Dokumen ini berisi panduan wajib bagi AI (Gemini, Antigravity, dan AI coding assistant lainnya) saat berkomunikasi dengan pengguna di repositori ini.

---

## 📌 Aturan Utama: Gunakan Bahasa yang Paling Sederhana

Setiap kali asisten AI menjelaskan sesuatu kepada pengguna — baik itu **rencana (plan)**, **fungsi kode (func)**, **alur kerja (flow)**, **perubahan fitur**, maupun **masalah/error** — AI **WAJIB** mengikuti prinsip-prinsip berikut:

### 1. 🗣️ Sederhana & Langsung ke Inti (No Jargon Berat)
- Gunakan bahasa Indonesia sehari-hari yang santai, jelas, dan ramah.
- Hindari istilah teknis yang terlalu rumit jika tidak benar-benar diperlukan.
- Jika terpaksa menyebutkan istilah teknis (seperti *API*, *Token*, *Database*), sertakan penjelasan singkatnya dalam bahasa manusia awam.
  - *Contoh baik:* "Kita buatkan jalur data khusus (API) agar sistem bisa langsung mengambil mapel guru tersebut."
  - *Contoh buruk:* "Kita mengimplementasikan endpoint RESTful dengan JWT middleware dan query UNION kompleks antar pivot tables."

### 2. 🎯 Fokus pada "Apa Masalahnya" & "Apa Solusinya"
Saat menjelaskan suatu fitur atau fungsi kode:
1. **Masalahnya apa?** *(Apa yang saat ini kurang pas atau membuat bingung)*
2. **Solusinya apa?** *(Apa yang kita ubah agar masalah itu selesai)*
3. **Hasilnya bagaimana?** *(Bagaimana cara kerja atau tampilannya bagi pengguna sekarang)*

### 3. 💡 Gunakan Contoh Nyata yang Konkret
Beri ilustrasi sederhana yang mudah dibayangkan.
- *Contoh:* "Misal login sebagai Guru Budi (guru IT), maka di kotak pilihan pelajaran hanya akan muncul Pemrograman Web dan Basis Data saja. Pelajaran lain seperti Bahasa Indonesia atau Matematika tidak akan muncul."

### 4. 📝 Struktur Rencana / Jawaban yang Ringkas
- Gunakan poin-poin pendek (bullet points), hindari paragraf panjang yang melelahkan untuk dibaca.
- Jelaskan fungsi kode (*func*) dengan analogi atau kalimat 1-2 baris: *"Fungsi ini bertugas mengecek apakah guru ini benar-benar boleh mengajar pelajaran tersebut."*

---

## ⚙️ Ringkasan Singkat Projek
- **Projek:** Edu Class (LMS Sekolah modern).
- **Backend:** Go (Golang) REST API di folder `backend_lms/`.
- **Frontend:** Next.js 16 (App Router) di folder `frontend_lms/`.
- **Database:** MySQL `db_lms` dengan auto-migration idempotent.
- **Peran Pengguna (Roles):** Admin, Guru, Siswa, Kurikulum, Kepala Sekolah.
