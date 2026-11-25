package service

import (
	"context"

	domaininterface "github.com/CodeMachine0121/trading-board/backend/internal/domain/interface"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
)

// AuthenticationService tells who a request comes from: an agent-team member's API key or the owner's access token.
type AuthenticationService struct {
	ownerAccount     domains.OwnerAccountDomain
	apiKeyRepository domaininterface.IApiKeyRepository
	accessTokenProxy domaininterface.IAccessTokenProxy
	clockProxy       domaininterface.IClockProxy
	calendar         domains.BusinessCalendarDomain
}

func NewAuthenticationService(
	ownerAccount domains.OwnerAccountDomain,
	apiKeyRepository domaininterface.IApiKeyRepository,
	accessTokenProxy domaininterface.IAccessTokenProxy,
	clockProxy domaininterface.IClockProxy,
	calendar domains.BusinessCalendarDomain,
) *AuthenticationService {
	return &AuthenticationService{
		ownerAccount:     ownerAccount,
		apiKeyRepository: apiKeyRepository,
		accessTokenProxy: accessTokenProxy,
		clockProxy:       clockProxy,
		calendar:         calendar,
	}
}

// IdentifyCaller admits an active API key first and the owner's access token second; anything else is not
// authenticated. Using a key brings its "last used" up to date, at most once a minute.
func (authenticationService *AuthenticationService) IdentifyCaller(
	executionContext context.Context, credentialDto dto.CallerCredentialDto,
) (dto.CallerDto, error) {
	if credentialDto.ApiKey != "" {
		stored, isFound, findError := authenticationService.apiKeyRepository.FindByDigest(
			executionContext, domains.NewOpaqueSecretDomain(credentialDto.ApiKey).Digest())
		if findError != nil {
			return dto.CallerDto{}, findError
		}
		apiKey := domains.NewApiKeyDomain(stored, authenticationService.calendar)
		if isFound && apiKey.IsActive() {
			now := authenticationService.clockProxy.Now()
			if apiKey.NeedsLastUsedRefresh(now) {
				if touchError := authenticationService.apiKeyRepository.TouchLastUsed(
					executionContext, stored.ID, now); touchError != nil {
					return dto.CallerDto{}, touchError
				}
			}

			return apiKey.ToCallerDto(), nil
		}
	}

	if credentialDto.AccessToken != "" {
		claims, verifyError := authenticationService.accessTokenProxy.Verify(credentialDto.AccessToken)
		if verifyError == nil && claims.Subject == string(vo.CallerRoleOwner) {
			return dto.CallerDto{Role: string(vo.CallerRoleOwner), Name: vo.OwnerActorName}, nil
		}
	}

	return dto.CallerDto{}, domains.AuthenticationRequired("請先登入，或帶上有效的 API 金鑰")
}

func (authenticationService *AuthenticationService) CurrentOwner() dto.OwnerDto {
	return authenticationService.ownerAccount.ToOwnerDto()
}
