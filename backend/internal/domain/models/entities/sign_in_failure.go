package entities

import "time"

// SignInFailure is one refused sign-in; the latest few decide whether sign-in is locked.
type SignInFailure struct {
	ID       string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FailedAt time.Time `gorm:"type:timestamptz;not null;index"`
}

func (signInFailure SignInFailure) TableName() string {
	return "SignInFailures"
}
