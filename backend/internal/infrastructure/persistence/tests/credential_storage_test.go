package persistence_test

import (
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const unknownIdentity = "6f1c1e0e-8f9e-4c7e-9a43-0d3b6f0b1c2d"

var credentialNow = time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)

func TestAuthorizationCodeRepositoryStoresFindsAndConsumesACode(t *testing.T) {
	repository := persistence.NewAuthorizationCodeRepository(newTestDatabase(t))

	createError := repository.Create(t.Context(), entities.AuthorizationCode{CodeDigest: "digest-1", ClientID: "web",
		RedirectUri: "http://localhost:3000/auth/callback", CodeChallenge: "challenge",
		CreatedAt: credentialNow, ExpiresAt: credentialNow.Add(time.Minute)})
	require.NoError(t, createError)
	found, isFound, findError := repository.FindByDigest(t.Context(), "digest-1")
	require.NoError(t, findError)
	require.True(t, isFound)
	assert.Nil(t, found.ConsumedAt)

	require.NoError(t, repository.MarkConsumed(t.Context(), found.ID, credentialNow))

	consumed, _, _ := repository.FindByDigest(t.Context(), "digest-1")
	require.NotNil(t, consumed.ConsumedAt)
	assert.True(t, credentialNow.Equal(*consumed.ConsumedAt))
	_, isMissingFound, missingError := repository.FindByDigest(t.Context(), "digest-unknown")
	require.NoError(t, missingError)
	assert.False(t, isMissingFound)
}

func TestRefreshTokenRepositoryRotatesAndRevokesAWholeFamily(t *testing.T) {
	repository := persistence.NewRefreshTokenRepository(newTestDatabase(t))
	for _, digest := range []string{"first", "second"} {
		require.NoError(t, repository.Create(t.Context(), entities.RefreshToken{FamilyID: "family-1", TokenDigest: digest,
			IssuedAt: credentialNow, ExpiresAt: credentialNow.Add(time.Hour)}))
	}
	require.NoError(t, repository.Create(t.Context(), entities.RefreshToken{FamilyID: "family-2", TokenDigest: "other",
		IssuedAt: credentialNow, ExpiresAt: credentialNow.Add(time.Hour)}))
	first, _, _ := repository.FindByDigest(t.Context(), "first")

	require.NoError(t, repository.MarkRotated(t.Context(), first.ID, credentialNow))
	require.NoError(t, repository.RevokeFamily(t.Context(), "family-1", credentialNow))

	rotated, _, _ := repository.FindByDigest(t.Context(), "first")
	second, _, _ := repository.FindByDigest(t.Context(), "second")
	other, _, _ := repository.FindByDigest(t.Context(), "other")
	assert.NotNil(t, rotated.RotatedAt)
	assert.NotNil(t, rotated.RevokedAt)
	assert.NotNil(t, second.RevokedAt)
	assert.Nil(t, other.RevokedAt)
	_, isFound, findError := repository.FindByDigest(t.Context(), "unknown")
	require.NoError(t, findError)
	assert.False(t, isFound)
}

func TestApiKeyRepositoryKeepsActiveMemberNamesUnique(t *testing.T) {
	repository := persistence.NewApiKeyRepository(newTestDatabase(t))
	older, createError := repository.Create(t.Context(), entities.ApiKey{MemberName: "stock-scanner",
		MemberNameKey: "stock-scanner", KeyDigest: "digest-a", DisplayPrefix: "tbk_aaaaaaaa", CreatedAt: credentialNow})
	require.NoError(t, createError)
	assert.NotEmpty(t, older.ID)

	isTaken, checkError := repository.HasActiveMemberName(t.Context(), "stock-scanner")
	require.NoError(t, checkError)
	assert.True(t, isTaken)
	_, duplicateError := repository.Create(t.Context(), entities.ApiKey{MemberName: "Stock-Scanner",
		MemberNameKey: "stock-scanner", KeyDigest: "digest-b", DisplayPrefix: "tbk_bbbbbbbb", CreatedAt: credentialNow})
	assert.Error(t, duplicateError, "the database itself refuses a second active key for one member")

	require.NoError(t, repository.Revoke(t.Context(), older.ID, credentialNow))
	isStillTaken, _ := repository.HasActiveMemberName(t.Context(), "stock-scanner")
	assert.False(t, isStillTaken)
	newer, reuseError := repository.Create(t.Context(), entities.ApiKey{MemberName: "stock-scanner",
		MemberNameKey: "stock-scanner", KeyDigest: "digest-c", DisplayPrefix: "tbk_cccccccc",
		CreatedAt: credentialNow.Add(time.Minute)})
	require.NoError(t, reuseError)

	listed, listError := repository.List(t.Context())
	require.NoError(t, listError)
	require.Len(t, listed, 2)
	assert.Equal(t, newer.ID, listed[0].ID, "newest first")
	assert.Equal(t, older.ID, listed[1].ID)
}

func TestApiKeyRepositoryRevokesOnceAndTouchesLastUsed(t *testing.T) {
	repository := persistence.NewApiKeyRepository(newTestDatabase(t))
	created, _ := repository.Create(t.Context(), entities.ApiKey{MemberName: "desk-chief", MemberNameKey: "desk-chief",
		KeyDigest: "digest-a", DisplayPrefix: "tbk_aaaaaaaa", CreatedAt: credentialNow})

	require.NoError(t, repository.TouchLastUsed(t.Context(), created.ID, credentialNow.Add(time.Minute)))
	require.NoError(t, repository.Revoke(t.Context(), created.ID, credentialNow.Add(time.Hour)))
	require.NoError(t, repository.Revoke(t.Context(), created.ID, credentialNow.Add(2*time.Hour)))

	found, isFound, findError := repository.FindByDigest(t.Context(), "digest-a")
	require.NoError(t, findError)
	require.True(t, isFound)
	assert.True(t, credentialNow.Add(time.Minute).Equal(*found.LastUsedAt))
	assert.True(t, credentialNow.Add(time.Hour).Equal(*found.RevokedAt), "the first revoke time stays")
	byID, isFoundByID, _ := repository.FindByID(t.Context(), created.ID)
	assert.True(t, isFoundByID)
	assert.Equal(t, "desk-chief", byID.MemberName)
}

func TestApiKeyRepositoryAnswersNoneForUnknownOrMalformedIdentities(t *testing.T) {
	repository := persistence.NewApiKeyRepository(newTestDatabase(t))

	for _, apiKeyID := range []string{"not-a-uuid", unknownIdentity} {
		_, isFound, findError := repository.FindByID(t.Context(), apiKeyID)

		require.NoError(t, findError)
		assert.False(t, isFound, apiKeyID)
	}
	_, isFound, findError := repository.FindByDigest(t.Context(), "unknown")
	require.NoError(t, findError)
	assert.False(t, isFound)
}

func TestSignInFailureRepositoryListsTheLatestAndClears(t *testing.T) {
	repository := persistence.NewSignInFailureRepository(newTestDatabase(t))
	for minute := range 7 {
		require.NoError(t, repository.Record(t.Context(), credentialNow.Add(time.Duration(minute)*time.Minute)))
	}

	latest, listError := repository.ListLatest(t.Context(), 5)

	require.NoError(t, listError)
	require.Len(t, latest, 5)
	assert.True(t, credentialNow.Add(6*time.Minute).Equal(latest[0].FailedAt), "newest first")
	assert.True(t, credentialNow.Add(2*time.Minute).Equal(latest[4].FailedAt))
	require.NoError(t, repository.Clear(t.Context()))
	afterClear, _ := repository.ListLatest(t.Context(), 5)
	assert.Empty(t, afterClear)
}

func TestCredentialRepositoriesReportStorageFailures(t *testing.T) {
	database := closedDatabase(t)
	codeRepository := persistence.NewAuthorizationCodeRepository(database)
	tokenRepository := persistence.NewRefreshTokenRepository(database)
	keyRepository := persistence.NewApiKeyRepository(database)
	failureRepository := persistence.NewSignInFailureRepository(database)

	_, _, codeFindError := codeRepository.FindByDigest(t.Context(), "digest")
	_, _, tokenFindError := tokenRepository.FindByDigest(t.Context(), "digest")
	_, _, keyFindError := keyRepository.FindByID(t.Context(), unknownIdentity)
	_, keyCreateError := keyRepository.Create(t.Context(), entities.ApiKey{})
	_, keyListError := keyRepository.List(t.Context())
	_, keyCheckError := keyRepository.HasActiveMemberName(t.Context(), "desk-chief")
	_, failureListError := failureRepository.ListLatest(t.Context(), 5)

	for _, storageError := range []error{
		codeRepository.Create(t.Context(), entities.AuthorizationCode{}), codeFindError,
		codeRepository.MarkConsumed(t.Context(), unknownIdentity, credentialNow),
		tokenRepository.Create(t.Context(), entities.RefreshToken{}), tokenFindError,
		tokenRepository.MarkRotated(t.Context(), unknownIdentity, credentialNow),
		tokenRepository.RevokeFamily(t.Context(), "family", credentialNow),
		keyFindError, keyCreateError, keyListError, keyCheckError,
		keyRepository.Revoke(t.Context(), unknownIdentity, credentialNow),
		keyRepository.TouchLastUsed(t.Context(), unknownIdentity, credentialNow),
		failureRepository.Record(t.Context(), credentialNow), failureListError, failureRepository.Clear(t.Context()),
	} {
		assert.Error(t, storageError)
	}
}
