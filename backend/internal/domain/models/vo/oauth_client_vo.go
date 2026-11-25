package vo

// OAuthClientVo is the one registered public client and the return addresses the operator allows for it.
type OAuthClientVo struct {
	ClientID     string
	RedirectUris []string
}
