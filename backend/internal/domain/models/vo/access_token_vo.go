package vo

import "time"

type AccessTokenVo struct {
	AccessToken string
	ExpiresAt   time.Time
}
