package domains

import (
	"crypto/subtle"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
)

// AccessCredentialDomain decides whether presented credentials are the owner's or the agent team's.
type AccessCredentialDomain struct {
	credential vo.AccessCredentialVo
}

func NewAccessCredentialDomain(credential vo.AccessCredentialVo) AccessCredentialDomain {
	return AccessCredentialDomain{credential: credential}
}

// AdmitsOwner never admits anyone while the owner's credentials are unset, so a missing setting cannot open the board.
func (accessCredential AccessCredentialDomain) AdmitsOwner(username string, password string) bool {
	if accessCredential.credential.OwnerUsername == "" || accessCredential.credential.OwnerPassword == "" {
		return false
	}

	usernameMatches := subtle.ConstantTimeCompare([]byte(username), []byte(accessCredential.credential.OwnerUsername))
	passwordMatches := subtle.ConstantTimeCompare([]byte(password), []byte(accessCredential.credential.OwnerPassword))

	return usernameMatches&passwordMatches == 1
}

func (accessCredential AccessCredentialDomain) AdmitsAgent(apiKey string) bool {
	if accessCredential.credential.AgentApiKey == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(apiKey), []byte(accessCredential.credential.AgentApiKey)) == 1
}
