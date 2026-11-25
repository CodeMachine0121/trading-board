package application_test

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	rightPassword = "right-password"
	codeVerifier  = "a-code-verifier-long-enough-for-pkce-0123456789"
	thirtyDays    = 30 * 24 * time.Hour
)

func challengeOf(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))

	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func anAuthorizationRequest() dto.AuthorizationRequestDto {
	return dto.AuthorizationRequestDto{
		Email: "owner@example.com", Password: rightPassword, ResponseType: "code", ClientID: clientID,
		RedirectUri: redirectUri, CodeChallenge: challengeOf(codeVerifier), CodeChallengeMethod: "S256", State: "xyz",
	}
}

func failuresAt(instants ...string) []entities.SignInFailure {
	failures := []entities.SignInFailure{}
	for _, instant := range instants {
		failures = append(failures, entities.SignInFailure{FailedAt: taipeiTimeWithSeconds(instant)})
	}

	return failures
}

// checksPasswords makes the password proxy accept exactly rightPassword against the configured hash.
func (testHarness *harness) checksPasswords() *harness {
	testHarness.passwordHashProxy.EXPECT().Matches(ownerPasswordHash, gomock.Any()).DoAndReturn(
		func(_ string, password string) bool { return password == rightPassword }).AnyTimes()

	return testHarness
}

func requireOAuthError(t *testing.T, err error, expectedCode string, expectedDescription string) {
	t.Helper()
	var oauthError domains.OAuthError
	require.True(t, errors.As(err, &oauthError), "expected an OAuth error, got %v", err)
	assert.Equal(t, expectedCode, oauthError.Code)
	assert.Equal(t, expectedDescription, oauthError.Description)
}

func TestAuthorizationApplicationSignsInAndHandsBackABoundOneTimeCode(t *testing.T) {
	now := taipeiTimeWithSeconds("2026-11-25 20:00:00")
	testHarness := newHarnessAt(t, now).checksPasswords()
	testHarness.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, nil)
	testHarness.failureRepository.EXPECT().Clear(gomock.Any()).Return(nil)
	testHarness.randomTokenProxy.EXPECT().NewToken().Return("code-secret")
	testHarness.codeRepository.EXPECT().Create(gomock.Any(), entities.AuthorizationCode{
		CodeDigest: digestOf("code-secret"), ClientID: clientID, RedirectUri: redirectUri,
		CodeChallenge: challengeOf(codeVerifier), CreatedAt: now, ExpiresAt: now.Add(60 * time.Second),
	}).Return(nil)
	request := anAuthorizationRequest()
	request.Email = "  Owner@Example.COM "

	redirectDto, authorizeError := testHarness.authorizationApplication.Authorize(t.Context(), request)

	require.NoError(t, authorizeError)
	assert.Equal(t, "http://localhost:3000/auth/callback?code=code-secret&state=xyz", redirectDto.RedirectTo)
}

func TestAuthorizationApplicationRefusesWrongCredentialsAlike(t *testing.T) {
	cases := []struct {
		name     string
		email    string
		password string
	}{
		{name: "the right email and a wrong password", email: "owner@example.com", password: "wrong"},
		{name: "an unknown email and the right password", email: "stranger@example.com", password: rightPassword},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			now := taipeiTimeWithSeconds("2026-11-25 20:00:00")
			testHarness := newHarnessAt(t, now)
			testHarness.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, nil)
			// The hash is checked even for an unknown email, so both refusals take as long.
			testHarness.passwordHashProxy.EXPECT().Matches(ownerPasswordHash, testCase.password).
				Return(testCase.password == rightPassword)
			testHarness.failureRepository.EXPECT().Record(gomock.Any(), now).Return(nil)
			request := anAuthorizationRequest()
			request.Email, request.Password = testCase.email, testCase.password

			redirectDto, authorizeError := testHarness.authorizationApplication.Authorize(t.Context(), request)

			requireOAuthError(t, authorizeError, "access_denied", "電子郵件或密碼錯誤")
			assert.ErrorIs(t, authorizeError, domains.ErrAuthenticationRequired)
			assert.Zero(t, testHarness.rolledBack, "the failure stays recorded: the refusal must not roll it back")
			assert.Empty(t, redirectDto.RedirectTo)
		})
	}
}

func TestAuthorizationApplicationAdmitsNobodyWhileTheOwnerAccountIsNotSet(t *testing.T) {
	cases := []struct {
		name    string
		account vo.OwnerAccountVo
	}{
		{name: "no email set", account: vo.OwnerAccountVo{PasswordHash: ownerPasswordHash}},
		{name: "no password hash set", account: vo.OwnerAccountVo{Email: ownerEmail}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00").withOwnerAccount(testCase.account)
			testHarness.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, nil)
			// Even a proxy that would accept the password cannot open an unconfigured account.
			testHarness.passwordHashProxy.EXPECT().Matches(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
			testHarness.failureRepository.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)
			request := anAuthorizationRequest()
			request.Email = ""
			if testCase.account.Email != "" {
				request.Email = ownerEmail
			}

			_, authorizeError := testHarness.authorizationApplication.Authorize(t.Context(), request)

			requireOAuthError(t, authorizeError, "access_denied", "電子郵件或密碼錯誤")
		})
	}
}

func TestAuthorizationApplicationLocksSignInAfterFiveFailuresWithinFifteenMinutes(t *testing.T) {
	fiveFailuresEndingAt2000 := failuresAt("2026-11-25 20:00:00", "2026-11-25 19:58:00", "2026-11-25 19:55:00",
		"2026-11-25 19:50:00", "2026-11-25 19:46:00")

	cases := []struct {
		name          string
		failures      []entities.SignInFailure
		now           string
		password      string
		expectsLocked bool
	}{
		{name: "the fifth failure locks: the right password at 20:10 is refused", failures: fiveFailuresEndingAt2000,
			now: "2026-11-25 20:10:00", password: rightPassword, expectsLocked: true},
		{name: "the lock ends at 20:15", failures: fiveFailuresEndingAt2000,
			now: "2026-11-25 20:15:00", password: rightPassword},
		{name: "five failures 16 minutes apart do not lock", failures: failuresAt("2026-11-25 20:00:00",
			"2026-11-25 19:58:00", "2026-11-25 19:55:00", "2026-11-25 19:50:00", "2026-11-25 19:44:00"),
			now: "2026-11-25 20:00:30", password: rightPassword},
		{name: "two failures after a success (which cleared three) do not lock",
			failures: failuresAt("2026-11-25 20:00:00", "2026-11-25 19:59:00"),
			now:      "2026-11-25 20:00:30", password: rightPassword},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarnessAt(t, taipeiTimeWithSeconds(testCase.now))
			testHarness.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(testCase.failures, nil)
			if !testCase.expectsLocked {
				testHarness.checksPasswords()
				testHarness.failureRepository.EXPECT().Clear(gomock.Any()).Return(nil)
				testHarness.randomTokenProxy.EXPECT().NewToken().Return("code-secret")
				testHarness.codeRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			}

			redirectDto, authorizeError := testHarness.authorizationApplication.Authorize(t.Context(), anAuthorizationRequest())

			if testCase.expectsLocked {
				requireOAuthError(t, authorizeError, "temporarily_unavailable", "嘗試次數過多，請 15 分鐘後再試")
				assert.ErrorIs(t, authorizeError, domains.ErrTooManyAttempts)
				return
			}
			require.NoError(t, authorizeError)
			assert.Contains(t, redirectDto.RedirectTo, "code=code-secret")
		})
	}
}

func TestAuthorizationApplicationRefusesBadRequestsBeforeLookingAtCredentials(t *testing.T) {
	cases := []struct {
		name                string
		change              func(request *dto.AuthorizationRequestDto)
		expectedCode        string
		expectedDescription string
	}{
		{name: "an unknown application", change: func(request *dto.AuthorizationRequestDto) { request.ClientID = "evil-app" },
			expectedCode: "invalid_client", expectedDescription: "不明的應用程式"},
		{name: "a return address not on the allowed list",
			change:       func(request *dto.AuthorizationRequestDto) { request.RedirectUri = "https://evil.example/callback" },
			expectedCode: "invalid_request", expectedDescription: "不允許的返回網址"},
		{name: "the weak proof method",
			change:       func(request *dto.AuthorizationRequestDto) { request.CodeChallengeMethod = "plain" },
			expectedCode: "invalid_request", expectedDescription: "必須使用 S256 驗證"},
		{name: "no proof at all", change: func(request *dto.AuthorizationRequestDto) {
			request.CodeChallenge, request.CodeChallengeMethod = "", ""
		}, expectedCode: "invalid_request", expectedDescription: "必須使用 S256 驗證"},
		{name: "a challenge that is not a S256 digest",
			change:       func(request *dto.AuthorizationRequestDto) { request.CodeChallenge = "too-short" },
			expectedCode: "invalid_request", expectedDescription: "必須使用 S256 驗證"},
		{name: "a response type other than code",
			change:       func(request *dto.AuthorizationRequestDto) { request.ResponseType = "token" },
			expectedCode: "unsupported_response_type", expectedDescription: "只支援 code"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			request := anAuthorizationRequest()
			testCase.change(&request)

			redirectDto, authorizeError := testHarness.authorizationApplication.Authorize(t.Context(), request)

			requireOAuthError(t, authorizeError, testCase.expectedCode, testCase.expectedDescription)
			assert.ErrorIs(t, authorizeError, domains.ErrValidation)
			assert.Empty(t, redirectDto.RedirectTo)
		})
	}
}

func TestAuthorizationApplicationKeepsTheStateOutWhenTheWebAppSentNone(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00").checksPasswords()
	testHarness.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, nil)
	testHarness.failureRepository.EXPECT().Clear(gomock.Any()).Return(nil)
	testHarness.randomTokenProxy.EXPECT().NewToken().Return("code-secret")
	testHarness.codeRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	request := anAuthorizationRequest()
	request.State = ""

	redirectDto, authorizeError := testHarness.authorizationApplication.Authorize(t.Context(), request)

	require.NoError(t, authorizeError)
	assert.Equal(t, "http://localhost:3000/auth/callback?code=code-secret", redirectDto.RedirectTo)
}

func TestAuthorizationApplicationPassesOnFailuresWhileSigningIn(t *testing.T) {
	cases := []struct {
		name    string
		arrange func(testHarness *harness)
	}{
		{name: "reading the failures", arrange: func(testHarness *harness) {
			testHarness.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, errStorage)
		}},
		{name: "recording a failure", arrange: func(testHarness *harness) {
			testHarness.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, nil)
			testHarness.passwordHashProxy.EXPECT().Matches(gomock.Any(), gomock.Any()).Return(false)
			testHarness.failureRepository.EXPECT().Record(gomock.Any(), gomock.Any()).Return(errStorage)
		}},
		{name: "clearing the failures", arrange: func(testHarness *harness) {
			testHarness.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, nil)
			testHarness.checksPasswords()
			testHarness.failureRepository.EXPECT().Clear(gomock.Any()).Return(errStorage)
		}},
		{name: "storing the code", arrange: func(testHarness *harness) {
			testHarness.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, nil)
			testHarness.checksPasswords()
			testHarness.failureRepository.EXPECT().Clear(gomock.Any()).Return(nil)
			testHarness.randomTokenProxy.EXPECT().NewToken().Return("code-secret")
			testHarness.codeRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errStorage)
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			testCase.arrange(testHarness)

			_, authorizeError := testHarness.authorizationApplication.Authorize(t.Context(), anAuthorizationRequest())

			assert.ErrorIs(t, authorizeError, errStorage)
		})
	}
}

// aCodeIssuedAgo is a stored, unused code bound to the test client, return address and verifier.
func aCodeIssuedAgo(now time.Time, age time.Duration) entities.AuthorizationCode {
	return entities.AuthorizationCode{ID: "code-1", CodeDigest: digestOf("code-secret"), ClientID: clientID,
		RedirectUri: redirectUri, CodeChallenge: challengeOf(codeVerifier),
		CreatedAt: now.Add(-age), ExpiresAt: now.Add(-age).Add(60 * time.Second)}
}

func anExchange() dto.CodeExchangeDto {
	return dto.CodeExchangeDto{Code: "code-secret", CodeVerifier: codeVerifier, ClientID: clientID, RedirectUri: redirectUri}
}

// issuesTokens expects a 15-minute access token and the next refresh token of familyID (a new family when empty).
func (testHarness *harness) issuesTokens(now time.Time, familyID string) {
	testHarness.accessTokenProxy.EXPECT().Issue(vo.AccessTokenClaimsVo{Subject: "owner", ExpiresAt: now.Add(15 * time.Minute)}).
		Return(vo.AccessTokenVo{AccessToken: "signed-access-token", ExpiresAt: now.Add(15 * time.Minute)}, nil)
	testHarness.randomTokenProxy.EXPECT().NewToken().Return("refresh-secret")
	if familyID == "" {
		familyID = digestOf("refresh-secret")
	}
	testHarness.refreshRepository.EXPECT().Create(gomock.Any(), entities.RefreshToken{
		FamilyID: familyID, TokenDigest: digestOf("refresh-secret"), IssuedAt: now, ExpiresAt: now.Add(thirtyDays),
	}).Return(nil)
}

func TestAuthorizationApplicationTradesAFreshCodeForTokens(t *testing.T) {
	now := taipeiTimeWithSeconds("2026-11-25 20:00:20")
	testHarness := newHarnessAt(t, now)
	testHarness.codeRepository.EXPECT().FindByDigest(gomock.Any(), digestOf("code-secret")).
		Return(aCodeIssuedAgo(now, 20*time.Second), true, nil)
	testHarness.codeRepository.EXPECT().MarkConsumed(gomock.Any(), "code-1", now).Return(nil)
	testHarness.issuesTokens(now, "")

	tokenSet, exchangeError := testHarness.authorizationApplication.ExchangeCode(t.Context(), anExchange())

	require.NoError(t, exchangeError)
	assert.Equal(t, dto.TokenSetDto{AccessToken: "signed-access-token", TokenType: "Bearer", ExpiresIn: 900,
		RefreshToken: "refresh-secret", RefreshTokenExpiresAt: now.Add(thirtyDays)}, tokenSet)
}

func TestAuthorizationApplicationRefusesCodesThatDoNotRedeem(t *testing.T) {
	cases := []struct {
		name         string
		age          time.Duration
		change       func(exchange *dto.CodeExchangeDto)
		alreadyUsed  bool
		expectsUseUp bool
	}{
		{name: "a code 61 seconds old", age: 61 * time.Second, expectsUseUp: true},
		{name: "a code exactly 60 seconds old", age: 60 * time.Second, expectsUseUp: true},
		{name: "a code already traded in", age: 10 * time.Second, alreadyUsed: true},
		{name: "a verifier that does not match", age: 10 * time.Second, expectsUseUp: true,
			change: func(exchange *dto.CodeExchangeDto) { exchange.CodeVerifier = "a-guessed-verifier" }},
		{name: "no verifier", age: 10 * time.Second, expectsUseUp: true,
			change: func(exchange *dto.CodeExchangeDto) { exchange.CodeVerifier = "" }},
		{name: "a different return address", age: 10 * time.Second, expectsUseUp: true,
			change: func(exchange *dto.CodeExchangeDto) { exchange.RedirectUri = "https://other.example/callback" }},
		{name: "a different application", age: 10 * time.Second, expectsUseUp: true,
			change: func(exchange *dto.CodeExchangeDto) { exchange.ClientID = "other-app" }},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			now := taipeiTimeWithSeconds("2026-11-25 20:01:00")
			testHarness := newHarnessAt(t, now)
			stored := aCodeIssuedAgo(now, testCase.age)
			if testCase.alreadyUsed {
				stored.ConsumedAt = at(now.Add(-5 * time.Second))
			}
			testHarness.codeRepository.EXPECT().FindByDigest(gomock.Any(), digestOf("code-secret")).Return(stored, true, nil)
			if testCase.expectsUseUp {
				testHarness.codeRepository.EXPECT().MarkConsumed(gomock.Any(), "code-1", now).Return(nil)
			}
			exchange := anExchange()
			if testCase.change != nil {
				testCase.change(&exchange)
			}

			tokenSet, exchangeError := testHarness.authorizationApplication.ExchangeCode(t.Context(), exchange)

			requireOAuthError(t, exchangeError, "invalid_grant", "授權碼無效或已過期")
			assert.Empty(t, tokenSet.AccessToken)
			assert.Zero(t, testHarness.rolledBack, "the code stays used up: the refusal must not roll the write back")
		})
	}
}

func TestAuthorizationApplicationRefusesUnknownOrMissingCodes(t *testing.T) {
	for _, code := range []string{"never-issued", ""} {
		t.Run("code "+code, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			testHarness.codeRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
				Return(entities.AuthorizationCode{}, false, nil)
			exchange := anExchange()
			exchange.Code = code

			_, exchangeError := testHarness.authorizationApplication.ExchangeCode(t.Context(), exchange)

			requireOAuthError(t, exchangeError, "invalid_grant", "授權碼無效或已過期")
		})
	}
}

func TestAuthorizationApplicationPassesOnFailuresWhileExchangingACode(t *testing.T) {
	now := taipeiTimeWithSeconds("2026-11-25 20:00:20")
	cases := []struct {
		name    string
		arrange func(testHarness *harness)
	}{
		{name: "finding the code", arrange: func(testHarness *harness) {
			testHarness.codeRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
				Return(entities.AuthorizationCode{}, false, errStorage)
		}},
		{name: "using the code up", arrange: func(testHarness *harness) {
			testHarness.codeRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
				Return(aCodeIssuedAgo(now, time.Second), true, nil)
			testHarness.codeRepository.EXPECT().MarkConsumed(gomock.Any(), gomock.Any(), gomock.Any()).Return(errStorage)
		}},
		{name: "signing the access token", arrange: func(testHarness *harness) {
			testHarness.codeRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
				Return(aCodeIssuedAgo(now, time.Second), true, nil)
			testHarness.codeRepository.EXPECT().MarkConsumed(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			testHarness.accessTokenProxy.EXPECT().Issue(gomock.Any()).Return(vo.AccessTokenVo{}, errStorage)
		}},
		{name: "storing the refresh token", arrange: func(testHarness *harness) {
			testHarness.codeRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
				Return(aCodeIssuedAgo(now, time.Second), true, nil)
			testHarness.codeRepository.EXPECT().MarkConsumed(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			testHarness.accessTokenProxy.EXPECT().Issue(gomock.Any()).Return(vo.AccessTokenVo{AccessToken: "signed"}, nil)
			testHarness.randomTokenProxy.EXPECT().NewToken().Return("refresh-secret")
			testHarness.refreshRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errStorage)
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarnessAt(t, now)
			testCase.arrange(testHarness)

			_, exchangeError := testHarness.authorizationApplication.ExchangeCode(t.Context(), anExchange())

			assert.ErrorIs(t, exchangeError, errStorage)
		})
	}
}

// aRefreshTokenIssuedAgo is a stored refresh token of family "family-1", issued age ago, living 30 days.
func aRefreshTokenIssuedAgo(now time.Time, age time.Duration) entities.RefreshToken {
	return entities.RefreshToken{ID: "refresh-1", FamilyID: "family-1", TokenDigest: digestOf("presented-refresh"),
		IssuedAt: now.Add(-age), ExpiresAt: now.Add(-age).Add(thirtyDays)}
}

func TestAuthorizationApplicationRotatesACurrentRefreshToken(t *testing.T) {
	cases := []struct {
		name string
		age  time.Duration
	}{
		{name: "a refresh token issued a minute ago", age: time.Minute},
		{name: "a refresh token issued 29 days 23 hours ago", age: thirtyDays - time.Hour},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			now := taipeiTimeWithSeconds("2026-11-25 20:00:00")
			testHarness := newHarnessAt(t, now)
			testHarness.refreshRepository.EXPECT().FindByDigest(gomock.Any(), digestOf("presented-refresh")).
				Return(aRefreshTokenIssuedAgo(now, testCase.age), true, nil)
			testHarness.refreshRepository.EXPECT().MarkRotated(gomock.Any(), "refresh-1", now).Return(nil)
			testHarness.issuesTokens(now, "family-1")

			tokenSet, refreshError := testHarness.authorizationApplication.Refresh(t.Context(), "presented-refresh")

			require.NoError(t, refreshError)
			assert.Equal(t, "signed-access-token", tokenSet.AccessToken)
			assert.Equal(t, "refresh-secret", tokenSet.RefreshToken)
			assert.Equal(t, now.Add(thirtyDays), tokenSet.RefreshTokenExpiresAt)
		})
	}
}

func TestAuthorizationApplicationEndsTheWholeSignInWhenAReplacedRefreshTokenIsReused(t *testing.T) {
	now := taipeiTimeWithSeconds("2026-11-25 20:00:00")
	testHarness := newHarnessAt(t, now)
	replaced := aRefreshTokenIssuedAgo(now, time.Hour)
	replaced.RotatedAt = at(now.Add(-30 * time.Minute))
	testHarness.refreshRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).Return(replaced, true, nil)
	testHarness.refreshRepository.EXPECT().RevokeFamily(gomock.Any(), "family-1", now).Return(nil)

	_, refreshError := testHarness.authorizationApplication.Refresh(t.Context(), "presented-refresh")

	requireOAuthError(t, refreshError, "invalid_grant", "登入已失效，請重新登入")
	assert.Zero(t, testHarness.rolledBack, "the family stays revoked: the refusal must not roll the write back")
}

func TestAuthorizationApplicationRefusesRefreshTokensThatNoLongerAdmit(t *testing.T) {
	now := taipeiTimeWithSeconds("2026-11-25 20:00:00")
	thirtyDaysOld := aRefreshTokenIssuedAgo(now, thirtyDays)
	revoked := aRefreshTokenIssuedAgo(now, time.Hour)
	revoked.RevokedAt = at(now.Add(-time.Minute))

	cases := []struct {
		name    string
		stored  entities.RefreshToken
		isFound bool
		token   string
		looksUp bool
	}{
		{name: "a refresh token 30 days old", stored: thirtyDaysOld, isFound: true, token: "presented-refresh", looksUp: true},
		{name: "a revoked refresh token", stored: revoked, isFound: true, token: "presented-refresh", looksUp: true},
		{name: "an unknown refresh token", token: "presented-refresh", looksUp: true},
		{name: "no refresh token", token: ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarnessAt(t, now)
			if testCase.looksUp {
				testHarness.refreshRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
					Return(testCase.stored, testCase.isFound, nil)
			}

			tokenSet, refreshError := testHarness.authorizationApplication.Refresh(t.Context(), testCase.token)

			requireOAuthError(t, refreshError, "invalid_grant", "登入已失效，請重新登入")
			assert.Empty(t, tokenSet.AccessToken)
		})
	}
}

func TestAuthorizationApplicationPassesOnFailuresWhileRefreshing(t *testing.T) {
	now := taipeiTimeWithSeconds("2026-11-25 20:00:00")
	replaced := aRefreshTokenIssuedAgo(now, time.Hour)
	replaced.RotatedAt = at(now.Add(-time.Minute))
	cases := []struct {
		name    string
		arrange func(testHarness *harness)
	}{
		{name: "finding the refresh token", arrange: func(testHarness *harness) {
			testHarness.refreshRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
				Return(entities.RefreshToken{}, false, errStorage)
		}},
		{name: "revoking the family", arrange: func(testHarness *harness) {
			testHarness.refreshRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).Return(replaced, true, nil)
			testHarness.refreshRepository.EXPECT().RevokeFamily(gomock.Any(), gomock.Any(), gomock.Any()).Return(errStorage)
		}},
		{name: "marking it rotated", arrange: func(testHarness *harness) {
			testHarness.refreshRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
				Return(aRefreshTokenIssuedAgo(now, time.Hour), true, nil)
			testHarness.refreshRepository.EXPECT().MarkRotated(gomock.Any(), gomock.Any(), gomock.Any()).Return(errStorage)
		}},
		{name: "issuing the next tokens", arrange: func(testHarness *harness) {
			testHarness.refreshRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
				Return(aRefreshTokenIssuedAgo(now, time.Hour), true, nil)
			testHarness.refreshRepository.EXPECT().MarkRotated(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			testHarness.accessTokenProxy.EXPECT().Issue(gomock.Any()).Return(vo.AccessTokenVo{}, errStorage)
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarnessAt(t, now)
			testCase.arrange(testHarness)

			_, refreshError := testHarness.authorizationApplication.Refresh(t.Context(), "presented-refresh")

			assert.ErrorIs(t, refreshError, errStorage)
		})
	}
}

func TestAuthorizationApplicationSigningOutEndsTheSignIn(t *testing.T) {
	now := taipeiTimeWithSeconds("2026-11-25 20:00:00")
	testHarness := newHarnessAt(t, now)
	testHarness.refreshRepository.EXPECT().FindByDigest(gomock.Any(), digestOf("presented-refresh")).
		Return(aRefreshTokenIssuedAgo(now, time.Hour), true, nil)
	testHarness.refreshRepository.EXPECT().RevokeFamily(gomock.Any(), "family-1", now).Return(nil)

	revokeError := testHarness.authorizationApplication.Revoke(t.Context(), "presented-refresh")

	require.NoError(t, revokeError)
}

func TestAuthorizationApplicationSigningOutWithNothingToEndIsDone(t *testing.T) {
	t.Run("an unknown refresh token", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")
		testHarness.refreshRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).Return(entities.RefreshToken{}, false, nil)

		assert.NoError(t, testHarness.authorizationApplication.Revoke(t.Context(), "unknown"))
	})

	t.Run("no refresh token", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")

		assert.NoError(t, testHarness.authorizationApplication.Revoke(t.Context(), ""))
	})

	t.Run("a storage failure is passed on", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")
		testHarness.refreshRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
			Return(entities.RefreshToken{}, false, errStorage)

		assert.ErrorIs(t, testHarness.authorizationApplication.Revoke(t.Context(), "presented-refresh"), errStorage)
	})
}
