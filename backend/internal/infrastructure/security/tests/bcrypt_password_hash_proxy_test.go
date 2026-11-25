package security_test

import (
	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/security"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptPasswordHashProxyMatchesOnlyTheHashedPassword(t *testing.T) {
	passwordHash, hashError := bcrypt.GenerateFromPassword([]byte("fake-test-password"), bcrypt.MinCost)
	assert.NoError(t, hashError)
	hashProxy := security.NewBcryptPasswordHashProxy()

	cases := []struct {
		name         string
		passwordHash string
		password     string
		expected     bool
	}{
		{name: "the hashed password", passwordHash: string(passwordHash), password: "fake-test-password", expected: true},
		{name: "another password", passwordHash: string(passwordHash), password: "fake-test-password!"},
		{name: "no hash configured", passwordHash: "", password: "fake-test-password"},
		{name: "a malformed hash", passwordHash: "not-a-bcrypt-hash", password: "fake-test-password"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, hashProxy.Matches(testCase.passwordHash, testCase.password))
		})
	}
}
