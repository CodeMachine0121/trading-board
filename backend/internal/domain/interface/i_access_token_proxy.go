package _interface

import "github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"

//go:generate go tool mockgen -source=i_access_token_proxy.go -destination=mocks/mock_i_access_token_proxy.go -package=mocks

// IAccessTokenProxy issues and verifies the owner's short-lived access token.
type IAccessTokenProxy interface {
	Issue(claims vo.AccessTokenClaimsVo) (vo.AccessTokenVo, error)
	// Verify answers one refusal for every invalid, tampered or expired pass.
	Verify(accessToken string) (vo.AccessTokenClaimsVo, error)
}
