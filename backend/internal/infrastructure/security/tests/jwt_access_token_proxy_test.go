package security_test

import (
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/security"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwtAccessTokenProxyIssuesAPassItCanVerify(t *testing.T) {
	tokenProxy := security.NewJwtAccessTokenProxy("signing-key")
	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)

	issued, issueError := tokenProxy.Issue(vo.AccessTokenClaimsVo{Subject: "owner", ExpiresAt: expiresAt})
	claims, verifyError := tokenProxy.Verify(issued.AccessToken)

	require.NoError(t, issueError)
	require.NoError(t, verifyError)
	assert.True(t, expiresAt.Equal(issued.ExpiresAt))
	assert.Equal(t, "owner", claims.Subject)
	assert.True(t, expiresAt.Equal(claims.ExpiresAt))
}

func TestJwtAccessTokenProxyRefusesBadPasses(t *testing.T) {
	tokenProxy := security.NewJwtAccessTokenProxy("signing-key")
	expired, _ := tokenProxy.Issue(vo.AccessTokenClaimsVo{Subject: "owner", ExpiresAt: time.Now().Add(-time.Minute)})
	foreign, _ := security.NewJwtAccessTokenProxy("other-key").Issue(vo.AccessTokenClaimsVo{Subject: "owner", ExpiresAt: time.Now().Add(time.Hour)})
	unsigned, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{Subject: "owner",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}).SignedString(jwt.UnsafeAllowNoneSignatureType)

	cases := map[string]string{
		"an expired pass":           expired.AccessToken,
		"a pass signed by another":  foreign.AccessToken,
		"an unsigned pass":          unsigned,
		"something that is no pass": "garbage",
	}

	for name, accessToken := range cases {
		t.Run(name, func(t *testing.T) {
			_, verifyError := tokenProxy.Verify(accessToken)

			assert.ErrorIs(t, verifyError, domains.ErrAuthenticationRequired)
		})
	}
}

func TestJwtAccessTokenProxyRefusesToWorkWithoutAKey(t *testing.T) {
	tokenProxy := security.NewJwtAccessTokenProxy("")

	_, issueError := tokenProxy.Issue(vo.AccessTokenClaimsVo{Subject: "owner", ExpiresAt: time.Now().Add(time.Hour)})
	_, verifyError := tokenProxy.Verify("anything")

	assert.Error(t, issueError)
	assert.ErrorIs(t, verifyError, domains.ErrAuthenticationRequired)
}
