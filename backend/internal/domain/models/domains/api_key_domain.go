package domains

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
)

const (
	// ApiKeyLeadIn makes a member key recognisable at a glance (and to secret scanners).
	ApiKeyLeadIn            = "tbk_"
	memberNameMaximumLength = 40
	noteMaximumLength       = 200
	// lastUsedRefreshInterval keeps "last used" from costing a write on every call.
	lastUsedRefreshInterval = time.Minute
)

// ApiKeyDomain is an agent-team member's key: whether it admits, what the owner sees of it, and its validation
// when it is created.
type ApiKeyDomain struct {
	apiKey   entities.ApiKey
	calendar BusinessCalendarDomain
	key      OpaqueSecretDomain
}

func NewApiKeyDomain(apiKey entities.ApiKey, calendar BusinessCalendarDomain) ApiKeyDomain {
	return ApiKeyDomain{apiKey: apiKey, calendar: calendar}
}

// NewIssuedApiKeyDomain validates a new key for a member; randomPart is the key's secret randomness.
func NewIssuedApiKeyDomain(
	createDto dto.ApiKeyCreateDto, randomPart string, createdAt time.Time, calendar BusinessCalendarDomain,
) (ApiKeyDomain, error) {
	memberName := strings.TrimSpace(createDto.MemberName)
	memberNameLength := utf8.RuneCountInString(memberName)
	if memberNameLength < 1 || memberNameLength > memberNameMaximumLength {
		return ApiKeyDomain{}, ValidationFailed("invalidApiKey", "成員名稱須為 1 到 40 個字")
	}
	note := strings.TrimSpace(createDto.Note)
	if utf8.RuneCountInString(note) > noteMaximumLength {
		return ApiKeyDomain{}, ValidationFailed("invalidApiKey", "備註最多 200 個字")
	}

	key := NewOpaqueSecretDomain(ApiKeyLeadIn + randomPart)

	return ApiKeyDomain{
		apiKey: entities.ApiKey{
			MemberName:    memberName,
			MemberNameKey: strings.ToLower(memberName),
			Note:          note,
			KeyDigest:     key.Digest(),
			DisplayPrefix: key.DisplayPrefix(),
			CreatedAt:     createdAt,
		},
		calendar: calendar,
		key:      key,
	}, nil
}

// MemberNameKey is the member name as uniqueness compares it: letter case ignored.
func (apiKey ApiKeyDomain) MemberNameKey() string {
	return apiKey.apiKey.MemberNameKey
}

func (apiKey ApiKeyDomain) Key() string {
	return apiKey.key.Value()
}

func (apiKey ApiKeyDomain) IsActive() bool {
	return apiKey.apiKey.RevokedAt == nil
}

func (apiKey ApiKeyDomain) NeedsLastUsedRefresh(now time.Time) bool {
	return apiKey.apiKey.LastUsedAt == nil || now.Sub(*apiKey.apiKey.LastUsedAt) >= lastUsedRefreshInterval
}

func (apiKey ApiKeyDomain) ToEntity() entities.ApiKey {
	return apiKey.apiKey
}

// ToCallerDto attributes everything done with this key to its member.
func (apiKey ApiKeyDomain) ToCallerDto() dto.CallerDto {
	return dto.CallerDto{Role: string(vo.CallerRoleAgent), Name: apiKey.apiKey.MemberName}
}

// ToDto shows every time in the business time zone, like every other timestamp the board answers with.
func (apiKey ApiKeyDomain) ToDto() dto.ApiKeyDto {
	location := apiKey.calendar.Location()
	status := vo.ApiKeyStatusActive
	if !apiKey.IsActive() {
		status = vo.ApiKeyStatusRevoked
	}

	return dto.ApiKeyDto{
		ID:            apiKey.apiKey.ID,
		MemberName:    apiKey.apiKey.MemberName,
		Note:          apiKey.apiKey.Note,
		DisplayPrefix: apiKey.apiKey.DisplayPrefix,
		Status:        string(status),
		CreatedAt:     apiKey.apiKey.CreatedAt.In(location),
		LastUsedAt:    apiKey.calendar.InLocation(apiKey.apiKey.LastUsedAt),
		RevokedAt:     apiKey.calendar.InLocation(apiKey.apiKey.RevokedAt),
	}
}
