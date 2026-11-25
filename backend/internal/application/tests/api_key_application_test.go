package application_test

import (
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const keyRandomPart = "R4nd0mPartOfTheKeyThatIsLongEnough_0123456"

// storesKeys plays the database: whatever key is created comes back with identity "key-new".
func (testHarness *harness) storesKeys() *entities.ApiKey {
	stored := &entities.ApiKey{}
	testHarness.apiKeyRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, apiKey entities.ApiKey) (entities.ApiKey, error) {
			apiKey.ID = "key-new"
			*stored = apiKey
			return apiKey, nil
		})

	return stored
}

func TestApiKeyApplicationCreatesAKeyShownOnce(t *testing.T) {
	now := taipeiTimeWithSeconds("2026-11-25 20:00:00")
	testHarness := newHarnessAt(t, now)
	testHarness.randomTokenProxy.EXPECT().NewToken().Return(keyRandomPart)
	testHarness.apiKeyRepository.EXPECT().HasActiveMemberName(gomock.Any(), "stock-scanner").Return(false, nil)
	stored := testHarness.storesKeys()

	created, createError := testHarness.apiKeyApplication.Create(t.Context(),
		dto.ApiKeyCreateDto{MemberName: " stock-scanner ", Note: "daily scan"})

	require.NoError(t, createError)
	assert.Equal(t, "tbk_"+keyRandomPart, created.Key)
	assert.Equal(t, dto.ApiKeyDto{ID: "key-new", MemberName: "stock-scanner", Note: "daily scan",
		DisplayPrefix: "tbk_R4nd0mPa", Status: "active", CreatedAt: now}, created.ApiKeyDto)
	// Only the digest of the key is stored, never the key itself.
	assert.Equal(t, digestOf("tbk_"+keyRandomPart), stored.KeyDigest)
	assert.NotContains(t, stored.KeyDigest+stored.DisplayPrefix+stored.MemberName+stored.Note, keyRandomPart)
}

func TestApiKeyApplicationValidatesTheMemberName(t *testing.T) {
	cases := []struct {
		name            string
		memberName      string
		note            string
		expectsValid    bool
		expectedMessage string
	}{
		{name: "a 40-character member name", memberName: strings.Repeat("a", 40), expectsValid: true},
		{name: "a 40-character name in Chinese", memberName: strings.Repeat("顧", 40), expectsValid: true},
		{name: "a 41-character member name", memberName: strings.Repeat("a", 41), expectedMessage: "成員名稱須為 1 到 40 個字"},
		{name: "a member name of only spaces", memberName: "   ", expectedMessage: "成員名稱須為 1 到 40 個字"},
		{name: "a note of 200 characters", memberName: "desk-chief", note: strings.Repeat("n", 200), expectsValid: true},
		{name: "a note of 201 characters", memberName: "desk-chief", note: strings.Repeat("n", 201),
			expectedMessage: "備註最多 200 個字"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			testHarness.randomTokenProxy.EXPECT().NewToken().Return(keyRandomPart)
			if testCase.expectsValid {
				testHarness.apiKeyRepository.EXPECT().HasActiveMemberName(gomock.Any(), gomock.Any()).Return(false, nil)
				testHarness.storesKeys()
			}

			created, createError := testHarness.apiKeyApplication.Create(t.Context(),
				dto.ApiKeyCreateDto{MemberName: testCase.memberName, Note: testCase.note})

			if testCase.expectsValid {
				require.NoError(t, createError)
				assert.Equal(t, testCase.memberName, created.MemberName)
				assert.Equal(t, "active", created.Status)
				return
			}
			assert.ErrorIs(t, createError, domains.ErrValidation)
			assert.EqualError(t, createError, testCase.expectedMessage)
		})
	}
}

func TestApiKeyApplicationRefusesAMemberNameInUseIgnoringCase(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")
	testHarness.randomTokenProxy.EXPECT().NewToken().Return(keyRandomPart)
	testHarness.apiKeyRepository.EXPECT().HasActiveMemberName(gomock.Any(), "stock-scanner").Return(true, nil)

	_, createError := testHarness.apiKeyApplication.Create(t.Context(), dto.ApiKeyCreateDto{MemberName: "Stock-Scanner"})

	assert.ErrorIs(t, createError, domains.ErrConflict)
	assert.EqualError(t, createError, "這個成員名稱已有啟用中的金鑰")
}

func TestApiKeyApplicationPassesOnFailuresWhileCreating(t *testing.T) {
	cases := []struct {
		name    string
		arrange func(testHarness *harness)
	}{
		{name: "checking the member name", arrange: func(testHarness *harness) {
			testHarness.randomTokenProxy.EXPECT().NewToken().Return(keyRandomPart)
			testHarness.apiKeyRepository.EXPECT().HasActiveMemberName(gomock.Any(), gomock.Any()).Return(false, errStorage)
		}},
		{name: "storing the key", arrange: func(testHarness *harness) {
			testHarness.randomTokenProxy.EXPECT().NewToken().Return(keyRandomPart)
			testHarness.apiKeyRepository.EXPECT().HasActiveMemberName(gomock.Any(), gomock.Any()).Return(false, nil)
			testHarness.apiKeyRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(entities.ApiKey{}, errStorage)
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			testCase.arrange(testHarness)

			_, createError := testHarness.apiKeyApplication.Create(t.Context(), dto.ApiKeyCreateDto{MemberName: "desk-chief"})

			assert.ErrorIs(t, createError, errStorage)
		})
	}
}

func TestApiKeyApplicationListsKeysInStoredOrderShowingTimesInTaipei(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")
	createdA := time.Date(2026, 11, 20, 1, 0, 0, 0, time.UTC)
	createdB := time.Date(2026, 11, 24, 1, 0, 0, 0, time.UTC)
	revokedB := time.Date(2026, 11, 25, 1, 0, 0, 0, time.UTC)
	usedA := time.Date(2026, 11, 25, 2, 0, 0, 0, time.UTC)
	testHarness.apiKeyRepository.EXPECT().List(gomock.Any()).Return([]entities.ApiKey{
		{ID: "key-b", MemberName: "risk-budget-officer", DisplayPrefix: "tbk_bbbbbbbb", CreatedAt: createdB, RevokedAt: &revokedB},
		{ID: "key-a", MemberName: "stock-scanner", Note: "n", DisplayPrefix: "tbk_aaaaaaaa", CreatedAt: createdA, LastUsedAt: &usedA},
	}, nil)

	apiKeys, listError := testHarness.apiKeyApplication.List(t.Context())

	require.NoError(t, listError)
	require.Len(t, apiKeys, 2)
	assert.Equal(t, "key-b", apiKeys[0].ID)
	assert.Equal(t, "revoked", apiKeys[0].Status)
	assert.Equal(t, "2026-11-25T09:00:00+08:00", apiKeys[0].RevokedAt.Format(time.RFC3339))
	assert.Nil(t, apiKeys[0].LastUsedAt)
	assert.Equal(t, "key-a", apiKeys[1].ID)
	assert.Equal(t, "active", apiKeys[1].Status)
	assert.Equal(t, "2026-11-20T09:00:00+08:00", apiKeys[1].CreatedAt.Format(time.RFC3339))
	assert.Equal(t, "2026-11-25T10:00:00+08:00", apiKeys[1].LastUsedAt.Format(time.RFC3339))
	assert.Nil(t, apiKeys[1].RevokedAt)
}

func TestApiKeyApplicationPassesOnFailuresWhileListing(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")
	testHarness.apiKeyRepository.EXPECT().List(gomock.Any()).Return(nil, errStorage)

	_, listError := testHarness.apiKeyApplication.List(t.Context())

	assert.ErrorIs(t, listError, errStorage)
}

func TestApiKeyApplicationRevokesAnActiveKey(t *testing.T) {
	now := taipeiTimeWithSeconds("2026-11-25 20:00:00")
	testHarness := newHarnessAt(t, now)
	active := entities.ApiKey{ID: "key-a", MemberName: "stock-scanner"}
	revoked := active
	revoked.RevokedAt = &now
	gomock.InOrder(
		testHarness.apiKeyRepository.EXPECT().FindByID(gomock.Any(), "key-a").Return(active, true, nil),
		testHarness.apiKeyRepository.EXPECT().Revoke(gomock.Any(), "key-a", now).Return(nil),
		testHarness.apiKeyRepository.EXPECT().FindByID(gomock.Any(), "key-a").Return(revoked, true, nil),
	)

	revokedDto, revokeError := testHarness.apiKeyApplication.Revoke(t.Context(), "key-a")

	require.NoError(t, revokeError)
	assert.Equal(t, "revoked", revokedDto.Status)
	assert.True(t, now.Equal(*revokedDto.RevokedAt))
}

func TestApiKeyApplicationLeavesARevokedKeyAsItIs(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")
	firstRevoked := taipeiTime("2026-11-24 08:00")
	testHarness.apiKeyRepository.EXPECT().FindByID(gomock.Any(), "key-a").
		Return(entities.ApiKey{ID: "key-a", RevokedAt: &firstRevoked}, true, nil)

	revokedDto, revokeError := testHarness.apiKeyApplication.Revoke(t.Context(), "key-a")

	require.NoError(t, revokeError)
	assert.Equal(t, "revoked", revokedDto.Status)
	assert.True(t, firstRevoked.Equal(*revokedDto.RevokedAt))
}

func TestApiKeyApplicationRefusesToRevokeAnUnknownKey(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")
	testHarness.apiKeyRepository.EXPECT().FindByID(gomock.Any(), "missing").Return(entities.ApiKey{}, false, nil)

	_, revokeError := testHarness.apiKeyApplication.Revoke(t.Context(), "missing")

	assert.ErrorIs(t, revokeError, domains.ErrNotFound)
	assert.EqualError(t, revokeError, "找不到這把金鑰")
}

func TestApiKeyApplicationPassesOnFailuresWhileRevoking(t *testing.T) {
	active := entities.ApiKey{ID: "key-a"}
	cases := []struct {
		name    string
		arrange func(testHarness *harness)
	}{
		{name: "finding the key", arrange: func(testHarness *harness) {
			testHarness.apiKeyRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(entities.ApiKey{}, false, errStorage)
		}},
		{name: "revoking the key", arrange: func(testHarness *harness) {
			testHarness.apiKeyRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(active, true, nil)
			testHarness.apiKeyRepository.EXPECT().Revoke(gomock.Any(), gomock.Any(), gomock.Any()).Return(errStorage)
		}},
		{name: "reading it back", arrange: func(testHarness *harness) {
			gomock.InOrder(
				testHarness.apiKeyRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(active, true, nil),
				testHarness.apiKeyRepository.EXPECT().Revoke(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
				testHarness.apiKeyRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(entities.ApiKey{}, false, errStorage),
			)
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			testCase.arrange(testHarness)

			_, revokeError := testHarness.apiKeyApplication.Revoke(t.Context(), "key-a")

			assert.ErrorIs(t, revokeError, errStorage)
		})
	}
}
