package domains

import (
	"crypto/subtle"
	"strings"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
)

// OwnerAccountDomain is the one account: it recognises the owner's email; the password is checked against its hash
// by the password proxy.
type OwnerAccountDomain struct {
	account vo.OwnerAccountVo
}

// NewOwnerAccountDomain normalises the configured email the same way a typed one is normalised.
func NewOwnerAccountDomain(account vo.OwnerAccountVo) OwnerAccountDomain {
	return OwnerAccountDomain{account: vo.OwnerAccountVo{
		Email:        strings.ToLower(strings.TrimSpace(account.Email)),
		PasswordHash: strings.TrimSpace(account.PasswordHash),
	}}
}

func (ownerAccount OwnerAccountDomain) IsConfigured() bool {
	return ownerAccount.account.Email != "" && ownerAccount.account.PasswordHash != ""
}

// MatchesEmail ignores surrounding spaces and letter case, and never matches while the account is not configured.
func (ownerAccount OwnerAccountDomain) MatchesEmail(email string) bool {
	typedEmail := strings.ToLower(strings.TrimSpace(email))

	return ownerAccount.IsConfigured() &&
		subtle.ConstantTimeCompare([]byte(typedEmail), []byte(ownerAccount.account.Email)) == 1
}

func (ownerAccount OwnerAccountDomain) PasswordHash() string {
	return ownerAccount.account.PasswordHash
}

func (ownerAccount OwnerAccountDomain) ToOwnerDto() dto.OwnerDto {
	return dto.OwnerDto{Email: ownerAccount.account.Email}
}
