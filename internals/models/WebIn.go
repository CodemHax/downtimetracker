package models

type WebIn struct {
	Website string `json:"website" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
}
