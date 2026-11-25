package models

import "github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"

type ApiKeyRequest struct {
	MemberName string `json:"memberName"`
	Note       string `json:"note"`
}

func (apiKeyRequest ApiKeyRequest) ToCreateDto() dto.ApiKeyCreateDto {
	return dto.ApiKeyCreateDto{MemberName: apiKeyRequest.MemberName, Note: apiKeyRequest.Note}
}
