package application_test

import (
	"errors"
	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationApplicationSignsTheOwnerInFor24Hours(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")
	expiresAt := taipeiTime("2026-11-26 20:00")
	testHarness.accessTokenProxy.EXPECT().Issue(vo.AccessTokenClaimsVo{Subject: "owner", ExpiresAt: expiresAt}).
		Return(vo.AccessTokenVo{AccessToken: "signed", ExpiresAt: expiresAt}, nil)

	accessToken, signInError := testHarness.authenticationApplication.SignIn(t.Context(),
		dto.SignInDto{Username: "james", Password: "secret"})

	require.NoError(t, signInError)
	assert.Equal(t, "signed", accessToken.AccessToken)
	assert.True(t, expiresAt.Equal(accessToken.ExpiresAt))
}

func TestAuthenticationApplicationRefusesAWrongPassword(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")

	accessToken, signInError := testHarness.authenticationApplication.SignIn(t.Context(),
		dto.SignInDto{Username: "james", Password: "guess"})

	require.ErrorIs(t, signInError, domains.ErrAuthenticationRequired)
	assert.Empty(t, accessToken.AccessToken)
}

func TestAuthenticationApplicationPassesOnASigningFailure(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")
	testHarness.accessTokenProxy.EXPECT().Issue(gomockAny()).Return(vo.AccessTokenVo{}, errors.New("no key"))

	_, signInError := testHarness.authenticationApplication.SignIn(t.Context(),
		dto.SignInDto{Username: "james", Password: "secret"})

	assert.EqualError(t, signInError, "no key")
}

func TestAuthenticationApplicationIdentifiesTheCaller(t *testing.T) {
	cases := []struct {
		name         string
		credential   dto.CallerCredentialDto
		verifiesAs   *vo.AccessTokenClaimsVo
		expectedRole string
	}{
		{name: "the agent key", credential: dto.CallerCredentialDto{ApiKey: "agentkey"}, expectedRole: "agent"},
		{name: "the owner's valid pass", credential: dto.CallerCredentialDto{AccessToken: "good"},
			verifiesAs: &vo.AccessTokenClaimsVo{Subject: "owner"}, expectedRole: "owner"},
		{name: "a wrong key but a valid pass", credential: dto.CallerCredentialDto{ApiKey: "wrong", AccessToken: "good"},
			verifiesAs: &vo.AccessTokenClaimsVo{Subject: "owner"}, expectedRole: "owner"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			if testCase.verifiesAs != nil {
				testHarness.accessTokenProxy.EXPECT().Verify("good").Return(*testCase.verifiesAs, nil)
			}

			caller, identifyError := testHarness.authenticationApplication.IdentifyCaller(t.Context(), testCase.credential)

			require.NoError(t, identifyError)
			assert.Equal(t, testCase.expectedRole, caller.Role)
		})
	}
}

func TestAuthenticationApplicationRefusesUnknownCallers(t *testing.T) {
	cases := []struct {
		name        string
		credential  dto.CallerCredentialDto
		verifyClaim vo.AccessTokenClaimsVo
		verifyError error
	}{
		{name: "nothing presented", credential: dto.CallerCredentialDto{}},
		{name: "a wrong key", credential: dto.CallerCredentialDto{ApiKey: "wrong"}},
		{name: "an expired or tampered pass", credential: dto.CallerCredentialDto{AccessToken: "bad"},
			verifyError: domains.AuthenticationRequired("expired")},
		{name: "a pass that is not the owner's", credential: dto.CallerCredentialDto{AccessToken: "bad"},
			verifyClaim: vo.AccessTokenClaimsVo{Subject: "someone"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			if testCase.credential.AccessToken != "" {
				testHarness.accessTokenProxy.EXPECT().Verify("bad").Return(testCase.verifyClaim, testCase.verifyError)
			}

			_, identifyError := testHarness.authenticationApplication.IdentifyCaller(t.Context(), testCase.credential)

			assert.ErrorIs(t, identifyError, domains.ErrAuthenticationRequired)
		})
	}
}
