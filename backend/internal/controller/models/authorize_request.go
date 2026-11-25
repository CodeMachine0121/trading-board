package models

import "github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"

// AuthorizeRequest is the login form: the owner's credentials plus the OAuth authorization parameters.
type AuthorizeRequest struct {
	Email               string `json:"email"`
	Password            string `json:"password"`
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	RedirectUri         string `json:"redirect_uri"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	State               string `json:"state"`
}

func (authorizeRequest AuthorizeRequest) ToAuthorizationRequestDto() dto.AuthorizationRequestDto {
	return dto.AuthorizationRequestDto{
		Email:               authorizeRequest.Email,
		Password:            authorizeRequest.Password,
		ResponseType:        authorizeRequest.ResponseType,
		ClientID:            authorizeRequest.ClientID,
		RedirectUri:         authorizeRequest.RedirectUri,
		CodeChallenge:       authorizeRequest.CodeChallenge,
		CodeChallengeMethod: authorizeRequest.CodeChallengeMethod,
		State:               authorizeRequest.State,
	}
}
