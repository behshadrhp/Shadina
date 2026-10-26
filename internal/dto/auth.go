package dto

type GenerateTokenPayload struct {
	Email    string `json:"email" validate:"required,max=255"`
	Password string `json:"password" validate:"required,min=8,max=255"`
}
