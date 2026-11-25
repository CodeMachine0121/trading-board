package models

import "github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"

// TokenRequest is the token endpoint's form body (application/x-www-form-urlencoded, RFC 6749 §4.1.3 / §6).
// A refresh_token grant carries no token here: the refresh token travels only in its HttpOnly cookie.
type TokenRequest struct {
	GrantType    string `form:"grant_type"`
	Code         string `form:"code"`
	CodeVerifier string `form:"code_verifier"`
	ClientID     string `form:"client_id"`
	RedirectUri  string `form:"redirect_uri"`
}

func (tokenRequest TokenRequest) ToCodeExchangeDto() dto.CodeExchangeDto {
	return dto.CodeExchangeDto{
		Code:         tokenRequest.Code,
		CodeVerifier: tokenRequest.CodeVerifier,
		ClientID:     tokenRequest.ClientID,
		RedirectUri:  tokenRequest.RedirectUri,
	}
}
