package entities

import "time"

// ApiKey is one agent-team member's key; only its SHA-256 digest and a short display prefix are kept.
// MemberNameKey is the lower-cased member name: unique among active keys, so a revoked name can be reused.
type ApiKey struct {
	ID            string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MemberName    string     `gorm:"size:160;not null"`
	MemberNameKey string     `gorm:"size:160;not null;uniqueIndex:idx_api_keys_active_member_name,where:revoked_at IS NULL"`
	Note          string     `gorm:"type:text;not null;default:''"`
	KeyDigest     string     `gorm:"size:64;not null;uniqueIndex"`
	DisplayPrefix string     `gorm:"size:16;not null"`
	CreatedAt     time.Time  `gorm:"type:timestamptz;not null;index"`
	LastUsedAt    *time.Time `gorm:"type:timestamptz"`
	RevokedAt     *time.Time `gorm:"type:timestamptz"`
}

func (apiKey ApiKey) TableName() string {
	return "ApiKeys"
}
