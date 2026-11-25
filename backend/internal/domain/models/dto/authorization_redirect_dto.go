package dto

// AuthorizationRedirectDto is where the web app sends the owner after a successful sign-in: the return address
// carrying the one-time code and the web app's state.
type AuthorizationRedirectDto struct {
	RedirectTo string `json:"redirectTo"`
}
