package dto

// CallerCredentialDto is whatever proof a request carried: the agent key, the owner's pass, both or neither.
type CallerCredentialDto struct {
	ApiKey      string
	AccessToken string
}
