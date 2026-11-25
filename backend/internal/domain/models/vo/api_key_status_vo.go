package vo

// ApiKeyStatusVo is whether a member key still admits its member.
type ApiKeyStatusVo string

const (
	ApiKeyStatusActive  ApiKeyStatusVo = "active"
	ApiKeyStatusRevoked ApiKeyStatusVo = "revoked"
)
