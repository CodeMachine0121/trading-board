package models

import "github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"

type SignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (signInRequest SignInRequest) ToSignInDto() dto.SignInDto {
	return dto.SignInDto{Username: signInRequest.Username, Password: signInRequest.Password}
}
