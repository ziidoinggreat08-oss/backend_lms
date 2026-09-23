package models

// User merepresentasikan struktur data dari tabel users
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"` // tidak dikirim ke JSON
	Role     string `json:"role"`
}

// LoginRequest merepresentasikan payload dari React saat proses login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse merepresentasikan data JSON yang dikembalikan ke React
type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
	Role    string `json:"role"`
	User    User   `json:"user"`
}
