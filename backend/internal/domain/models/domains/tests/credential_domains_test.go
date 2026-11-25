package domains_test

import (
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
)

func TestOwnerAccountDomainMatchesTheEmail(t *testing.T) {
	configured := domains.NewOwnerAccountDomain(vo.OwnerAccountVo{Email: " Owner@Example.com ", PasswordHash: "hash"})
	withoutHash := domains.NewOwnerAccountDomain(vo.OwnerAccountVo{Email: "owner@example.com"})
	withoutEmail := domains.NewOwnerAccountDomain(vo.OwnerAccountVo{PasswordHash: "hash"})

	cases := []struct {
		name     string
		account  domains.OwnerAccountDomain
		email    string
		expected bool
	}{
		{name: "the same email in another case with spaces", account: configured, email: "  OWNER@example.COM", expected: true},
		{name: "another email", account: configured, email: "someone@example.com"},
		{name: "an account without a password hash", account: withoutHash, email: "owner@example.com"},
		{name: "an account without an email, typed empty", account: withoutEmail, email: ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, testCase.account.MatchesEmail(testCase.email))
		})
	}
	assert.Equal(t, dto.OwnerDto{Email: "owner@example.com"}, configured.ToOwnerDto())
}

func TestSignInThrottleDomainLocksOnFiveFailuresWithinFifteenMinutes(t *testing.T) {
	failures := func(instants ...string) []entities.SignInFailure {
		list := []entities.SignInFailure{}
		for _, instant := range instants {
			list = append(list, entities.SignInFailure{FailedAt: taipeiTime(instant)})
		}
		return list
	}

	cases := []struct {
		name     string
		failures []entities.SignInFailure
		now      string
		expected bool
	}{
		{name: "four failures never lock", failures: failures("2026-11-25 20:00", "2026-11-25 19:59",
			"2026-11-25 19:58", "2026-11-25 19:57"), now: "2026-11-25 20:01"},
		{name: "five failures exactly 15 minutes apart lock", failures: failures("2026-11-25 19:45", "2026-11-25 20:00",
			"2026-11-25 19:50", "2026-11-25 19:55", "2026-11-25 19:58"), now: "2026-11-25 20:14", expected: true},
		{name: "five failures 16 minutes apart do not lock", failures: failures("2026-11-25 19:44", "2026-11-25 20:00",
			"2026-11-25 19:50", "2026-11-25 19:55", "2026-11-25 19:58"), now: "2026-11-25 20:01"},
		{name: "the lock is over 15 minutes after the fifth failure", failures: failures("2026-11-25 20:00",
			"2026-11-25 19:59", "2026-11-25 19:58", "2026-11-25 19:57", "2026-11-25 19:56"), now: "2026-11-25 20:15"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			throttle := domains.NewSignInThrottleDomain(testCase.failures)

			assert.Equal(t, testCase.expected, throttle.IsLockedAt(taipeiTime(testCase.now)))
		})
	}
}

func TestOpaqueSecretDomainKeepsOnlyADigestAndAShortPrefix(t *testing.T) {
	secret := domains.NewOpaqueSecretDomain("tbk_abcdefghijklmnop")
	short := domains.NewOpaqueSecretDomain("tbk_abc")

	// SHA-256 of "abc" is a published test vector.
	assert.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		domains.NewOpaqueSecretDomain("abc").Digest())
	assert.Equal(t, "tbk_abcdefgh", secret.DisplayPrefix())
	assert.Equal(t, "tbk_abc", short.DisplayPrefix())
}

func TestAuthorizationRequestDomainKeepsTheReturnAddressQuery(t *testing.T) {
	request, requestError := domains.NewAuthorizationRequestDomain(dto.AuthorizationRequestDto{
		ResponseType: "code", ClientID: "web", RedirectUri: "https://board.example/auth/callback?from=menu",
		CodeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", CodeChallengeMethod: "S256", State: "s t",
	}, vo.OAuthClientVo{ClientID: "web", RedirectUris: []string{"https://board.example/auth/callback?from=menu"}})

	redirectDto := request.ToRedirectDto(domains.NewOpaqueSecretDomain("c0de"))

	assert.NoError(t, requestError)
	assert.Equal(t, "https://board.example/auth/callback?code=c0de&from=menu&state=s+t", redirectDto.RedirectTo)
}

func TestAuthorizationRequestDomainRefusesEveryClientWhenNoneIsConfigured(t *testing.T) {
	_, requestError := domains.NewAuthorizationRequestDomain(dto.AuthorizationRequestDto{ClientID: ""}, vo.OAuthClientVo{})

	assert.ErrorIs(t, requestError, domains.ErrValidation)
	assert.EqualError(t, requestError, "invalid_client: 不明的應用程式")
}

func TestAuthorizationRequestDomainRefusesAnAllowedReturnAddressThatIsNotAnAddress(t *testing.T) {
	_, requestError := domains.NewAuthorizationRequestDomain(
		dto.AuthorizationRequestDto{ClientID: "web", RedirectUri: "http://[::1"},
		vo.OAuthClientVo{ClientID: "web", RedirectUris: []string{"http://[::1"}})

	assert.EqualError(t, requestError, "invalid_request: 不允許的返回網址")
}

func TestAuthorizationCodeDomainChecksTheVerifierAgainstThePublishedPkceExample(t *testing.T) {
	// RFC 7636 appendix B: this verifier hashes to this challenge.
	code := domains.NewAuthorizationCodeDomain(entities.AuthorizationCode{
		ClientID: "web", RedirectUri: "https://board.example/cb",
		CodeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", ExpiresAt: taipeiTime("2026-11-25 20:01"),
	})
	exchange := dto.CodeExchangeDto{CodeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
		ClientID: "web", RedirectUri: "https://board.example/cb"}

	assert.NoError(t, code.Redeem(exchange, taipeiTime("2026-11-25 20:00")))
	assert.False(t, code.IsConsumed())
}

func TestRefreshTokenDomainContinuesItsFamily(t *testing.T) {
	issuedAt := taipeiTime("2026-11-25 20:00")
	first := domains.NewOpaqueSecretDomain("first").ToRefreshToken("", issuedAt, time.Hour)
	next := domains.NewOpaqueSecretDomain("next").ToRefreshToken(first.FamilyID, issuedAt, time.Hour)

	assert.Equal(t, domains.NewOpaqueSecretDomain("first").Digest(), first.FamilyID)
	assert.Equal(t, first.FamilyID, next.FamilyID)
	assert.Equal(t, domains.NewOpaqueSecretDomain("next").Digest(), next.TokenDigest)
	assert.Equal(t, first.FamilyID, domains.NewRefreshTokenDomain(next).FamilyID())
}

func TestBusinessCalendarDomainShowsAnOptionalInstantInTaipei(t *testing.T) {
	instant := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)

	assert.Nil(t, calendar.InLocation(nil))
	assert.Equal(t, "2026-11-25T20:00:00+08:00", calendar.InLocation(&instant).Format(time.RFC3339))
}
