package models

// OAuthErrorResponse is the error body of the OAuth endpoints (RFC 6749 §5.2).
type OAuthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
