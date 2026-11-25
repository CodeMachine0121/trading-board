package dto

import "time"

// ApiKeyDto is a member key as the owner sees it in the list; the key itself is never part of it.
type ApiKeyDto struct {
	ID            string     `json:"id"`
	MemberName    string     `json:"memberName"`
	Note          string     `json:"note"`
	DisplayPrefix string     `json:"displayPrefix"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastUsedAt    *time.Time `json:"lastUsedAt"`
	RevokedAt     *time.Time `json:"revokedAt"`
}
