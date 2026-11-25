package service

import (
	"context"
	"time"

	domaininterface "github.com/CodeMachine0121/trading-board/backend/internal/domain/interface"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
)

// AuthenticationService signs the owner in and tells who a request comes from.
type AuthenticationService struct {
	accessCredential domains.AccessCredentialDomain
	accessTokenTTL   time.Duration
	accessTokenProxy domaininterface.IAccessTokenProxy
	clockProxy       domaininterface.IClockProxy
	calendar         domains.BusinessCalendarDomain
}

func NewAuthenticationService(
	accessCredential domains.AccessCredentialDomain,
	accessTokenTTL time.Duration,
	accessTokenProxy domaininterface.IAccessTokenProxy,
	clockProxy domaininterface.IClockProxy,
	calendar domains.BusinessCalendarDomain,
) *AuthenticationService {
	return &AuthenticationService{
		accessCredential: accessCredential,
		accessTokenTTL:   accessTokenTTL,
		accessTokenProxy: accessTokenProxy,
		clockProxy:       clockProxy,
		calendar:         calendar,
	}
}

// SignIn says only "not authenticated" on a mismatch, never which of the two was wrong.
func (authenticationService *AuthenticationService) SignIn(
	_ context.Context, signInDto dto.SignInDto,
) (dto.AccessTokenDto, error) {
	if !authenticationService.accessCredential.AdmitsOwner(signInDto.Username, signInDto.Password) {
		return dto.AccessTokenDto{}, domains.AuthenticationRequired("帳號或密碼錯誤")
	}

	accessToken, issueError := authenticationService.accessTokenProxy.Issue(vo.AccessTokenClaimsVo{
		Subject:   string(vo.CallerRoleOwner),
		ExpiresAt: authenticationService.clockProxy.Now().Add(authenticationService.accessTokenTTL),
	})
	if issueError != nil {
		return dto.AccessTokenDto{}, issueError
	}

	// Shown in the business time zone, like every other timestamp the board answers with.
	return dto.AccessTokenDto{
		AccessToken: accessToken.AccessToken,
		ExpiresAt:   accessToken.ExpiresAt.In(authenticationService.calendar.Location()),
	}, nil
}

// IdentifyCaller admits the agent key first and the owner's pass second; anything else is not authenticated.
func (authenticationService *AuthenticationService) IdentifyCaller(
	_ context.Context, credentialDto dto.CallerCredentialDto,
) (dto.CallerDto, error) {
	if credentialDto.ApiKey != "" && authenticationService.accessCredential.AdmitsAgent(credentialDto.ApiKey) {
		return dto.CallerDto{Role: string(vo.CallerRoleAgent)}, nil
	}

	if credentialDto.AccessToken != "" {
		claims, verifyError := authenticationService.accessTokenProxy.Verify(credentialDto.AccessToken)
		if verifyError == nil && claims.Subject == string(vo.CallerRoleOwner) {
			return dto.CallerDto{Role: string(vo.CallerRoleOwner)}, nil
		}
	}

	return dto.CallerDto{}, domains.AuthenticationRequired("請先登入，或帶上有效的 API 金鑰")
}
