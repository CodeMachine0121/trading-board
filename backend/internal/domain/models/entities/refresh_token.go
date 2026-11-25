package entities

import "time"

// RefreshToken is one link of a sign-in's rotation chain; only its SHA-256 digest is kept. FamilyID is the digest
// of the chain's first token, so every token of one sign-in can be revoked together.
type RefreshToken struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FamilyID    string     `gorm:"size:64;not null;index"`
	TokenDigest string     `gorm:"size:64;not null;uniqueIndex"`
	IssuedAt    time.Time  `gorm:"type:timestamptz;not null"`
	ExpiresAt   time.Time  `gorm:"type:timestamptz;not null"`
	RotatedAt   *time.Time `gorm:"type:timestamptz"`
	RevokedAt   *time.Time `gorm:"type:timestamptz"`
}

func (refreshToken RefreshToken) TableName() string {
	return "RefreshTokens"
}
