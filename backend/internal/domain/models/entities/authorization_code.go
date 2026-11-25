package entities

import "time"

// AuthorizationCode is a one-time OAuth code; only its SHA-256 digest is kept.
type AuthorizationCode struct {
	ID            string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CodeDigest    string     `gorm:"size:64;not null;uniqueIndex"`
	ClientID      string     `gorm:"size:64;not null"`
	RedirectUri   string     `gorm:"type:text;not null"`
	CodeChallenge string     `gorm:"size:128;not null"`
	CreatedAt     time.Time  `gorm:"type:timestamptz;not null"`
	ExpiresAt     time.Time  `gorm:"type:timestamptz;not null"`
	ConsumedAt    *time.Time `gorm:"type:timestamptz"`
}

func (authorizationCode AuthorizationCode) TableName() string {
	return "AuthorizationCodes"
}
