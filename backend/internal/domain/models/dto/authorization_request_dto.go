package dto

// AuthorizationRequestDto is the owner's credentials together with the web app's OAuth authorization parameters.
type AuthorizationRequestDto struct {
	Email               string
	Password            string
	ResponseType        string
	ClientID            string
	RedirectUri         string
	CodeChallenge       string
	CodeChallengeMethod string
	State               string
}
