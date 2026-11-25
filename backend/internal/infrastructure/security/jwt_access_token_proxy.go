package security

import (
	"errors"
	"fmt"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/golang-jwt/jwt/v5"
)

// accessTokenSigningMethod is a shared HMAC key because only this service ever reads its passes.
var accessTokenSigningMethod = jwt.SigningMethodHS256

var errSigningKeyMissing = errors.New("access token signing key is not configured")

// JwtAccessTokenProxy issues stateless signed access tokens, so one cannot be revoked before it expires.
type JwtAccessTokenProxy struct {
	signingKey []byte
}

func NewJwtAccessTokenProxy(signingKey string) *JwtAccessTokenProxy {
	return &JwtAccessTokenProxy{signingKey: []byte(signingKey)}
}

// Issue refuses to sign without a key rather than hand out forgeable passes.
func (tokenProxy *JwtAccessTokenProxy) Issue(claims vo.AccessTokenClaimsVo) (vo.AccessTokenVo, error) {
	if len(tokenProxy.signingKey) == 0 {
		return vo.AccessTokenVo{}, errSigningKeyMissing
	}

	signedToken, signError := jwt.NewWithClaims(accessTokenSigningMethod, jwt.RegisteredClaims{
		Subject:   claims.Subject,
		ExpiresAt: jwt.NewNumericDate(claims.ExpiresAt),
	}).SignedString(tokenProxy.signingKey)
	if signError != nil {
		return vo.AccessTokenVo{}, fmt.Errorf("sign access token: %w", signError)
	}

	return vo.AccessTokenVo{AccessToken: signedToken, ExpiresAt: jwt.NewNumericDate(claims.ExpiresAt).Time}, nil
}

// Verify pins the algorithm, so a pass cannot claim a weaker one such as none, and requires an expiry.
func (tokenProxy *JwtAccessTokenProxy) Verify(accessToken string) (vo.AccessTokenClaimsVo, error) {
	if len(tokenProxy.signingKey) == 0 {
		return vo.AccessTokenClaimsVo{}, domains.AuthenticationRequired("尚未設定憑證簽章鑰匙")
	}

	claims := jwt.RegisteredClaims{}
	_, parseError := jwt.ParseWithClaims(accessToken, &claims,
		func(*jwt.Token) (any, error) { return tokenProxy.signingKey, nil },
		jwt.WithValidMethods([]string{accessTokenSigningMethod.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if parseError != nil {
		return vo.AccessTokenClaimsVo{}, domains.AuthenticationRequired("登入憑證無效或已過期")
	}

	return vo.AccessTokenClaimsVo{Subject: claims.Subject, ExpiresAt: claims.ExpiresAt.Time}, nil
}
