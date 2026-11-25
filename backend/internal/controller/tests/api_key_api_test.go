package controller_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestOwnerOnlyRoutesRefuseMemberKeysAndStrangers(t *testing.T) {
	routes := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/me"},
		{method: http.MethodGet, path: "/api/v1/api-keys"},
		{method: http.MethodPost, path: "/api/v1/api-keys"},
		{method: http.MethodPost, path: "/api/v1/api-keys/key-1/revoke"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			testApi := newApi(t)

			asMember := testApi.call(t, route.method, route.path, `{"memberName":"x"}`, asAgent())
			anonymous := testApi.call(t, route.method, route.path, `{"memberName":"x"}`, nil)

			assert.Equal(t, http.StatusForbidden, asMember.status)
			assert.Equal(t, "forbidden", errorCodeOf(asMember))
			assert.Equal(t, http.StatusUnauthorized, anonymous.status)
			assert.Equal(t, "unauthenticated", errorCodeOf(anonymous))
		})
	}
}

func TestMeAnswersTheOwnersEmail(t *testing.T) {
	result := newApi(t).call(t, http.MethodGet, "/api/v1/me", "", asOwner())

	require.Equal(t, http.StatusOK, result.status)
	assert.Equal(t, map[string]any{"email": "owner@example.com"}, result.body)
}

func TestCreatingAnApiKeyAnswersTheFullKeyOnce(t *testing.T) {
	testApi := newApi(t)
	testApi.randomTokenProxy.EXPECT().NewToken().Return("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG")
	testApi.apiKeyRepository.EXPECT().HasActiveMemberName(gomock.Any(), "stock-scanner").Return(false, nil)
	testApi.apiKeyRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, apiKey entities.ApiKey) (entities.ApiKey, error) {
			apiKey.ID = "key-new"
			return apiKey, nil
		})

	result := testApi.call(t, http.MethodPost, "/api/v1/api-keys", `{"memberName":"stock-scanner","note":"daily scan"}`,
		asOwner())

	require.Equal(t, http.StatusCreated, result.status)
	assert.Equal(t, map[string]any{"id": "key-new", "memberName": "stock-scanner", "note": "daily scan",
		"displayPrefix": "tbk_abcdefgh", "status": "active", "createdAt": "2026-11-25T20:00:00+08:00",
		"lastUsedAt": nil, "revokedAt": nil, "key": "tbk_abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG"}, result.body)
	assert.Equal(t, "no-store", result.header.Get("Cache-Control"))
}

func TestCreatingAnApiKeyMapsRefusalsToStatuses(t *testing.T) {
	cases := []struct {
		name           string
		arrange        func(testApi *api)
		body           string
		expectedStatus int
		expectedCode   string
	}{
		{name: "a member name in use", arrange: func(testApi *api) {
			testApi.randomTokenProxy.EXPECT().NewToken().Return("random")
			testApi.apiKeyRepository.EXPECT().HasActiveMemberName(gomock.Any(), "stock-scanner").Return(true, nil)
		}, body: `{"memberName":"Stock-Scanner"}`, expectedStatus: http.StatusConflict, expectedCode: "memberNameInUse"},
		{name: "a blank member name", arrange: func(testApi *api) {
			testApi.randomTokenProxy.EXPECT().NewToken().Return("random")
		}, body: `{"memberName":"  "}`, expectedStatus: http.StatusBadRequest, expectedCode: "invalidApiKey"},
		{name: "an unreadable body", arrange: func(*api) {}, body: `{`,
			expectedStatus: http.StatusBadRequest, expectedCode: "invalidRequest"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testApi := newApi(t)
			testCase.arrange(testApi)

			result := testApi.call(t, http.MethodPost, "/api/v1/api-keys", testCase.body, asOwner())

			assert.Equal(t, testCase.expectedStatus, result.status)
			assert.Equal(t, testCase.expectedCode, errorCodeOf(result))
		})
	}
}

func TestListingApiKeysNeverShowsAKey(t *testing.T) {
	testApi := newApi(t)
	testApi.apiKeyRepository.EXPECT().List(gomock.Any()).Return([]entities.ApiKey{
		{ID: "key-b", MemberName: "risk-budget-officer", KeyDigest: "secret-digest", DisplayPrefix: "tbk_bbbbbbbb",
			CreatedAt: now.Add(-time.Hour)},
	}, nil)

	result := testApi.call(t, http.MethodGet, "/api/v1/api-keys", "", asOwner())

	require.Equal(t, http.StatusOK, result.status)
	require.Len(t, result.list, 1)
	assert.Equal(t, "key-b", result.list[0]["id"])
	assert.Equal(t, "active", result.list[0]["status"])
	assert.NotContains(t, result.list[0], "key")
	assert.NotContains(t, result.list[0], "keyDigest")
}

func TestListingApiKeysReportsAnOutage(t *testing.T) {
	testApi := newApi(t)
	testApi.apiKeyRepository.EXPECT().List(gomock.Any()).Return(nil, errors.New("down"))

	result := testApi.call(t, http.MethodGet, "/api/v1/api-keys", "", asOwner())

	assert.Equal(t, http.StatusInternalServerError, result.status)
}

func TestRevokingAnApiKey(t *testing.T) {
	testApi := newApi(t)
	revokedAt := now
	gomock.InOrder(
		testApi.apiKeyRepository.EXPECT().FindByID(gomock.Any(), "key-1").Return(entities.ApiKey{ID: "key-1"}, true, nil),
		testApi.apiKeyRepository.EXPECT().Revoke(gomock.Any(), "key-1", now).Return(nil),
		testApi.apiKeyRepository.EXPECT().FindByID(gomock.Any(), "key-1").
			Return(entities.ApiKey{ID: "key-1", RevokedAt: &revokedAt}, true, nil),
	)
	testApi.apiKeyRepository.EXPECT().FindByID(gomock.Any(), "missing").Return(entities.ApiKey{}, false, nil)

	revoked := testApi.call(t, http.MethodPost, "/api/v1/api-keys/key-1/revoke", "", asOwner())
	missing := testApi.call(t, http.MethodPost, "/api/v1/api-keys/missing/revoke", "", asOwner())

	require.Equal(t, http.StatusOK, revoked.status)
	assert.Equal(t, "revoked", revoked.body["status"])
	assert.Equal(t, "2026-11-25T20:00:00+08:00", revoked.body["revokedAt"])
	assert.Equal(t, http.StatusNotFound, missing.status)
	assert.Equal(t, "apiKeyNotFound", errorCodeOf(missing))
}
