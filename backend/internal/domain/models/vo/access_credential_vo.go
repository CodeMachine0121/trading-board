package vo

// AccessCredentialVo carries the configured owner credentials and agent key; empty means "not configured".
type AccessCredentialVo struct {
	OwnerUsername string
	OwnerPassword string
	AgentApiKey   string
}
