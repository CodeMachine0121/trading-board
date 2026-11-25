package domains

import (
	"net/url"
	"regexp"
	"slices"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
)

const pkceMethodS256 = "S256"

// s256CodeChallenge is BASE64URL(SHA256(verifier)) without padding: always 43 characters of the URL-safe alphabet.
var s256CodeChallenge = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

// AuthorizationRequestDomain is an authorization request the board has accepted: known client, allowed return
// address, response type "code" and a PKCE S256 challenge. A refused request is never redirected anywhere.
type AuthorizationRequestDomain struct {
	request     dto.AuthorizationRequestDto
	redirectUrl *url.URL
}

func NewAuthorizationRequestDomain(
	request dto.AuthorizationRequestDto, client vo.OAuthClientVo,
) (AuthorizationRequestDomain, error) {
	if client.ClientID == "" || request.ClientID != client.ClientID {
		return AuthorizationRequestDomain{}, InvalidClient("不明的應用程式")
	}
	redirectUrl, parseError := url.Parse(request.RedirectUri)
	if !slices.Contains(client.RedirectUris, request.RedirectUri) || parseError != nil {
		return AuthorizationRequestDomain{}, InvalidRequest("不允許的返回網址")
	}
	if request.ResponseType != "code" {
		return AuthorizationRequestDomain{}, UnsupportedResponseType("只支援 code")
	}
	if request.CodeChallengeMethod != pkceMethodS256 || !s256CodeChallenge.MatchString(request.CodeChallenge) {
		return AuthorizationRequestDomain{}, InvalidRequest("必須使用 S256 驗證")
	}

	return AuthorizationRequestDomain{request: request, redirectUrl: redirectUrl}, nil
}

// ToAuthorizationCode binds the code to this client, return address and challenge.
func (authorizationRequest AuthorizationRequestDomain) ToAuthorizationCode(
	code OpaqueSecretDomain, issuedAt time.Time, lifetime time.Duration,
) entities.AuthorizationCode {
	return entities.AuthorizationCode{
		CodeDigest:    code.Digest(),
		ClientID:      authorizationRequest.request.ClientID,
		RedirectUri:   authorizationRequest.request.RedirectUri,
		CodeChallenge: authorizationRequest.request.CodeChallenge,
		CreatedAt:     issuedAt,
		ExpiresAt:     issuedAt.Add(lifetime),
	}
}

// ToRedirectDto appends code and state to the return address, keeping whatever query it already had.
func (authorizationRequest AuthorizationRequestDomain) ToRedirectDto(code OpaqueSecretDomain) dto.AuthorizationRedirectDto {
	redirectUrl := *authorizationRequest.redirectUrl
	query := redirectUrl.Query()
	query.Set("code", code.Value())
	if authorizationRequest.request.State != "" {
		query.Set("state", authorizationRequest.request.State)
	}
	redirectUrl.RawQuery = query.Encode()

	return dto.AuthorizationRedirectDto{RedirectTo: redirectUrl.String()}
}
