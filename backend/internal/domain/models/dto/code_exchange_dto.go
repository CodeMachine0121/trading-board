package dto

// CodeExchangeDto is the authorization_code grant: the code, the PKCE verifier and what the code was bound to.
type CodeExchangeDto struct {
	Code         string
	CodeVerifier string
	ClientID     string
	RedirectUri  string
}
