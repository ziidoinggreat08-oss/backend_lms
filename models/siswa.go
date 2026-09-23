package models

// Siswa adalah data siswa yang ditampilkan kepada dashboard admin.
type Siswa struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	NIS          string `json:"nis"`
	NISN         string `json:"nisn"`
	NamaLengkap  string `json:"nama_lengkap"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	JenisKelamin string `json:"jenis_kelamin"`
	TempatLahir  string `json:"tempat_lahir"`
	TanggalLahir string `json:"tanggal_lahir"`
	NoTelepon    string `json:"no_telepon"`
	Alamat       string `json:"alamat"`
	KelasID      int    `json:"kelas_id"`
	NamaKelas    string `json:"nama_kelas"`
	JurusanID    int    `json:"jurusan_id"`
	NamaJurusan  string `json:"nama_jurusan"`
	Status       string `json:"status"`
}

// SiswaRequest dipakai ketika admin menambah atau memperbarui siswa.
// Password hanya wajib saat membuat akun baru.
type SiswaRequest struct {
	NIS          string `json:"nis"`
	NISN         string `json:"nisn"`
	NamaLengkap  string `json:"nama_lengkap"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	JenisKelamin string `json:"jenis_kelamin"`
	TempatLahir  string `json:"tempat_lahir"`
	TanggalLahir string `json:"tanggal_lahir"`
	NoTelepon    string `json:"no_telepon"`
	Alamat       string `json:"alamat"`
	KelasID      int    `json:"kelas_id"`
	JurusanID    int    `json:"jurusan_id"`
	Status       string `json:"status"`
}

type KelasOption struct {
	ID          int    `json:"id"`
	NamaKelas   string `json:"nama_kelas"`
	JurusanID   int    `json:"jurusan_id"`
	NamaJurusan string `json:"nama_jurusan"`
}
