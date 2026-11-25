package vo

// CallerRoleVo is who is calling: the signed-in owner or an agent-team member holding its own API key.
type CallerRoleVo string

const (
	CallerRoleOwner CallerRoleVo = "owner"
	CallerRoleAgent CallerRoleVo = "agent"
)

// OwnerActorName is what a record made by the signed-in owner says about who made it.
const OwnerActorName = "owner"
