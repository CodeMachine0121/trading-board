package service

import (
	"context"
	"time"

	domaininterface "github.com/CodeMachine0121/trading-board/backend/internal/domain/interface"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
)

const (
	wrongCredentialsDescription = "電子郵件或密碼錯誤"
	lockedDescription           = "嘗試次數過多，請 15 分鐘後再試"
	signInEndedDescription      = "登入已失效，請重新登入"
	bearerTokenType             = "Bearer"
)

// AuthorizationService is the board's own OAuth 2.0 authorization server for its one public client.
//
// Each use case runs inside ITransactionRepository.Exclusively, so "use a code once", "rotate a refresh token once"
// and the sign-in lock cannot race. Some refusals must still keep what they wrote (a recorded failure, a used-up
// code, a revoked family): those are carried out of the transaction as the refusal value instead of being returned
// from the work, which would roll the write back.
type AuthorizationService struct {
	ownerAccount          domains.OwnerAccountDomain
	client                vo.OAuthClientVo
	policy                vo.AuthorizationPolicyVo
	authorizationCodes    domaininterface.IAuthorizationCodeRepository
	refreshTokens         domaininterface.IRefreshTokenRepository
	signInFailures        domaininterface.ISignInFailureRepository
	transactionRepository domaininterface.ITransactionRepository
	passwordHashProxy     domaininterface.IPasswordHashProxy
	randomTokenProxy      domaininterface.IRandomTokenProxy
	accessTokenProxy      domaininterface.IAccessTokenProxy
	clockProxy            domaininterface.IClockProxy
}

func NewAuthorizationService(
	ownerAccount domains.OwnerAccountDomain,
	client vo.OAuthClientVo,
	policy vo.AuthorizationPolicyVo,
	authorizationCodes domaininterface.IAuthorizationCodeRepository,
	refreshTokens domaininterface.IRefreshTokenRepository,
	signInFailures domaininterface.ISignInFailureRepository,
	transactionRepository domaininterface.ITransactionRepository,
	passwordHashProxy domaininterface.IPasswordHashProxy,
	randomTokenProxy domaininterface.IRandomTokenProxy,
	accessTokenProxy domaininterface.IAccessTokenProxy,
	clockProxy domaininterface.IClockProxy,
) *AuthorizationService {
	return &AuthorizationService{
		ownerAccount:          ownerAccount,
		client:                client,
		policy:                policy,
		authorizationCodes:    authorizationCodes,
		refreshTokens:         refreshTokens,
		signInFailures:        signInFailures,
		transactionRepository: transactionRepository,
		passwordHashProxy:     passwordHashProxy,
		randomTokenProxy:      randomTokenProxy,
		accessTokenProxy:      accessTokenProxy,
		clockProxy:            clockProxy,
	}
}

// Authorize checks the request before the credentials, so a bad client or return address is never redirected to.
// The password hash is checked even for an unknown email, so the answer takes as long either way.
func (authorizationService *AuthorizationService) Authorize(
	executionContext context.Context, requestDto dto.AuthorizationRequestDto,
) (dto.AuthorizationRedirectDto, error) {
	authorizationRequest, requestError := domains.NewAuthorizationRequestDomain(requestDto, authorizationService.client)
	if requestError != nil {
		return dto.AuthorizationRedirectDto{}, requestError
	}

	var redirectDto dto.AuthorizationRedirectDto
	var refusal error
	workError := authorizationService.transactionRepository.Exclusively(executionContext,
		func(transactionContext context.Context) error {
			now := authorizationService.clockProxy.Now()
			latestFailures, listError := authorizationService.signInFailures.ListLatest(
				transactionContext, domains.SignInFailuresConsidered)
			if listError != nil {
				return listError
			}
			if domains.NewSignInThrottleDomain(latestFailures).IsLockedAt(now) {
				refusal = domains.TemporarilyUnavailable(lockedDescription)
				return nil
			}

			emailMatches := authorizationService.ownerAccount.MatchesEmail(requestDto.Email)
			passwordMatches := authorizationService.passwordHashProxy.Matches(
				authorizationService.ownerAccount.PasswordHash(), requestDto.Password)
			if !emailMatches || !passwordMatches {
				refusal = domains.AccessDenied(wrongCredentialsDescription)
				return authorizationService.signInFailures.Record(transactionContext, now)
			}
			if clearError := authorizationService.signInFailures.Clear(transactionContext); clearError != nil {
				return clearError
			}

			code := domains.NewOpaqueSecretDomain(authorizationService.randomTokenProxy.NewToken())
			if createError := authorizationService.authorizationCodes.Create(transactionContext,
				authorizationRequest.ToAuthorizationCode(code, now, authorizationService.policy.AuthorizationCodeLifetime),
			); createError != nil {
				return createError
			}
			redirectDto = authorizationRequest.ToRedirectDto(code)

			return nil
		})
	if workError != nil {
		return dto.AuthorizationRedirectDto{}, workError
	}
	if refusal != nil {
		return dto.AuthorizationRedirectDto{}, refusal
	}

	return redirectDto, nil
}

// ExchangeCode uses the code up before judging it, so a wrong verifier cannot be retried with another guess.
func (authorizationService *AuthorizationService) ExchangeCode(
	executionContext context.Context, exchangeDto dto.CodeExchangeDto,
) (dto.TokenSetDto, error) {
	var tokenSet dto.TokenSetDto
	var refusal error
	workError := authorizationService.transactionRepository.Exclusively(executionContext,
		func(transactionContext context.Context) error {
			now := authorizationService.clockProxy.Now()
			code, isFound, findError := authorizationService.authorizationCodes.FindByDigest(
				transactionContext, domains.NewOpaqueSecretDomain(exchangeDto.Code).Digest())
			if findError != nil {
				return findError
			}
			authorizationCode := domains.NewAuthorizationCodeDomain(code)
			if exchangeDto.Code == "" || !isFound || authorizationCode.IsConsumed() {
				refusal = domains.InvalidGrant("授權碼無效或已過期")
				return nil
			}
			if markError := authorizationService.authorizationCodes.MarkConsumed(transactionContext, code.ID, now); markError != nil {
				return markError
			}
			if refusal = authorizationCode.Redeem(exchangeDto, now); refusal != nil {
				return nil
			}

			issued, issueError := authorizationService.issueTokens(transactionContext, "", now)
			tokenSet = issued

			return issueError
		})
	if workError != nil {
		return dto.TokenSetDto{}, workError
	}
	if refusal != nil {
		return dto.TokenSetDto{}, refusal
	}

	return tokenSet, nil
}

// Refresh rotates the refresh token; presenting one that was already rotated revokes the whole sign-in.
func (authorizationService *AuthorizationService) Refresh(
	executionContext context.Context, refreshToken string,
) (dto.TokenSetDto, error) {
	if refreshToken == "" {
		return dto.TokenSetDto{}, domains.InvalidGrant(signInEndedDescription)
	}

	var tokenSet dto.TokenSetDto
	var refusal error
	workError := authorizationService.transactionRepository.Exclusively(executionContext,
		func(transactionContext context.Context) error {
			now := authorizationService.clockProxy.Now()
			stored, isFound, findError := authorizationService.refreshTokens.FindByDigest(
				transactionContext, domains.NewOpaqueSecretDomain(refreshToken).Digest())
			if findError != nil {
				return findError
			}
			if !isFound {
				refusal = domains.InvalidGrant(signInEndedDescription)
				return nil
			}

			presented := domains.NewRefreshTokenDomain(stored)
			if presented.IsReplaced() {
				refusal = domains.InvalidGrant(signInEndedDescription)
				return authorizationService.refreshTokens.RevokeFamily(transactionContext, presented.FamilyID(), now)
			}
			if !presented.AdmitsAt(now) {
				refusal = domains.InvalidGrant(signInEndedDescription)
				return nil
			}
			if markError := authorizationService.refreshTokens.MarkRotated(transactionContext, stored.ID, now); markError != nil {
				return markError
			}

			issued, issueError := authorizationService.issueTokens(transactionContext, presented.FamilyID(), now)
			tokenSet = issued

			return issueError
		})
	if workError != nil {
		return dto.TokenSetDto{}, workError
	}
	if refusal != nil {
		return dto.TokenSetDto{}, refusal
	}

	return tokenSet, nil
}

// Revoke signs out (RFC 7009): an empty or unknown token is answered as done, since there is nothing to end.
func (authorizationService *AuthorizationService) Revoke(executionContext context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	return authorizationService.transactionRepository.Exclusively(executionContext,
		func(transactionContext context.Context) error {
			stored, isFound, findError := authorizationService.refreshTokens.FindByDigest(
				transactionContext, domains.NewOpaqueSecretDomain(refreshToken).Digest())
			if findError != nil || !isFound {
				return findError
			}

			return authorizationService.refreshTokens.RevokeFamily(
				transactionContext, stored.FamilyID, authorizationService.clockProxy.Now())
		})
}

// issueTokens hands out an access token and the next refresh token of the family (a new family when empty).
func (authorizationService *AuthorizationService) issueTokens(
	transactionContext context.Context, familyID string, now time.Time,
) (dto.TokenSetDto, error) {
	accessToken, issueError := authorizationService.accessTokenProxy.Issue(vo.AccessTokenClaimsVo{
		Subject:   string(vo.CallerRoleOwner),
		ExpiresAt: now.Add(authorizationService.policy.AccessTokenLifetime),
	})
	if issueError != nil {
		return dto.TokenSetDto{}, issueError
	}

	refreshToken := domains.NewOpaqueSecretDomain(authorizationService.randomTokenProxy.NewToken())
	storedToken := refreshToken.ToRefreshToken(familyID, now, authorizationService.policy.RefreshTokenLifetime)
	if createError := authorizationService.refreshTokens.Create(transactionContext, storedToken); createError != nil {
		return dto.TokenSetDto{}, createError
	}

	return dto.TokenSetDto{
		AccessToken:           accessToken.AccessToken,
		TokenType:             bearerTokenType,
		ExpiresIn:             int(authorizationService.policy.AccessTokenLifetime / time.Second),
		RefreshToken:          refreshToken.Value(),
		RefreshTokenExpiresAt: storedToken.ExpiresAt,
	}, nil
}
