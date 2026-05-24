package models

type AuthIn struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthOut struct {
	Token   string `json:"token,omitempty"`
	Message string `json:"message"`
}

type ForgotPasswordIn struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordIn struct {
	Email    string `json:"email" binding:"required,email"`
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}
