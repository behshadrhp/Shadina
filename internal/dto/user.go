package dto

type UserRegisterPayload struct {
	Email    string `json:"email" validate:"required,max=255"`
	Username string `json:"username" validate:"required,max=255"`
	Password string `json:"password" validate:"required,min=8,max=255"`
}
