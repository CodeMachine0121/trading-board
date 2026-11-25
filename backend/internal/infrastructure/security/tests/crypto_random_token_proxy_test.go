package security_test

import (
	"encoding/base64"
	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptoRandomTokenProxyDrawsDistinct256BitTokens(t *testing.T) {
	tokenProxy := security.NewCryptoRandomTokenProxy()

	first := tokenProxy.NewToken()
	second := tokenProxy.NewToken()

	decoded, decodeError := base64.RawURLEncoding.DecodeString(first)
	require.NoError(t, decodeError)
	assert.Len(t, decoded, 32)
	assert.NotEqual(t, first, second)
}
