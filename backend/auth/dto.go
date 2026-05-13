package auth

type RegisterDTO struct {
	Username string `json:"username" validate:"required,min=3,max=60"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

type LoginDTO struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
