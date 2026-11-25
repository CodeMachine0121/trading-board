package vo

import "time"

// AuthorizationPolicyVo holds how long each credential handed out by the sign-in lives.
type AuthorizationPolicyVo struct {
	AuthorizationCodeLifetime time.Duration
	AccessTokenLifetime       time.Duration
	RefreshTokenLifetime      time.Duration
}
