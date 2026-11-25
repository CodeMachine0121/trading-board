package vo

import "time"

// AccessTokenClaimsVo is what a pass asserts: whose it is and until when.
type AccessTokenClaimsVo struct {
	Subject   string
	ExpiresAt time.Time
}
