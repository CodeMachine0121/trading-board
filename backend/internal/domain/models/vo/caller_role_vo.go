package vo

// CallerRoleVo is who is calling: the signed-in owner or the agent team holding the API key.
type CallerRoleVo string

const (
	CallerRoleOwner CallerRoleVo = "owner"
	CallerRoleAgent CallerRoleVo = "agent"
)
