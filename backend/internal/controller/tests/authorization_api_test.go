package controller_test

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/controller"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// The RFC 7636 appendix B pair: this verifier hashes to this challenge.
const (
	pkceVerifier  = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	pkceChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
)

func authorizeBody(email string, password string, clientID string) string {
	return `{"email":"` + email + `","password":"` + password + `","response_type":"code","client_id":"` + clientID +
		`","redirect_uri":"` + redirectUri + `","code_challenge":"` + pkceChallenge +
		`","code_challenge_method":"S256","state":"st4te"}`
}

func oauthErrorOf(result response) (string, string) {
	code, _ := result.body["error"].(string)
	description, _ := result.body["error_description"].(string)

	return code, description
}

func refreshCookieOf(result response) *http.Cookie {
	for _, cookie := range result.cookies {
		if cookie.Name == controller.RefreshTokenCookieName {
			return cookie
		}
	}

	return nil
}

func fromTrustedOrigin() map[string]string {
	return map[string]string{"Origin": trustedOrigin}
}

func TestAuthorizeAnswersTheRedirectWithCodeAndState(t *testing.T) {
	testApi := newApi(t)
	testApi.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, nil)
	testApi.passwordHashProxy.EXPECT().Matches("fake-hash", "fake-password").Return(true)
	testApi.failureRepository.EXPECT().Clear(gomock.Any()).Return(nil)
	testApi.randomTokenProxy.EXPECT().NewToken().Return("one-time-code")
	testApi.codeRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	result := testApi.call(t, http.MethodPost, "/api/v1/oauth/authorize",
		authorizeBody("owner@example.com", "fake-password", clientID), nil)

	require.Equal(t, http.StatusOK, result.status)
	assert.Equal(t, redirectUri+"?code=one-time-code&state=st4te", result.body["redirectTo"])
	assert.Equal(t, "no-store", result.header.Get("Cache-Control"))
}

func TestAuthorizeAnswersRefusalsInOAuthShape(t *testing.T) {
	cases := []struct {
		name                string
		arrange             func(testApi *api)
		body                string
		expectedStatus      int
		expectedError       string
		expectedDescription string
	}{
		{name: "wrong credentials", arrange: func(testApi *api) {
			testApi.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, nil)
			testApi.passwordHashProxy.EXPECT().Matches(gomock.Any(), gomock.Any()).Return(false)
			testApi.failureRepository.EXPECT().Record(gomock.Any(), now).Return(nil)
		}, body: authorizeBody("owner@example.com", "wrong", clientID), expectedStatus: http.StatusUnauthorized,
			expectedError: "access_denied", expectedDescription: "電子郵件或密碼錯誤"},
		{name: "locked after five failures", arrange: func(testApi *api) {
			failures := []entities.SignInFailure{}
			for minute := range 5 {
				failures = append(failures, entities.SignInFailure{FailedAt: now.Add(-time.Duration(minute) * time.Minute)})
			}
			testApi.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(failures, nil)
		}, body: authorizeBody("owner@example.com", "fake-password", clientID), expectedStatus: http.StatusTooManyRequests,
			expectedError: "temporarily_unavailable", expectedDescription: "嘗試次數過多，請 15 分鐘後再試"},
		{name: "an unknown application", arrange: func(*api) {}, body: authorizeBody("owner@example.com", "x", "evil"),
			expectedStatus: http.StatusBadRequest, expectedError: "invalid_client", expectedDescription: "不明的應用程式"},
		{name: "an unreadable body", arrange: func(*api) {}, body: `{`,
			expectedStatus: http.StatusBadRequest, expectedError: "invalid_request", expectedDescription: "無法讀取送出的內容"},
		{name: "a storage outage", arrange: func(testApi *api) {
			testApi.failureRepository.EXPECT().ListLatest(gomock.Any(), 5).Return(nil, errors.New("down"))
		}, body: authorizeBody("owner@example.com", "x", clientID), expectedStatus: http.StatusInternalServerError,
			expectedError: "server_error", expectedDescription: "伺服器發生錯誤，請稍後再試"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testApi := newApi(t)
			testCase.arrange(testApi)

			result := testApi.call(t, http.MethodPost, "/api/v1/oauth/authorize", testCase.body, nil)

			assert.Equal(t, testCase.expectedStatus, result.status)
			errorCode, description := oauthErrorOf(result)
			assert.Equal(t, testCase.expectedError, errorCode)
			assert.Equal(t, testCase.expectedDescription, description)
			assert.Nil(t, result.body["redirectTo"])
		})
	}
}

// issuesTokens expects one access token and one refresh token "next-refresh" to be handed out.
func (testApi *api) issuesTokens() {
	testApi.accessTokenProxy.EXPECT().Issue(vo.AccessTokenClaimsVo{Subject: "owner", ExpiresAt: now.Add(15 * time.Minute)}).
		Return(vo.AccessTokenVo{AccessToken: "signed-access-token"}, nil)
	testApi.randomTokenProxy.EXPECT().NewToken().Return("next-refresh")
	testApi.refreshRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
}

func TestTokenTradesACodeAndKeepsTheRefreshTokenInAnHttpOnlyCookie(t *testing.T) {
	testApi := newApi(t)
	testApi.codeRepository.EXPECT().FindByDigest(gomock.Any(), domains.NewOpaqueSecretDomain("one-time-code").Digest()).
		Return(entities.AuthorizationCode{ID: "code-1", ClientID: clientID, RedirectUri: redirectUri,
			CodeChallenge: pkceChallenge, ExpiresAt: now.Add(40 * time.Second)}, true, nil)
	testApi.codeRepository.EXPECT().MarkConsumed(gomock.Any(), "code-1", now).Return(nil)
	testApi.issuesTokens()

	result := testApi.postForm(t, "/api/v1/oauth/token", url.Values{"grant_type": {"authorization_code"},
		"code": {"one-time-code"}, "code_verifier": {pkceVerifier}, "client_id": {clientID},
		"redirect_uri": {redirectUri}}, nil)

	require.Equal(t, http.StatusOK, result.status)
	assert.Equal(t, map[string]any{"access_token": "signed-access-token", "token_type": "Bearer",
		"expires_in": float64(900)}, result.body)
	assert.Equal(t, "no-store", result.header.Get("Cache-Control"))
	cookie := refreshCookieOf(result)
	require.NotNil(t, cookie)
	assert.Equal(t, "next-refresh", cookie.Value)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.Equal(t, "/api/v1/oauth", cookie.Path)
	assert.True(t, now.Add(30*24*time.Hour).Equal(cookie.Expires))
}

func TestTokenRefusesABadCodeAndAnUnknownGrant(t *testing.T) {
	testApi := newApi(t)
	testApi.codeRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).Return(entities.AuthorizationCode{}, false, nil)

	badCode := testApi.postForm(t, "/api/v1/oauth/token", url.Values{"grant_type": {"authorization_code"},
		"code": {"guess"}, "code_verifier": {pkceVerifier}, "client_id": {clientID}, "redirect_uri": {redirectUri}}, nil)
	unknownGrant := testApi.postForm(t, "/api/v1/oauth/token", url.Values{"grant_type": {"password"}}, nil)

	assert.Equal(t, http.StatusBadRequest, badCode.status)
	badCodeError, badCodeDescription := oauthErrorOf(badCode)
	assert.Equal(t, "invalid_grant", badCodeError)
	assert.Equal(t, "授權碼無效或已過期", badCodeDescription)
	assert.Nil(t, refreshCookieOf(badCode))
	assert.Equal(t, http.StatusBadRequest, unknownGrant.status)
	unknownGrantError, _ := oauthErrorOf(unknownGrant)
	assert.Equal(t, "unsupported_grant_type", unknownGrantError)
}

func TestTokenRefusesAnUnreadableForm(t *testing.T) {
	testApi := newApi(t)

	result := testApi.call(t, http.MethodPost, "/api/v1/oauth/token", `{`, nil)

	assert.Equal(t, http.StatusBadRequest, result.status)
	errorCode, _ := oauthErrorOf(result)
	assert.Equal(t, "invalid_request", errorCode)
}

func TestTokenRotatesTheRefreshCookieForTheWebApp(t *testing.T) {
	testApi := newApi(t)
	testApi.refreshRepository.EXPECT().FindByDigest(gomock.Any(), domains.NewOpaqueSecretDomain("current-refresh").Digest()).
		Return(entities.RefreshToken{ID: "refresh-1", FamilyID: "family-1", ExpiresAt: now.Add(time.Hour)}, true, nil)
	testApi.refreshRepository.EXPECT().MarkRotated(gomock.Any(), "refresh-1", now).Return(nil)
	testApi.issuesTokens()

	result := testApi.postForm(t, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}}, fromTrustedOrigin(),
		&http.Cookie{Name: controller.RefreshTokenCookieName, Value: "current-refresh"})

	require.Equal(t, http.StatusOK, result.status)
	assert.Equal(t, "signed-access-token", result.body["access_token"])
	assert.Equal(t, "next-refresh", refreshCookieOf(result).Value)
}

func TestTokenRefusesARefreshThatNoLongerWorksAndForgetsTheCookie(t *testing.T) {
	testApi := newApi(t)
	testApi.refreshRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).Return(entities.RefreshToken{}, false, nil)

	result := testApi.postForm(t, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}}, fromTrustedOrigin(),
		&http.Cookie{Name: controller.RefreshTokenCookieName, Value: "unknown-refresh"})

	assert.Equal(t, http.StatusBadRequest, result.status)
	errorCode, description := oauthErrorOf(result)
	assert.Equal(t, "invalid_grant", errorCode)
	assert.Equal(t, "登入已失效，請重新登入", description)
	cookie := refreshCookieOf(result)
	require.NotNil(t, cookie)
	assert.Empty(t, cookie.Value)
	assert.Less(t, cookie.MaxAge, 0)
}

func TestCookieRequestsFromAnotherWebsiteAreRefused(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		form    url.Values
		headers map[string]string
	}{
		{name: "renewal from another origin", path: "/api/v1/oauth/token", form: url.Values{"grant_type": {"refresh_token"}},
			headers: map[string]string{"Origin": "https://evil.example"}},
		{name: "renewal without an origin", path: "/api/v1/oauth/token", form: url.Values{"grant_type": {"refresh_token"}}},
		{name: "sign-out from another origin", path: "/api/v1/oauth/revoke", form: url.Values{},
			headers: map[string]string{"Origin": "https://evil.example"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			// No repository expectation: the refresh token is not even looked at, so it keeps working.
			testApi := newApi(t)

			result := testApi.postForm(t, testCase.path, testCase.form, testCase.headers,
				&http.Cookie{Name: controller.RefreshTokenCookieName, Value: "current-refresh"})

			assert.Equal(t, http.StatusForbidden, result.status)
			errorCode, description := oauthErrorOf(result)
			assert.Equal(t, "invalid_request", errorCode)
			assert.Equal(t, "請求來源不被允許", description)
			assert.Nil(t, refreshCookieOf(result))
		})
	}
}

func TestRevokeEndsTheSignInAndForgetsTheCookie(t *testing.T) {
	testApi := newApi(t)
	testApi.refreshRepository.EXPECT().FindByDigest(gomock.Any(), domains.NewOpaqueSecretDomain("current-refresh").Digest()).
		Return(entities.RefreshToken{ID: "refresh-1", FamilyID: "family-1"}, true, nil)
	testApi.refreshRepository.EXPECT().RevokeFamily(gomock.Any(), "family-1", now).Return(nil)

	result := testApi.postForm(t, "/api/v1/oauth/revoke", url.Values{}, fromTrustedOrigin(),
		&http.Cookie{Name: controller.RefreshTokenCookieName, Value: "current-refresh"})

	assert.Equal(t, http.StatusOK, result.status)
	assert.Less(t, refreshCookieOf(result).MaxAge, 0)
}

func TestRevokeWithoutACookieIsDoneAndAnOutageIsAServerError(t *testing.T) {
	testApi := newApi(t)
	testApi.refreshRepository.EXPECT().FindByDigest(gomock.Any(), gomock.Any()).
		Return(entities.RefreshToken{}, false, errors.New("down"))

	withoutCookie := testApi.postForm(t, "/api/v1/oauth/revoke", url.Values{}, fromTrustedOrigin())
	outage := testApi.postForm(t, "/api/v1/oauth/revoke", url.Values{}, fromTrustedOrigin(),
		&http.Cookie{Name: controller.RefreshTokenCookieName, Value: "current-refresh"})

	assert.Equal(t, http.StatusOK, withoutCookie.status)
	assert.Equal(t, http.StatusInternalServerError, outage.status)
	errorCode, _ := oauthErrorOf(outage)
	assert.Equal(t, "server_error", errorCode)
}
