package dto

import "time"

// TokenSetDto is what a successful grant hands out. The refresh token never goes into the JSON body:
// the controller puts it in the HttpOnly cookie, out of reach of page scripts.
type TokenSetDto struct {
	AccessToken           string    `json:"access_token"`
	TokenType             string    `json:"token_type"`
	ExpiresIn             int       `json:"expires_in"`
	RefreshToken          string    `json:"-"`
	RefreshTokenExpiresAt time.Time `json:"-"`
}
