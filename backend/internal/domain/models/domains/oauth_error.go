package domains

// OAuthError is a refusal on the sign-in endpoints, worded the way OAuth 2.0 answers ({error, error_description});
// it unwraps to the same kinds as BusinessError so the status code is decided in one place.
type OAuthError struct {
	kind        error
	Code        string
	Description string
}

func (oauthError OAuthError) Error() string {
	return oauthError.Code + ": " + oauthError.Description
}

func (oauthError OAuthError) Unwrap() error {
	return oauthError.kind
}

// InvalidRequest covers a malformed or disallowed authorization request, including a cookie request from a stranger.
func InvalidRequest(description string) error {
	return OAuthError{kind: ErrValidation, Code: "invalid_request", Description: description}
}

func InvalidClient(description string) error {
	return OAuthError{kind: ErrValidation, Code: "invalid_client", Description: description}
}

func UnsupportedResponseType(description string) error {
	return OAuthError{kind: ErrValidation, Code: "unsupported_response_type", Description: description}
}

func UnsupportedGrantType(description string) error {
	return OAuthError{kind: ErrValidation, Code: "unsupported_grant_type", Description: description}
}

// InvalidGrant refuses a code or refresh token that is unknown, used, expired, revoked or does not match.
func InvalidGrant(description string) error {
	return OAuthError{kind: ErrValidation, Code: "invalid_grant", Description: description}
}

func AccessDenied(description string) error {
	return OAuthError{kind: ErrAuthenticationRequired, Code: "access_denied", Description: description}
}

func TemporarilyUnavailable(description string) error {
	return OAuthError{kind: ErrTooManyAttempts, Code: "temporarily_unavailable", Description: description}
}

// ForbiddenOrigin refuses a request carrying the refresh cookie that did not come from the board's own web app.
func ForbiddenOrigin(description string) error {
	return OAuthError{kind: ErrForbidden, Code: "invalid_request", Description: description}
}
