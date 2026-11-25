package application_test

import (
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

const memberKey = "tbk_member-key-for-tests"

func digestOf(secret string) string {
	return domains.NewOpaqueSecretDomain(secret).Digest()
}

func anActiveKey(memberName string, lastUsedAt *time.Time) entities.ApiKey {
	return entities.ApiKey{ID: "key-1", MemberName: memberName, KeyDigest: digestOf(memberKey), LastUsedAt: lastUsedAt}
}

func at(instant time.Time) *time.Time {
	return &instant
}

func TestAuthenticationApplicationAdmitsAnActiveMemberKeyAsItsMember(t *testing.T) {
	testHarness := newHarnessAt(t, taipeiTimeWithSeconds("2026-11-25 20:00:50"))
	testHarness.apiKeyRepository.EXPECT().FindByDigest(gomock.Any(), digestOf(memberKey)).
		Return(anActiveKey("desk-chief", at(taipeiTimeWithSeconds("2026-11-25 20:00:10"))), true, nil)

	caller, identifyError := testHarness.authenticationApplication.IdentifyCaller(t.Context(),
		dto.CallerCredentialDto{ApiKey: memberKey})

	require.NoError(t, identifyError)
	assert.Equal(t, dto.CallerDto{Role: "agent", Name: "desk-chief"}, caller)
}

func TestAuthenticationApplicationRefreshesLastUsedAtMostOnceAMinute(t *testing.T) {
	cases := []struct {
		name         string
		lastUsedAt   *time.Time
		now          string
		expectsTouch bool
	}{
		{name: "used 40 seconds ago", lastUsedAt: at(taipeiTimeWithSeconds("2026-11-25 20:00:10")), now: "2026-11-25 20:00:50"},
		{name: "used 65 seconds ago", lastUsedAt: at(taipeiTimeWithSeconds("2026-11-25 20:00:10")), now: "2026-11-25 20:01:15", expectsTouch: true},
		{name: "never used", now: "2026-11-25 20:01:15", expectsTouch: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			now := taipeiTimeWithSeconds(testCase.now)
			testHarness := newHarnessAt(t, now)
			testHarness.apiKeyRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
				Return(anActiveKey("desk-chief", testCase.lastUsedAt), true, nil)
			if testCase.expectsTouch {
				testHarness.apiKeyRepository.EXPECT().TouchLastUsed(gomock.Any(), "key-1", now).Return(nil)
			}

			_, identifyError := testHarness.authenticationApplication.IdentifyCaller(t.Context(),
				dto.CallerCredentialDto{ApiKey: memberKey})

			require.NoError(t, identifyError)
		})
	}
}

func TestAuthenticationApplicationIdentifiesTheOwnerByAccessToken(t *testing.T) {
	cases := []struct {
		name       string
		credential dto.CallerCredentialDto
	}{
		{name: "the owner's valid access token", credential: dto.CallerCredentialDto{AccessToken: "good"}},
		{name: "an unknown key but a valid access token", credential: dto.CallerCredentialDto{ApiKey: "tbk_unknown", AccessToken: "good"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			testHarness.apiKeyRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
				Return(entities.ApiKey{}, false, nil).AnyTimes()
			testHarness.accessTokenProxy.EXPECT().Verify("good").Return(vo.AccessTokenClaimsVo{Subject: "owner"}, nil)

			caller, identifyError := testHarness.authenticationApplication.IdentifyCaller(t.Context(), testCase.credential)

			require.NoError(t, identifyError)
			assert.Equal(t, dto.CallerDto{Role: "owner", Name: "owner"}, caller)
		})
	}
}

func TestAuthenticationApplicationRefusesUnknownCallers(t *testing.T) {
	revokedKey := anActiveKey("stock-scanner", nil)
	revokedKey.RevokedAt = at(taipeiTime("2026-11-25 19:00"))

	cases := []struct {
		name        string
		credential  dto.CallerCredentialDto
		storedKey   *entities.ApiKey
		verifyClaim vo.AccessTokenClaimsVo
		verifyError error
	}{
		{name: "nothing presented", credential: dto.CallerCredentialDto{}},
		{name: "an unknown key", credential: dto.CallerCredentialDto{ApiKey: "tbk_unknown"}},
		{name: "a revoked key", credential: dto.CallerCredentialDto{ApiKey: memberKey}, storedKey: &revokedKey},
		{name: "an expired or tampered access token", credential: dto.CallerCredentialDto{AccessToken: "bad"},
			verifyError: domains.AuthenticationRequired("expired")},
		{name: "an access token that is not the owner's", credential: dto.CallerCredentialDto{AccessToken: "bad"},
			verifyClaim: vo.AccessTokenClaimsVo{Subject: "someone"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			if testCase.credential.ApiKey != "" {
				found := testCase.storedKey != nil
				stored := entities.ApiKey{}
				if found {
					stored = *testCase.storedKey
				}
				testHarness.apiKeyRepository.EXPECT().FindByDigest(gomock.Any(), digestOf(testCase.credential.ApiKey)).
					Return(stored, found, nil)
			}
			if testCase.credential.AccessToken != "" {
				testHarness.accessTokenProxy.EXPECT().Verify("bad").Return(testCase.verifyClaim, testCase.verifyError)
			}

			_, identifyError := testHarness.authenticationApplication.IdentifyCaller(t.Context(), testCase.credential)

			assert.ErrorIs(t, identifyError, domains.ErrAuthenticationRequired)
		})
	}
}

func TestAuthenticationApplicationPassesOnStorageFailuresWhileIdentifying(t *testing.T) {
	t.Run("finding the key", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")
		testHarness.apiKeyRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).Return(entities.ApiKey{}, false, errStorage)

		_, identifyError := testHarness.authenticationApplication.IdentifyCaller(t.Context(),
			dto.CallerCredentialDto{ApiKey: memberKey})

		assert.ErrorIs(t, identifyError, errStorage)
	})

	t.Run("touching last used", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")
		testHarness.apiKeyRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).Return(anActiveKey("desk-chief", nil), true, nil)
		testHarness.apiKeyRepository.EXPECT().TouchLastUsed(gomock.Any(), gomock.Any(), gomock.Any()).Return(errStorage)

		_, identifyError := testHarness.authenticationApplication.IdentifyCaller(t.Context(),
			dto.CallerCredentialDto{ApiKey: memberKey})

		assert.ErrorIs(t, identifyError, errStorage)
	})
}

func TestAuthenticationApplicationTellsWhoIsSignedIn(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")

	assert.Equal(t, dto.OwnerDto{Email: "owner@example.com"}, testHarness.authenticationApplication.CurrentOwner())
}
